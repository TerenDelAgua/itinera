package handlers

import (
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/http/middleware"
	"backend/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
// Only the methods the tests cover are non-trivial; the rest panic so a
// regression is loud.
type fakeAuthStore struct {
	byEmail map[string]*models.User
	byID    map[uuid.UUID]*models.User
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

func (f *fakeAuthStore) CreateUser(context.Context, string, string, string) (*models.User, error) {
	panic("not used in tests")
}
func (f *fakeAuthStore) ClaimGuestTrips(context.Context, string, uuid.UUID) (int, error) {
	panic("not used in tests")
}
func (f *fakeAuthStore) SoftDeleteUser(context.Context, uuid.UUID) (string, error) {
	panic("not used in tests")
}
func (f *fakeAuthStore) MarkUserAsHardDeleted(context.Context, uuid.UUID, string) error {
	panic("not used in tests")
}
func (f *fakeAuthStore) UpdateUserPassword(context.Context, uuid.UUID, string) error {
	panic("not used in tests")
}

var _ database.AuthStore = (*fakeAuthStore)(nil)

// fakeSessionStore implements database.SessionStore with counters and a
// small session map keyed by refresh-hash. Each test seeds the map with the
// scenarios it needs (active session, revoked session, unknown hash, ...)
// and asserts behaviour by reading back the counters.
type fakeSessionStore struct {
	createdCalls       int
	revokedByHashCalls int
	lastRevokedHash    string
	rotatedCalls       int

	// byRefreshHash is the map the Find* queries read from. Tests can
	// pre-seed it with a row whose RefreshFamily / RevokedAt match the
	// scenario they want to exercise.
	byRefreshHash map[string]*models.Session

	// revokedFamilies accumulates UUIDs passed to RevokeFamily so a test
	// can assert that the reuse-detection path actually fires.
	revokedFamilies []uuid.UUID
}

func (f *fakeSessionStore) CreateSession(_ context.Context, _ uuid.UUID, _ string, _ string, _ time.Time, _, _ *string) (*models.Session, error) {
	f.createdCalls++
	return nil, nil
}

func (f *fakeSessionStore) FindSessionByAccessTokenHash(_ context.Context, accessHash string) (*models.Session, error) {
	if accessHash == "" {
		return nil, pgx.ErrNoRows
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
func (f *fakeSessionStore) RevokeAllSessionsForUser(context.Context, uuid.UUID) error {
	panic("not used")
}
func (f *fakeSessionStore) CountActiveSessionsForUser(context.Context, uuid.UUID) (int, error) {
	panic("not used")
}
func (f *fakeSessionStore) CleanupExpiredSessions(context.Context) (int64, error) {
	panic("not used")
}

var _ database.SessionStore = (*fakeSessionStore)(nil)

// ---------- Wiring helpers -----------------------------------------------

func newHandlers(authStore database.AuthStore, sessStore database.SessionStore) *Handlers {
	return &Handlers{
		AuthRepo:    authStore,
		SessionRepo: sessStore,
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

// newTestRouter wires only AuthMiddlewareV2 in front of logout+me so the
// tests don't have to set JWT cookies (which is the legacy tree, not the
// one we're exercising here). Refresh is mounted WITHOUT AuthMiddlewareV2
// because, by definition, refresh is called when the access cookie is
// missing or expired.
func newTestRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	// Refresh stays at the root of the mux (no middleware).
	mux.HandleFunc("/auth/v2/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		h.RefreshOpaque(w, r)
	})

	// Everything else goes through the opaque middleware to emulate the
	// real router tree.
	protected := middleware.AuthMiddlewareV2(h.SessionRepo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/auth/v2/logout" && r.Method == http.MethodPost:
			h.LogoutOpaque(w, r)
		case path == "/auth/v2/me" && r.Method == http.MethodGet:
			h.MeOpaque(w, r)
		case path == "/auth/v2/login" && r.Method == http.MethodPost:
			h.LoginOpaque(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.Handle("/auth/v2/logout", protected)
	mux.Handle("/auth/v2/me", protected)
	mux.Handle("/auth/v2/login", protected)

	return mux
}

// putUserIDIntoContext returns a request whose Context already has
// ContextKeyUserId set, simulating the middleware having run. Use this
// when testing MeOpaque() in isolation; the doGet/dopost helpers route
// through the middleware which does the same.
func putUserIDIntoContext(r *http.Request, uid uuid.UUID) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyUserId{}, uid)
	return r.WithContext(ctx)
}

// ---------- LoginOpaque tests (kept from prior phase) ---------------------

func TestLoginOpaque_EmptyBody(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{})
	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/v2/login", nil)
	h.LoginOpaque(rw, r)
	require.Equal(t, http.StatusBadRequest, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeValidationError, out.Error.Code)
}

func TestLoginOpaque_MissingFields(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{})
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
	h := newHandlers(authStore, &fakeSessionStore{})
	rw := doPost(t, h, "/auth/v2/login",
		map[string]string{"email": "nobody@here.test", "password": "Pa55word!"})
	require.Equal(t, http.StatusUnauthorized, rw.Code)
}

func TestLoginOpaque_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("right-password"), bcrypt.MinCost)
	u := &models.User{ID: uuid.New(), Email: "x@y.z", PasswordHash: string(hash)}
	authStore := &fakeAuthStore{byEmail: map[string]*models.User{"x@y.z": u}}
	h := newHandlers(authStore, &fakeSessionStore{})
	rw := doPost(t, h, "/auth/v2/login",
		map[string]string{"email": "x@y.z", "password": "wrong-password"})
	require.Equal(t, http.StatusUnauthorized, rw.Code)
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
	h := newHandlers(authStore, sessStore)

	rw := doPost(t, h, "/auth/v2/login",
		map[string]string{"email": u.Email, "password": "Pa55word!"})
	require.Equal(t, http.StatusOK, rw.Code, "body: %s", rw.Body.String())

	cookies := rw.Result().Cookies()
	require.Len(t, cookies, 2, "expected exactly 2 cookies (access + refresh)")
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

// TestLogoutOpaque_NoCookie_StillClears: middleware doesn't authenticate,
// handler short-circuits, cookies are still cleared (so the browser is
// left clean even if the user already lost the cookie).
func TestLogoutOpaque_NoCookie_StillClears(t *testing.T) {
	h := newHandlers(&fakeAuthStore{}, &fakeSessionStore{})
	rw := doPost(t, h, "/auth/v2/logout", nil)
	require.Equal(t, http.StatusNoContent, rw.Code)
	cookies := rw.Result().Cookies()
	// ClearAuthCookies clears itinera_access + itinera_refresh = 2 cookies.
	assert.GreaterOrEqual(t, len(cookies), 2)
	for _, c := range cookies {
		assert.Equal(t, -1, c.MaxAge, "clear cookies must have MaxAge=-1")
	}
}

// TestLogoutOpaque_WithCookie_RevokesSession: the supplied cookie's hash
// reaches RevokeSessionByAccessHash exactly once.
func TestLogoutOpaque_WithCookie_RevokesSession(t *testing.T) {
	sessStore := &fakeSessionStore{}
	h := newHandlers(&fakeAuthStore{}, sessStore)

	raw := "raw-access-cookie-value-1234567890abcdef"
	cookie := &http.Cookie{Name: middleware.CookieAccessToken, Value: raw}

	rw := doPost(t, h, "/auth/v2/logout", nil, cookie)
	require.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, 1, sessStore.revokedByHashCalls)
	assert.Equal(t, auth.HashToken(raw), sessStore.lastRevokedHash,
		"the cookie value must be hashed before being passed to RevokeSessionByAccessHash")
}

// ---------- MeOpaque tests -------------------------------------------------

// TestMeOpaque_NoAuth: middleware didn't authenticate → UNAUTHENTICATED.
func TestMeOpaque_NoAuth(t *testing.T) {
	h := newHandlers(&fakeAuthStore{byID: map[uuid.UUID]*models.User{}}, &fakeSessionStore{})
	rw := doGet(t, h, "/auth/v2/me")
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeUnauthenticated, out.Error.Code)
}

// TestMeOpaque_Authenticated: middleware set the user → handler returns it.
// We bypass the middleware here and exercise MeOpaque directly because the
// real middleware doesn't populate byID in the fake; future DB-backed tests
// will cover the end-to-end happy path.
func TestMeOpaque_Authenticated_Direct(t *testing.T) {
	uid := uuid.New()
	u := &models.User{
		ID:    uid,
		Email: "u@example.test",
		Tier:  "free",
	}
	authStore := &fakeAuthStore{
		byEmail: map[string]*models.User{u.Email: u},
		byID:    map[uuid.UUID]*models.User{uid: u},
	}
	h := newHandlers(authStore, &fakeSessionStore{})

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

// TestMeOpaque_SoftDeleted returns ACCOUNT_DELETED.
func TestMeOpaque_SoftDeleted_Direct(t *testing.T) {
	uid := uuid.New()
	deletedAt := "2026-08-01T00:00:00Z"
	u := &models.User{ID: uid, Email: "x@y.z", DeletedAt: &deletedAt, Tier: "free"}
	authStore := &fakeAuthStore{byID: map[uuid.UUID]*models.User{uid: u}}
	h := newHandlers(authStore, &fakeSessionStore{})

	r := httptest.NewRequest(http.MethodGet, "/auth/v2/me", nil)
	r = putUserIDIntoContext(r, uid)
	rw := httptest.NewRecorder()
	h.MeOpaque(rw, r)

	require.Equal(t, http.StatusForbidden, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeAccountDeleted, out.Error.Code)
}

// ---------- Utilities -----------------------------------------------------

// isLowerHex is a tiny smoke check used by the success-path test to confirm
// the cookie value matches the format NewSecureToken produces.
func isLowerHex(s string) bool {
	for _, r := range s {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}
	return true
}

// ---------- RefreshOpaque tests --------------------------------------------

// TestRefreshOpaque_NoCookie: missing refresh cookie → 401, NO rotation.
func TestRefreshOpaque_NoCookie(t *testing.T) {
	sessStore := &fakeSessionStore{}
	h := newHandlers(&fakeAuthStore{}, sessStore)

	rw := doPost(t, h, "/auth/v2/refresh", nil)
	require.Equal(t, http.StatusUnauthorized, rw.Code)

	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeUnauthenticated, out.Error.Code)
	assert.Equal(t, 0, sessStore.rotatedCalls)
}

// TestRefreshOpaque_UnknownRefresh: cookie value doesn't resolve to any
// session → 401. Same anti-enumeration copy as the unknown-user branch.
func TestRefreshOpaque_UnknownRefresh(t *testing.T) {
	sessStore := &fakeSessionStore{byRefreshHash: map[string]*models.Session{}}
	h := newHandlers(&fakeAuthStore{}, sessStore)

	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: "any-value"}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	assert.Equal(t, 0, sessStore.rotatedCalls)
}

// TestRefreshOpaque_Success_RotatesAndSetsCookies: real row in store →
// rotation, two new cookies.
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
	h := newHandlers(&fakeAuthStore{}, sessStore)

	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: refreshRaw}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)

	require.Equal(t, http.StatusOK, rw.Code, "body: %s", rw.Body.String())
	assert.Equal(t, 1, sessStore.rotatedCalls)

	cookies := rw.Result().Cookies()
	require.Len(t, cookies, 2)
	gotAccess, gotRefresh := false, false
	var accessValue, refreshValue string
	for _, c := range cookies {
		switch c.Name {
		case "itinera_access":
			gotAccess = true
			accessValue = c.Value
		case "itinera_refresh":
			gotRefresh = true
			refreshValue = c.Value
		}
	}
	assert.True(t, gotAccess)
	assert.True(t, gotRefresh)
	assert.NotEqual(t, accessValue, refreshValue, "new access and refresh must be different secrets")
	assert.NotEqual(t, refreshValue, refreshRaw, "new refresh must NOT be the old one (defeats the rotation)")

	// No family-wide revocation in the happy path.
	assert.Empty(t, sessStore.revokedFamilies)
}

// TestRefreshOpaque_ReuseDetected_KillsFamily: row's RevokedAt set →
// RevokeFamily invoked with the row's family, 403 TOKEN_REUSE_DETECTED.
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
	h := newHandlers(&fakeAuthStore{}, sessStore)

	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: refreshRaw}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)

	require.Equal(t, http.StatusForbidden, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeReuseDetected, out.Error.Code)
	require.Len(t, sessStore.revokedFamilies, 1, "RevokeFamily must fire once on reuse detection")
	assert.Equal(t, family, sessStore.revokedFamilies[0],
		"RevokeFamily must be called with the row's family")
	assert.Equal(t, 0, sessStore.rotatedCalls, "no rotation on reuse-detected requests")
}

// TestRefreshOpaque_ExpiredRefresh: ExpiresAt in the past → 401, no rotation.
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
	h := newHandlers(&fakeAuthStore{}, sessStore)

	cookie := &http.Cookie{Name: middleware.CookieRefreshToken, Value: refreshRaw}
	rw := doPost(t, h, "/auth/v2/refresh", nil, cookie)

	require.Equal(t, http.StatusUnauthorized, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeSessionExpired, out.Error.Code)
	assert.Equal(t, 0, sessStore.rotatedCalls)
}
