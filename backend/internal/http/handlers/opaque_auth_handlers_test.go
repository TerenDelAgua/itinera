package handlers

import (
	"backend/internal/config"
	"backend/internal/database"
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

// fakeAuthStore implements database.AuthStore with hand-rolled behaviour.
// Only the methods LoginOpaque calls are non-trivial — the rest panic so a
// regression that calls e.g. UpdateUserPassword is loud.
type fakeAuthStore struct {
	userByEmail map[string]*models.User
}

func (f *fakeAuthStore) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	if u, ok := f.userByEmail[email]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeAuthStore) CreateUser(context.Context, string, string, string) (*models.User, error) {
	panic("CreateUser not implemented in fake")
}
func (f *fakeAuthStore) GetUserByID(context.Context, uuid.UUID) (*models.User, error) {
	panic("not used by LoginOpaque")
}
func (f *fakeAuthStore) ClaimGuestTrips(context.Context, string, uuid.UUID) (int, error) {
	panic("not used by LoginOpaque")
}
func (f *fakeAuthStore) SoftDeleteUser(context.Context, uuid.UUID) (string, error) {
	panic("not used by LoginOpaque")
}
func (f *fakeAuthStore) MarkUserAsHardDeleted(context.Context, uuid.UUID, string) error {
	panic("not used by LoginOpaque")
}
func (f *fakeAuthStore) UpdateUserPassword(context.Context, uuid.UUID, string) error {
	panic("not used by LoginOpaque")
}

var _ database.AuthStore = (*fakeAuthStore)(nil)

// fakeSessionStore implements database.SessionStore with a counter so the
// success path can verify a session row was created with the expected hashes.
type fakeSessionStore struct {
	createdCalls int
	lastUserID   uuid.UUID
	lastAccess   string
	lastRefresh  string
}

func (f *fakeSessionStore) CreateSession(_ context.Context, userID uuid.UUID, accessHash, refreshHash string, _ time.Time, _, _ *string) (*models.Session, error) {
	f.createdCalls++
	f.lastUserID = userID
	f.lastAccess = accessHash
	f.lastRefresh = refreshHash
	return &models.Session{UserID: userID, AccessTokenHash: accessHash, RefreshTokenHash: refreshHash}, nil
}

func (f *fakeSessionStore) FindSessionByAccessTokenHash(context.Context, string) (*models.Session, error) {
	panic("not used by LoginOpaque")
}
func (f *fakeSessionStore) RotateSession(context.Context, uuid.UUID, string, string, time.Time) error {
	panic("not used by LoginOpaque")
}
func (f *fakeSessionStore) RevokeSession(context.Context, uuid.UUID) error { panic("not used") }
func (f *fakeSessionStore) RevokeFamily(context.Context, uuid.UUID) error { panic("not used") }
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

func newLoginOpaqueHandler(authStore database.AuthStore, sessStore database.SessionStore) *Handlers {
	return &Handlers{
		AuthRepo:    authStore,
		SessionRepo: sessStore,
		Config:      config.Load(),
	}
}

// doLogin posts the supplied body to LoginOpaque and returns the recorder.
func doLogin(t *testing.T, h *Handlers, body any) *httptest.ResponseRecorder {
	t.Helper()
	buf, _ := json.Marshal(body)
	r := httptest.NewRequest(http.MethodPost, "/auth/v2/login", bytes.NewReader(buf))
	r.Header.Set("Content-Type", "application/json")
	r.RemoteAddr = "203.0.113.1:54321"
	rw := httptest.NewRecorder()
	h.LoginOpaque(rw, r)
	return rw
}

// TestLoginOpaque_EmptyBody returns 400 + VALIDATION_ERROR.
func TestLoginOpaque_EmptyBody(t *testing.T) {
	h := newLoginOpaqueHandler(&fakeAuthStore{}, &fakeSessionStore{})
	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/v2/login", nil)
	h.LoginOpaque(rw, r)
	require.Equal(t, http.StatusBadRequest, rw.Code)

	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeValidationError, out.Error.Code)
}

// TestLoginOpaque_MissingFields returns 400 + VALIDATION_ERROR with per-field map.
func TestLoginOpaque_MissingFields(t *testing.T) {
	h := newLoginOpaqueHandler(&fakeAuthStore{}, &fakeSessionStore{})
	rw := doLogin(t, h, map[string]string{"email": "x@x.com"})
	require.Equal(t, http.StatusBadRequest, rw.Code)

	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeValidationError, out.Error.Code)
	require.NotNil(t, out.Error.Fields)
	assert.Equal(t, "REQUIRED", out.Error.Fields["password"])
}

// TestLoginOpaque_UnknownUser returns 401 (anti-enumeration).
func TestLoginOpaque_UnknownUser(t *testing.T) {
	authStore := &fakeAuthStore{userByEmail: map[string]*models.User{}}
	sessStore := &fakeSessionStore{}
	h := newLoginOpaqueHandler(authStore, sessStore)

	rw := doLogin(t, h, map[string]string{"email": "nobody@here.test", "password": "Pa55word!"})
	require.Equal(t, http.StatusUnauthorized, rw.Code)

	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeInvalidCredentials, out.Error.Code)
	assert.Equal(t, 0, sessStore.createdCalls, "unknown user must not create a session")
}

// TestLoginOpaque_WrongPassword returns 401 + INVALID_CREDENTIALS and uses
// the same copy as the unknown-user path (anti-enumeration).
func TestLoginOpaque_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("right-password"), bcrypt.MinCost)
	u := &models.User{ID: uuid.New(), Email: "x@y.z", PasswordHash: string(hash)}

	authStore := &fakeAuthStore{userByEmail: map[string]*models.User{
		"x@y.z": u,
	}}
	sessStore := &fakeSessionStore{}
	h := newLoginOpaqueHandler(authStore, sessStore)

	rw := doLogin(t, h, map[string]string{"email": "x@y.z", "password": "wrong-password"})

	require.Equal(t, http.StatusUnauthorized, rw.Code)
	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeInvalidCredentials, out.Error.Code)
	assert.Equal(t, 0, sessStore.createdCalls)
}

// TestLoginOpaque_Success_ReturnsTokensAndSetsCookies walks the happy path.
// We don't seed the DB — we use the fakes — so this test doesn't depend
// on migration 017 being applied.
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

	authStore := &fakeAuthStore{userByEmail: map[string]*models.User{u.Email: u}}
	sessStore := &fakeSessionStore{}
	h := newLoginOpaqueHandler(authStore, sessStore)

	rw := doLogin(t, h, map[string]string{"email": u.Email, "password": "Pa55word!"})

	require.Equal(t, http.StatusOK, rw.Code, "body: %s", rw.Body.String())

	var out opaqueLoginResponse
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, "Bearer", out.TokenType)
	assert.Equal(t, accessTokenMaxAge, out.ExpiresInS)

	cookies := rw.Result().Cookies()
	require.Len(t, cookies, 2, "expected exactly 2 cookies (access + refresh)")
	gotAccess, gotRefresh := false, false
	var rawAccess, rawRefresh string
	for _, c := range cookies {
		switch c.Name {
		case "itinera_access":
			gotAccess = true
			rawAccess = c.Value
			assert.True(t, c.HttpOnly)
			assert.Len(t, c.Value, 64, "access token must be 64 hex chars (32 bytes)")
			assert.True(t, isLowerHex(c.Value), "access token must be lowercase hex")
		case "itinera_refresh":
			gotRefresh = true
			rawRefresh = c.Value
			assert.True(t, c.HttpOnly)
			assert.Len(t, c.Value, 64)
			assert.True(t, isLowerHex(c.Value))
		}
	}
	assert.True(t, gotAccess)
	assert.True(t, gotRefresh)

	assert.Equal(t, 1, sessStore.createdCalls)
	assert.Equal(t, uid, sessStore.lastUserID)
	assert.NotEmpty(t, sessStore.lastAccess)
	assert.NotEmpty(t, sessStore.lastRefresh)
	assert.NotEqual(t, sessStore.lastAccess, sessStore.lastRefresh,
		"access and refresh must be different secrets")

	assert.NotEqual(t, rawAccess, rawRefresh,
		"the cookies themselves must be different (raw tokens)")
}

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
