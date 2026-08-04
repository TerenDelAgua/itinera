package handlers

import (
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/http/middleware"
	"backend/internal/models"
	"backend/internal/services/email"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ---------- Fakes ---------------------------------------------------------

// fakeAuthStore implements database.AuthStore with hand-rolled behaviour.
// Only the methods the v2 handlers exercise are non-trivial; the rest panic
// so a regression in a different handler is loud.
type fakeAuthStore struct {
	byEmail                map[string]*models.User
	byID                   map[uuid.UUID]*models.User
	passwordUpdateCalls    []uuid.UUID
	softDeleteCascadeCalls []uuid.UUID
	createErr              error
	softDeleteCascadeErr   error
}

func (f *fakeAuthStore) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeAuthStore) GetUserByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeAuthStore) CreateUser(_ context.Context, email, password, locale string) (*models.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	u := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "hashed:" + password,
		Tier:         "free",
		Locale:       locale,
	}
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeAuthStore) UpdateUserPassword(_ context.Context, userID uuid.UUID, _ string) error {
	f.passwordUpdateCalls = append(f.passwordUpdateCalls, userID)
	return nil
}

func (f *fakeAuthStore) ClaimGuestTrips(context.Context, string, uuid.UUID) (int, error) {
	panic("not used in tests")
}
func (f *fakeAuthStore) SoftDeleteUser(context.Context, uuid.UUID) (string, error) {
	panic("not used in tests")
}

func (f *fakeAuthStore) SoftDeleteUserCascade(_ context.Context, userID uuid.UUID) (string, error) {
	f.softDeleteCascadeCalls = append(f.softDeleteCascadeCalls, userID)
	if f.softDeleteCascadeErr != nil {
		return "", f.softDeleteCascadeErr
	}
	return "user@example.test", nil
}
func (f *fakeAuthStore) MarkUserAsHardDeleted(context.Context, uuid.UUID, string) error {
	panic("not used in tests")
}

var _ database.AuthStore = (*fakeAuthStore)(nil)

// fakeSessionStore implements database.SessionStore with counters so each
// test can verify how the handler exercised it.
type fakeSessionStore struct {
	createdCalls       int
	revokedByHashCalls int
	lastRevokedHash    string
	rotatedCalls       int

	// byRefreshHash is the map the Find* queries read from. Tests pre-seed
	// it with rows whose RefreshFamily / RevokedAt match the scenario.
	byRefreshHash map[string]*models.Session

	// byAccessHash mirrors byRefreshHash for the access-token lookup path.
	// Same map type but the key is the SHA-256 of the access token. Tests
	// pre-seed it to authenticate the request via the middleware.
	byAccessHash map[string]*models.Session

	// revokedFamilies accumulates UUIDs passed to RevokeFamily so a test
	// can assert that the reuse-detection path actually fires.
	revokedFamilies []uuid.UUID

	// revokedAllFor accumulates UUIDs passed to RevokeAllSessionsForUser.
	revokedAllFor []uuid.UUID
}

func (f *fakeSessionStore) CreateSession(_ context.Context, _ uuid.UUID, _ string, _ string, _ time.Time, _, _ *string) (*models.Session, error) {
	f.createdCalls++
	return nil, nil
}

func (f *fakeSessionStore) FindSessionByAccessTokenHash(_ context.Context, accessHash string) (*models.Session, error) {
	if accessHash == "" {
		return nil, pgx.ErrNoRows
	}
	if s, ok := f.byAccessHash[accessHash]; ok && s != nil {
		return s, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeSessionStore) FindSessionByRefreshTokenHash(_ context.Context, refreshHash string) (*models.Session, error) {
	if s, ok := f.byRefreshHash[refreshHash]; ok && s != nil {
		return s, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeSessionStore) RevokeSessionByAccessHash(_ context.Context, accessHash string) (int, error) {
	f.revokedByHashCalls++
	f.lastRevokedHash = accessHash
	return 1, nil
}

func (f *fakeSessionStore) RotateSession(_ context.Context, _ uuid.UUID, _ string, _ string, _ time.Time) error {
	f.rotatedCalls++
	return nil
}

func (f *fakeSessionStore) RevokeSession(context.Context, uuid.UUID) error { panic("not used") }
func (f *fakeSessionStore) RevokeFamily(_ context.Context, familyID uuid.UUID) error {
	f.revokedFamilies = append(f.revokedFamilies, familyID)
	return nil
}
func (f *fakeSessionStore) RevokeAllSessionsForUser(_ context.Context, userID uuid.UUID) error {
	f.revokedAllFor = append(f.revokedAllFor, userID)
	return nil
}
func (f *fakeSessionStore) CountActiveSessionsForUser(context.Context, uuid.UUID) (int, error) {
	panic("not used")
}
func (f *fakeSessionStore) CleanupExpiredSessions(context.Context) (int64, error) {
	panic("not used")
}

var _ database.SessionStore = (*fakeSessionStore)(nil)

// fakeResetStore implements database.PasswordResetStore with counters and
// a small in-memory map of active codes keyed by their hash. The FindActive
// methods return ErrNoRows on miss; pre-seed to exercise the happy path.
type fakeResetStore struct {
	createCalls   int
	markUsedCalls []uuid.UUID
	byHash        map[string]*models.PasswordResetToken
	byUser        map[uuid.UUID]*models.PasswordResetToken
	failAttempt   bool
}

func newFakeResetStore() *fakeResetStore {
	return &fakeResetStore{
		byHash: map[string]*models.PasswordResetToken{},
		byUser: map[uuid.UUID]*models.PasswordResetToken{},
	}
}

func (f *fakeResetStore) Create(_ context.Context, userID uuid.UUID, hash string, _ time.Time, _ *string) error {
	f.createCalls++
	t := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	}
	f.byHash[hash] = t
	f.byUser[userID] = t
	return nil
}

func (f *fakeResetStore) MarkPreviousAsUsed(context.Context, uuid.UUID) error { return nil }

func (f *fakeResetStore) FindActiveByHash(_ context.Context, hash string) (*models.PasswordResetToken, error) {
	if f.failAttempt {
		return nil, pgx.ErrNoRows
	}
	if t, ok := f.byHash[hash]; ok && t != nil {
		return t, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeResetStore) FindActiveByUser(_ context.Context, userID uuid.UUID) (*models.PasswordResetToken, error) {
	if t, ok := f.byUser[userID]; ok && t != nil {
		return t, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeResetStore) RecordFailedAttempt(_ context.Context, _ uuid.UUID) (int, bool, error) {
	if f.failAttempt {
		return 6, true, nil
	}
	return 1, false, nil
}

func (f *fakeResetStore) MarkUsed(_ context.Context, tokenID uuid.UUID) error {
	f.markUsedCalls = append(f.markUsedCalls, tokenID)
	return nil
}

func (f *fakeResetStore) HardDeleteOldTokens(context.Context) (int64, error) {
	return 0, nil
}

var _ database.PasswordResetStore = (*fakeResetStore)(nil)

// fakeEmailSender is a buffer-aware Sender so tests can verify the email
// body without touching the network. Thread-safe.
type fakeEmailSender struct {
	mu       sync.Mutex
	codes    []string
	users    []models.User
	locales  []string
	welcomes int
	resets   int
	failNext bool
}

func (f *fakeEmailSender) SendWelcome(_ context.Context, user models.User, locale string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext {
		f.failNext = false
		return errors.New("simulated send failure")
	}
	f.welcomes++
	f.users = append(f.users, user)
	f.locales = append(f.locales, locale)
	return nil
}

func (f *fakeEmailSender) SendPasswordReset(_ context.Context, user models.User, code, locale string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext {
		f.failNext = false
		return errors.New("simulated send failure")
	}
	f.resets++
	f.codes = append(f.codes, code)
	f.users = append(f.users, user)
	f.locales = append(f.locales, locale)
	return nil
}

func (f *fakeEmailSender) lastResetCode() (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.codes) == 0 {
		return "", false
	}
	return f.codes[len(f.codes)-1], true
}

var _ email.Sender = (*fakeEmailSender)(nil)

// ---------- Wiring helpers ----------------------------------------------

func newHandlers(authStore database.AuthStore, sessStore database.SessionStore, resetStore database.PasswordResetStore, emailSender email.Sender) *Handlers {
	return &Handlers{
		AuthRepo:    authStore,
		SessionRepo: sessStore,
		ResetRepo:   resetStore,
		EmailSender: emailSender,
		Config:      config.Load(),
	}
}

func doPost(t *testing.T, h *Handlers, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	buf, _ := json.Marshal(body)
	r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(buf))
	r.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		r.AddCookie(c)
	}
	r.RemoteAddr = "203.0.113.1:54321"
	rw := httptest.NewRecorder()
	router := newTestRouter(h)
	router.ServeHTTP(rw, r)
	return rw
}

func doGet(t *testing.T, h *Handlers, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}
	rw := httptest.NewRecorder()
	router := newTestRouter(h)
	router.ServeHTTP(rw, r)
	return rw
}

func doDelete(t *testing.T, h *Handlers, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodDelete, path, nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}
	rw := httptest.NewRecorder()
	router := newTestRouter(h)
	router.ServeHTTP(rw, r)
	return rw
}

func newTestRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	// Public auth endpoints (no middleware).
	mux.HandleFunc("/auth/v2/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		h.RegisterOpaque(w, r)
	})
	mux.HandleFunc("/auth/v2/forgot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		h.ForgotOpaque(w, r)
	})
	mux.HandleFunc("/auth/v2/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		h.ResetOpaque(w, r)
	})
	mux.HandleFunc("/auth/v2/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		h.RefreshOpaque(w, r)
	})
	mux.HandleFunc("/auth/v2/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		h.LoginOpaque(w, r)
	})

	// Authenticated endpoints behind AuthMiddlewareV2.
	protected := middleware.AuthMiddlewareV2(h.SessionRepo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/auth/v2/logout" && r.Method == http.MethodPost:
			h.LogoutOpaque(w, r)
		case path == "/auth/v2/me" && r.Method == http.MethodGet:
			h.MeOpaque(w, r)
		case path == "/auth/v2/account" && r.Method == http.MethodDelete:
			h.DeleteAccountOpaque(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.Handle("/auth/v2/logout", protected)
	mux.Handle("/auth/v2/me", protected)
	mux.Handle("/auth/v2/account", protected)

	return mux
}

func putUserIDIntoContext(r *http.Request, uid uuid.UUID) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyUserId{}, uid)
	return r.WithContext(ctx)
}

// ---------- LoginOpaque tests --------------------------------------------

func TestLoginOpaque_EmptyBody(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/v2/login", nil)
	h.LoginOpaque(rw, r)
	require.Equal(t, http.StatusBadRequest, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeValidationError, out.Error.Code)
}

func TestLoginOpaque_MissingFields(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doPost(t, h, "/auth/v2/login", map[string]string{"email": "x@x.com"})
	require.Equal(t, http.StatusBadRequest, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeValidationError, out.Error.Code)
	require.NotNil(t, out.Error.Fields)
	assert.Equal(t, "REQUIRED", out.Error.Fields["password"])
}

func TestLoginOpaque_UnknownUser(t *testing.T) {
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{}}
	h := newHandlers(authStore, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doPost(t, h, "/auth/v2/login",
		map[string]string{"email": "nobody@here.test", "password": "Pa55word!"})
	require.Equal(t, http.StatusUnauthorized, rw.Code)
}

func TestLoginOpaque_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("right-password"), bcrypt.MinCost)
	u := &models.User{ID: uuid.New(), Email: "x@y.z", PasswordHash: string(hash)}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{"x@y.z": u}}
	h := newHandlers(authStore, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doPost(t, h, "/auth/v2/login",
		map[string]string{"email": "x@y.z", "password": "wrong-password"})
	require.Equal(t, http.StatusUnauthorized, rw.Code)
}

// TestLoginOpaque_SoftDeletedBlocked: a soft-deleted user with the
// correct password must NOT be able to sign in. Caught by the smoke
// test on 2026-08-04: the original LoginOpaque skipped the DeletedAt
// check, so the deleted account could re-login and silently undelete
// itself via the refresh endpoint.
func TestLoginOpaque_SoftDeletedBlocked(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("right-password"), bcrypt.MinCost)
	deletedAt := "2026-08-04T00:00:00Z"
	u := &models.User{
		ID:           uuid.New(),
		Email:        "deleted@x.z",
		PasswordHash: string(hash),
		DeletedAt:    &deletedAt,
	}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{u.Email: u}}
	h := newHandlers(authStore, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doPost(t, h, "/auth/v2/login",
		map[string]string{"email": u.Email, "password": "right-password"})
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	assert.Equal(t, CodeInvalidCredentials, mustDecodeError(t, rw).Code,
		"deleted accounts must NOT surface a distinguishable error code")
}

func TestLoginOpaque_Success_ReturnsTokensAndSetsCookies(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Pa55word!"), bcrypt.MinCost)
	uid := uuid.New()
	u := &models.User{
		ID:           uid,
		Email:        "user@example.test",
		PasswordHash: string(hash),
		Tier:         "free",
		Locale:       "es",
	}
	authStore := &fakeAuthStore{
		byEmail: map[string]*models.User{u.Email: u},
		byID:    map[uuid.UUID]*models.User{uid: u},
	}
	sessStore := &fakeSessionStore{}
	h := newHandlers(authStore, sessStore, newFakeResetStore(), &fakeEmailSender{})

	rw := doPost(t, h, "/auth/v2/login",
		map[string]string{"email": u.Email, "password": "Pa55word!"})
	require.Equal(t, http.StatusOK, rw.Code, "body: %s", rw.Body.String())

	cookies := rw.Result().Cookies()
	require.Len(t, cookies, 2)
	for _, c := range cookies {
		assert.Equal(t, "/", c.Path)
		assert.True(t, c.HttpOnly)
		assert.True(t, strings.Contains(c.Name, "itinera_"))
		assert.Len(t, c.Value, 64)
		assert.True(t, isLowerHex(c.Value))
	}

	assert.Equal(t, 1, sessStore.createdCalls)
}

// ---------- LogoutOpaque tests --------------------------------------------

func TestLogoutOpaque_NoCookie_StillClears(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doPost(t, h, "/auth/v2/logout", nil)
	require.Equal(t, http.StatusNoContent, rw.Code)
	cookies := rw.Result().Cookies()
	assert.GreaterOrEqual(t, len(cookies), 2)
	for _, c := range cookies {
		assert.Equal(t, -1, c.MaxAge)
	}
}

func TestLogoutOpaque_WithCookie_RevokesSession(t *testing.T) {
	sessStore := &fakeSessionStore{}
	h := newHandlers(&fakeAuthStore{}, sessStore, newFakeResetStore(), &fakeEmailSender{})

	raw := "raw-access-cookie-value-1234567890abcdef"
	cookie := &http.Cookie{Name: middleware.CookieAccessToken, Value: raw}

	rw := doPost(t, h, "/auth/v2/logout", nil, cookie)
	require.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, 1, sessStore.revokedByHashCalls)
	assert.Equal(t, auth.HashToken(raw), sessStore.lastRevokedHash)
}

// ---------- MeOpaque tests -------------------------------------------------

func TestMeOpaque_NoAuth(t *testing.T) {
	h := newHandlers(&fakeAuthStore{byID: map[uuid.UUID]*models.User{}}, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doGet(t, h, "/auth/v2/me")
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeUnauthenticated, out.Error.Code)
}

func TestMeOpaque_Authenticated_Direct(t *testing.T) {
	uid := uuid.New()
	u := &models.User{ID: uid, Email: "u@example.test", Tier: "free"}
	authStore := &fakeAuthStore{
		byEmail: map[string]*models.User{u.Email: u},
		byID:    map[uuid.UUID]*models.User{uid: u},
	}
	h := newHandlers(authStore, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})

	r := httptest.NewRequest(http.MethodGet, "/auth/v2/me", nil)
	r = putUserIDIntoContext(r, uid)
	rw := httptest.NewRecorder()
	h.MeOpaque(rw, r)

	require.Equal(t, http.StatusOK, rw.Code)
	var got models.User
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&got))
	assert.Equal(t, u.Email, got.Email)
	assert.Equal(t, "free", got.Tier)
}

func TestMeOpaque_SoftDeleted_Direct(t *testing.T) {
	uid := uuid.New()
	deletedAt := "2026-08-01T00:00:00Z"
	u := &models.User{ID: uid, Email: "x@y.z", DeletedAt: &deletedAt, Tier: "free"}
	authStore := &fakeAuthStore{byID: map[uuid.UUID]*models.User{uid: u}}
	h := newHandlers(authStore, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})

	r := httptest.NewRequest(http.MethodGet, "/auth/v2/me", nil)
	r = putUserIDIntoContext(r, uid)
	rw := httptest.NewRecorder()
	h.MeOpaque(rw, r)

	require.Equal(t, http.StatusForbidden, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeAccountDeleted, out.Error.Code)
}

// ---------- RefreshOpaque tests --------------------------------------------

func TestRefreshOpaque_NoCookie(t *testing.T) {
	sessStore := &fakeSessionStore{}
	h := newHandlers(&fakeAuthStore{}, sessStore, newFakeResetStore(), &fakeEmailSender{})
	rw := doPost(t, h, "/auth/v2/refresh", nil)
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	assert.Equal(t, 0, sessStore.rotatedCalls)
}

func TestRefreshOpaque_UnknownRefresh(t *testing.T) {
	sessStore := &fakeSessionStore{byRefreshHash: map[string]*models.Session{}}
	h := newHandlers(&fakeAuthStore{}, sessStore, newFakeResetStore(), &fakeEmailSender{})
	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: "any-value"}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)
	require.Equal(t, http.StatusUnauthorized, rw.Code)
}

func TestRefreshOpaque_Success_RotatesAndSetsCookies(t *testing.T) {
	family := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	refreshRaw := "raw-refresh-cookie-aaaaaaaabbbbbbbb"
	sessStore := &fakeSessionStore{
		byRefreshHash: map[string]*models.Session{
			auth.HashToken(refreshRaw): {
				ID:               uuid.New(),
				UserID:           uuid.New(),
				RefreshTokenHash: auth.HashToken(refreshRaw),
				RefreshFamily:    family,
				ExpiresAt:        expiresAt,
			},
		},
	}
	h := newHandlers(&fakeAuthStore{}, sessStore, newFakeResetStore(), &fakeEmailSender{})
	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: refreshRaw}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)

	require.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, 1, sessStore.rotatedCalls)

	cookies := rw.Result().Cookies()
	require.Len(t, cookies, 2)
	var accessValue, refreshValue string
	for _, c := range cookies {
		if c.Name == "itinera_access" {
			accessValue = c.Value
		}
		if c.Name == "itinera_refresh" {
			refreshValue = c.Value
		}
	}
	assert.NotEqual(t, accessValue, refreshValue)
	assert.NotEqual(t, refreshValue, refreshRaw)
	assert.Empty(t, sessStore.revokedFamilies)
}

func TestRefreshOpaque_ReuseDetected_KillsFamily(t *testing.T) {
	family := uuid.New()
	revokedAt := time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339)
	refreshRaw := "stolen-refresh-cookie-ccccddddeee"
	sessStore := &fakeSessionStore{
		byRefreshHash: map[string]*models.Session{
			auth.HashToken(refreshRaw): {
				ID:               uuid.New(),
				UserID:           uuid.New(),
				RefreshTokenHash: auth.HashToken(refreshRaw),
				RefreshFamily:    family,
				RevokedAt:        &revokedAt,
				ExpiresAt:        time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339),
			},
		},
	}
	h := newHandlers(&fakeAuthStore{}, sessStore, newFakeResetStore(), &fakeEmailSender{})
	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: refreshRaw}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)
	require.Equal(t, http.StatusForbidden, rw.Code)
	assert.Equal(t, CodeReuseDetected, mustDecodeError(t, rw).Code)
	require.Len(t, sessStore.revokedFamilies, 1)
	assert.Equal(t, family, sessStore.revokedFamilies[0])
	assert.Equal(t, 0, sessStore.rotatedCalls)
}

func TestRefreshOpaque_ExpiredRefresh(t *testing.T) {
	family := uuid.New()
	expiresAt := time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339)
	refreshRaw := "expired-refresh-cookie-ffffffffff"
	sessStore := &fakeSessionStore{
		byRefreshHash: map[string]*models.Session{
			auth.HashToken(refreshRaw): {
				ID:               uuid.New(),
				UserID:           uuid.New(),
				RefreshTokenHash: auth.HashToken(refreshRaw),
				RefreshFamily:    family,
				ExpiresAt:        expiresAt,
			},
		},
	}
	h := newHandlers(&fakeAuthStore{}, sessStore, newFakeResetStore(), &fakeEmailSender{})
	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: refreshRaw}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	assert.Equal(t, CodeSessionExpired, mustDecodeError(t, rw).Code)
}

// ---------- RegisterOpaque tests ------------------------------------------

func TestRegisterOpaque_WeakPassword(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doPost(t, h, "/auth/v2/register",
		map[string]string{"email": "x@y.z", "password": "short"})
	require.Equal(t, http.StatusBadRequest, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeWeakPassword, out.Error.Code)
}

func TestRegisterOpaque_Success_CreatesUserAndSetsCookies(t *testing.T) {
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{}, byID: map[uuid.UUID]*models.User{}}
	sessStore := &fakeSessionStore{}
	mailer := &fakeEmailSender{}
	h := newHandlers(authStore, sessStore, newFakeResetStore(), mailer)

	rw := doPost(t, h, "/auth/v2/register",
		map[string]string{"email": "new@example.test", "password": "GoodPass1!", "locale": "es"})
	require.Equal(t, http.StatusOK, rw.Code)

	cookies := rw.Result().Cookies()
	assert.GreaterOrEqual(t, len(cookies), 2)

	// Welcome email was sent to the new user.
	mailer.mu.Lock()
	assert.Equal(t, 1, mailer.welcomes)
	assert.Equal(t, "new@example.test", mailer.users[0].Email)
	assert.Equal(t, "es", mailer.locales[0])
	mailer.mu.Unlock()

	// Session was created.
	assert.Equal(t, 1, sessStore.createdCalls)
}

// ---------- ForgotOpaque tests --------------------------------------------

func TestForgotOpaque_UnknownEmail_ReturnsAck(t *testing.T) {
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{}}
	resetStore := newFakeResetStore()
	mailer := &fakeEmailSender{}
	h := newHandlers(authStore, &fakeSessionStore{}, resetStore, mailer)

	rw := doPost(t, h, "/auth/v2/forgot", map[string]string{"email": "ghost@example.test"})
	require.Equal(t, http.StatusAccepted, rw.Code)
	var out forgotResponse
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, forgotAck, out.Message)
	assert.Equal(t, 0, resetStore.createCalls, "no code should be persisted for unknown email")
	mailer.mu.Lock()
	assert.Equal(t, 0, mailer.resets)
	mailer.mu.Unlock()
}

func TestForgotOpaque_KnownEmail_PersistsAndSendsCode(t *testing.T) {
	uid := uuid.New()
	u := &models.User{ID: uid, Email: "real@example.test", Locale: "ja", Tier: "free"}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{u.Email: u}, byID: map[uuid.UUID]*models.User{uid: u}}
	resetStore := newFakeResetStore()
	mailer := &fakeEmailSender{}
	h := newHandlers(authStore, &fakeSessionStore{}, resetStore, mailer)

	rw := doPost(t, h, "/auth/v2/forgot", map[string]string{"email": u.Email, "locale": "ja"})
	require.Equal(t, http.StatusAccepted, rw.Code)

	assert.Equal(t, 1, resetStore.createCalls, "the code must be persisted")
	mailer.mu.Lock()
	require.Equal(t, 1, mailer.resets, "the email must be sent")
	// lastResetCode also takes mu internally — release here so the helper
	// doesn't re-enter on the same goroutine and deadlock against itself.
	mailer.mu.Unlock()
	code, ok := mailer.lastResetCode()
	require.True(t, ok)
	assert.Len(t, code, 6, "code must be 6 digits")
}

func TestForgotOpaque_SoftDeletedUser_ReturnsAckWithoutEmail(t *testing.T) {
	deletedAt := "2026-08-01T00:00:00Z"
	u := &models.User{ID: uuid.New(), Email: "ghost2@example.test", DeletedAt: &deletedAt}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{u.Email: u}}
	resetStore := newFakeResetStore()
	mailer := &fakeEmailSender{}
	h := newHandlers(authStore, &fakeSessionStore{}, resetStore, mailer)

	rw := doPost(t, h, "/auth/v2/forgot", map[string]string{"email": u.Email})
	require.Equal(t, http.StatusAccepted, rw.Code)
	assert.Equal(t, 0, resetStore.createCalls, "soft-deleted users must NOT receive a code")
}

// ---------- ResetOpaque tests --------------------------------------------

func TestResetOpaque_WrongCode(t *testing.T) {
	uid := uuid.New()
	u := &models.User{ID: uid, Email: "u@y.z"}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{u.Email: u}}
	h := newHandlers(authStore, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})

	rw := doPost(t, h, "/auth/v2/reset",
		map[string]string{"email": u.Email, "code": "000000", "new_password": "NewPass1!"})
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	assert.Equal(t, CodeInvalidToken, mustDecodeError(t, rw).Code)
}

func TestResetOpaque_WeakPassword(t *testing.T) {
	uid := uuid.New()
	u := &models.User{ID: uid, Email: "u@y.z"}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{u.Email: u}}
	h := newHandlers(authStore, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})

	rw := doPost(t, h, "/auth/v2/reset",
		map[string]string{"email": u.Email, "code": "000000", "new_password": "short"})
	require.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, CodeWeakPassword, mustDecodeError(t, rw).Code)
}

func TestResetOpaque_HappyPath_UpdatesPasswordAndRevokesSessions(t *testing.T) {
	uid := uuid.New()
	u := &models.User{ID: uid, Email: "u@y.z"}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{u.Email: u}}

	// Pre-seed a valid code in the reset store.
	code := "482915"
	hash := auth.HashToken(code)
	resetStore := newFakeResetStore()
	resetStore.byHash[hash] = &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    uid,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	}

	sessStore := &fakeSessionStore{}
	h := newHandlers(authStore, sessStore, resetStore, &fakeEmailSender{})

	rw := doPost(t, h, "/auth/v2/reset",
		map[string]string{"email": u.Email, "code": code, "new_password": "NewPass1!"})
	require.Equal(t, http.StatusNoContent, rw.Code)

	assert.Equal(t, 1, len(authStore.passwordUpdateCalls))
	assert.Equal(t, uid, authStore.passwordUpdateCalls[0])
	assert.Equal(t, 1, len(resetStore.markUsedCalls))
	assert.Equal(t, 1, len(sessStore.revokedAllFor))
	assert.Equal(t, uid, sessStore.revokedAllFor[0])
}

// ---------- Utilities -----------------------------------------------------

// ---------- DeleteAccountOpaque tests -------------------------------------

// TestDeleteAccountOpaque_NoAuth verifies the middleware short-circuits
// unauthenticated calls with UNAUTHENTICATED.
func TestDeleteAccountOpaque_NoAuth(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{}, newFakeResetStore(), &fakeEmailSender{})
	rw := doDelete(t, h, "/auth/v2/account")
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	assert.Equal(t, CodeUnauthenticated, mustDecodeError(t, rw).Code)
}

// TestDeleteAccountOpaque_HappyPath walks the full GDPR cascade path:
// middleware authenticates → handler reads uid → cascade runs →
// 204 + cookies cleared. The fakeAuthStore records the userID so the
// test can assert the right user was soft-deleted.
//
// The middleware requires a cookie that maps to a session row, so the
// test seeds fakeSessionStore.byHash with a precomputed hash. We compute
// the raw value and let the store SHA-256 it via auth.HashToken.
func TestDeleteAccountOpaque_HappyPath(t *testing.T) {
	uid := uuid.New()
	authStore := &fakeAuthStore{byID: map[uuid.UUID]*models.User{uid: {ID: uid}}}
	sessStore := &fakeSessionStore{
		byRefreshHash: map[string]*models.Session{},
	}
	rawCookie := "happy-path-access-cookie-1234567890"
	sessStore.byAccessHash = map[string]*models.Session{
		auth.HashToken(rawCookie): {
			ID:     uuid.New(),
			UserID: uid,
		},
	}
	h := newHandlers(authStore, sessStore, newFakeResetStore(), &fakeEmailSender{})

	rw := doDelete(t, h, "/auth/v2/account", &http.Cookie{
		Name:  middleware.CookieAccessToken,
		Value: rawCookie,
	})
	require.Equal(t, http.StatusNoContent, rw.Code, "body: %s", rw.Body.String())
	assert.Equal(t, 1, len(authStore.softDeleteCascadeCalls))
	assert.Equal(t, uid, authStore.softDeleteCascadeCalls[0])

	// Cookies must be cleared on the way out.
	cookies := rw.Result().Cookies()
	assert.GreaterOrEqual(t, len(cookies), 2, "expect both itinera_access and itinera_refresh cleared")
	for _, c := range cookies {
		assert.Equal(t, -1, c.MaxAge)
	}
}

// TestDeleteAccountOpaque_DBError: cascade failure returns 500 with a
// generic INTERNAL_ERROR (no err.Error() leak per Spec §9.3).
func TestDeleteAccountOpaque_DBError(t *testing.T) {
	uid := uuid.New()
	authStore := &fakeAuthStore{
		byID:                 map[uuid.UUID]*models.User{uid: {ID: uid}},
		softDeleteCascadeErr: errors.New("connection reset"),
	}
	sessStore := &fakeSessionStore{
		byAccessHash: map[string]*models.Session{
			"any-hash": {ID: uuid.New(), UserID: uid},
		},
	}
	h := newHandlers(authStore, sessStore, newFakeResetStore(), &fakeEmailSender{})

	// Bypass the middleware by setting the context directly: this test
	// exercises the handler's error-mapping branch, not the auth flow.
	r := httptest.NewRequest(http.MethodDelete, "/auth/v2/account", nil)
	r = putUserIDIntoContext(r, uid)
	rw := httptest.NewRecorder()
	h.DeleteAccountOpaque(rw, r)
	require.Equal(t, http.StatusInternalServerError, rw.Code)
	out := mustDecodeError(t, rw)
	assert.Equal(t, CodeInternalError, out.Code)
	assert.NotContains(t, out.Message, "connection reset",
		"raw DB error must NOT leak into the client-facing message")
}

// ---------- Utilities -----------------------------------------------------

// TestStripPort covers the helper used to coerce http.Request.RemoteAddr
// into a Postgres INET-compatible value. The function MUST handle IPv6
// (which Go reports as "[::1]:54321") without dropping to "[" — that
// mistake was caught by the smoke test on 2026-08-04 and would have
// silently broken every register/login attempt over IPv6.
func TestStripPort(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ipv4 with port", "127.0.0.1:54321", "127.0.0.1"},
		{"ipv4 no port", "127.0.0.1", "127.0.0.1"},
		{"ipv6 with port", "[::1]:54321", "::1"},
		{"ipv6 no port", "::1", "::1"},
		{"empty", "", ""},
		{"hostname with port", "localhost:8080", "localhost"},
		{"hostname no port", "localhost", "localhost"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, stripPort(c.in))
		})
	}
}

func isLowerHex(s string) bool {
	for _, r := range s {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}
	return true
}

func mustDecodeError(t *testing.T, rw *httptest.ResponseRecorder) APIError {
	t.Helper()
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	return out.Error
}
