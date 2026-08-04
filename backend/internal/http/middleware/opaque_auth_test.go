package middleware

import (
	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSessionStore implements database.SessionStore with a hand-rolled map
// keyed on access hash. It's deliberately minimal: only the methods the
// middleware calls are non-trivial; the rest panic on call so a regression
// in the middleware (e.g. accidentally hitting CreateSession) is loud.
type fakeSessionStore struct {
	hashes map[string]*models.Session
	dbErr  error
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{hashes: make(map[string]*models.Session)}
}

func (f *fakeSessionStore) FindSessionByAccessTokenHash(ctx context.Context, accessHash string) (*models.Session, error) {
	if f.dbErr != nil {
		return nil, f.dbErr
	}
	s, ok := f.hashes[accessHash]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return s, nil
}

// All other methods of the interface are unimplemented on purpose.

func (f *fakeSessionStore) CreateSession(context.Context, uuid.UUID, string, string, time.Time, *string, *string) (*models.Session, error) {
	panic("CreateSession should not be called from AuthMiddlewareV2")
}
func (f *fakeSessionStore) RotateSession(context.Context, uuid.UUID, string, string, time.Time) error {
	panic("RotateSession should not be called from AuthMiddlewareV2")
}
func (f *fakeSessionStore) RevokeSession(context.Context, uuid.UUID) error {
	panic("RevokeSession should not be called from AuthMiddlewareV2")
}
func (f *fakeSessionStore) RevokeSessionByAccessHash(context.Context, string) (int, error) {
	panic("RevokeSessionByAccessHash should not be called from AuthMiddlewareV2")
}
func (f *fakeSessionStore) RevokeFamily(context.Context, uuid.UUID) error {
	panic("RevokeFamily should not be called from AuthMiddlewareV2")
}
func (f *fakeSessionStore) RevokeAllSessionsForUser(context.Context, uuid.UUID) error {
	panic("RevokeAllSessionsForUser should not be called from AuthMiddlewareV2")
}
func (f *fakeSessionStore) CountActiveSessionsForUser(context.Context, uuid.UUID) (int, error) {
	panic("CountActiveSessionsForUser should not be called from AuthMiddlewareV2")
}
func (f *fakeSessionStore) CleanupExpiredSessions(context.Context) (int64, error) {
	panic("CleanupExpiredSessions should not be called from AuthMiddlewareV2")
}

var _ database.SessionStore = (*fakeSessionStore)(nil)

// runWithCookie is a helper that wires AuthMiddlewareV2 around a no-op
// next handler and records what came through the context.
func runWithCookie(t *testing.T, store database.SessionStore, cookie *http.Cookie) (userID uuid.UUID, refreshFamily string, ok bool, status int) {
	t.Helper()
	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}

	var seenUID uuid.UUID
	var seenFamily string
	var seenOK bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, ok := r.Context().Value(ContextKeyUserId{}).(uuid.UUID); ok {
			seenUID = uid
			seenOK = true
		}
		if fam, ok := r.Context().Value(ContextKeySessionId{}).(string); ok {
			seenFamily = fam
		}
		w.WriteHeader(http.StatusOK)
	})

	AuthMiddlewareV2(store)(next).ServeHTTP(rw, r)
	return seenUID, seenFamily, seenOK, rw.Code
}

// TestOpaqueAuth_NoCookie_Anonymous: no cookie → next runs, no context.
func TestOpaqueAuth_NoCookie_Anonymous(t *testing.T) {
	store := newFakeSessionStore()
	_, _, ok, status := runWithCookie(t, store, nil)
	require.False(t, ok, "no user id should be in context")
	assert.Equal(t, http.StatusOK, status)
}

// TestOpaqueAuth_ValidCookie_DecoratesContext: cookie present and row exists
// → ContextKeyUserId and ContextKeySessionId both populated.
func TestOpaqueAuth_ValidCookie_DecoratesContext(t *testing.T) {
	store := newFakeSessionStore()
	raw := "raw-access-token-abc123"
	hash := auth.HashToken(raw)

	uid := uuid.New()
	fam := uuid.New()
	store.hashes[hash] = &models.Session{
		ID:              uuid.New(),
		UserID:          uid,
		AccessTokenHash: hash,
		RefreshFamily:   fam,
	}

	seenUID, seenFamily, ok, _ := runWithCookie(t, store, &http.Cookie{Name: CookieAccessToken, Value: raw})

	require.True(t, ok)
	assert.Equal(t, uid, seenUID)
	assert.Equal(t, fam.String(), seenFamily, "refresh family must round-trip as the canonical UUID string")
}

// TestOpaqueAuth_UnknownCookie_Anonymous: cookie present but not in store →
// next runs anonymously (no 401).
func TestOpaqueAuth_UnknownCookie_Anonymous(t *testing.T) {
	store := newFakeSessionStore()
	_, _, ok, status := runWithCookie(t, store, &http.Cookie{Name: CookieAccessToken, Value: "never-issued"})
	require.False(t, ok)
	assert.Equal(t, http.StatusOK, status, "unknown tokens must NOT 401; they fall back to anonymous")
}

// TestOpaqueAuth_DBError_500: store returns a non-ErrNoRows error → 500.
func TestOpaqueAuth_DBError_500(t *testing.T) {
	store := newFakeSessionStore()
	store.dbErr = errors.New("connection reset by peer")
	_, _, ok, status := runWithCookie(t, store, &http.Cookie{Name: CookieAccessToken, Value: "any-token"})
	require.False(t, ok)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.False(t, ok, "user must NOT be authenticated when the lookup fails")
}

// TestOpaqueAuth_OptionsBypass: CORS preflight passes through without cookie.
func TestOpaqueAuth_OptionsBypass(t *testing.T) {
	store := newFakeSessionStore()
	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/", nil)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	AuthMiddlewareV2(store)(next).ServeHTTP(rw, r)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rw.Code)
}

// TestSetClearCookies_Shape: cookie helpers set the flags Spec 017 §7.7
// requires (Path=/, HttpOnly, SameSite=Lax by default).
func TestSetClearCookies_Shape(t *testing.T) {
	rw := httptest.NewRecorder()

	SetAccessCookie(rw, "abc", 60, false)
	SetRefreshCookie(rw, "def", 3600, false)
	ClearAuthCookies(rw, false)

	cookies := rw.Result().Cookies()
	assert.Len(t, cookies, 4, "SetAccess + SetRefresh + ClearAccess + ClearRefresh")

	for _, c := range cookies {
		assert.Equal(t, "/", c.Path, "Path must be / for cookie to be sent on every endpoint")
		assert.True(t, c.HttpOnly, "auth cookies must be HttpOnly")
		assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	}
}
