package handlers

import (
	"backend/internal/auth"
	"backend/internal/http/middleware"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// opaque_auth_handlers.go implements the post-cutover auth endpoints
// (Spec 017 §5). They mirror the JWT-cookie handlers in auth.go but use
// opaque access + refresh cookies hashed in `sessions`. They coexist with
// the JWT tree during dual-stack; we'll retire auth.go's JWT handlers
// when AUTH_V2_ENABLED=true and these have proven stable in production.

// accessTokenMaxAge / refreshTokenMaxAge document the cookie lifetimes
// from Spec 017 §4.2. Centralised so we can tune both from one place.
const (
	accessTokenMaxAge  = 24 * 60 * 60      // 24 h
	refreshTokenMaxAge = 30 * 24 * 60 * 60 // 30 d
)

// LoginOpaque is the post-cutover login endpoint. Route name in the
// public API remains `/auth/v2/login` for compatibility with the spec
// table (Spec 017 §5.2) but the internal handler name reflects that it
// belongs to the opaque-token family, not to a "version 2".
//
// godoc
// @Summary      Login (post-cutover)
// @Description  Authenticate with email + password and issue HttpOnly
// @Description  access + refresh cookies. Returns the user payload.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      object  true  "Login data (email, password)"
// @Success      200   {object}  handlers.opaqueLoginResponse
// @Failure      400   {object}  handlers.JSONErrorBody "Validation error"
// @Failure      401   {object}  handlers.JSONErrorBody "Invalid credentials"
// @Failure      429   {object}  handlers.JSONErrorBody "Too many attempts"
// @Router       /auth/v2/login [post]
func (h *Handlers) LoginOpaque(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, CodeValidationError, "Invalid request body")
		return
	}
	if input.Email == "" || input.Password == "" {
		WriteErrorWithFields(w, http.StatusBadRequest, CodeValidationError,
			"Email and password are required",
			map[string]any{"email": "REQUIRED", "password": "REQUIRED"})
		return
	}

	user, err := h.AuthRepo.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials")
			return
		}
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Login lookup failed")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		// Anti-enumeration: same code + message on wrong password.
		WriteError(w, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials")
		return
	}

	rawAccess, err := auth.NewSecureToken()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Token generation failed")
		return
	}
	rawRefresh, err := auth.NewSecureToken()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Token generation failed")
		return
	}

	expiresAt := time.Now().Add(time.Duration(refreshTokenMaxAge) * time.Second)
	ua := r.UserAgent()
	userAgent := &ua
	ip := r.RemoteAddr
	ipAddress := &ip

	_, err = h.SessionRepo.CreateSession(r.Context(), user.ID,
		auth.HashToken(rawAccess), auth.HashToken(rawRefresh),
		expiresAt, userAgent, ipAddress,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Session creation failed")
		return
	}

	secure := h.Config.IsProduction() || r.Header.Get("X-Forwarded-Proto") == "https"
	middleware.SetAccessCookie(w, rawAccess, accessTokenMaxAge, secure)
	middleware.SetRefreshCookie(w, rawRefresh, refreshTokenMaxAge, secure)

	WriteJSON(w, http.StatusOK, opaqueLoginResponse{
		User:       user,
		ExpiresInS: accessTokenMaxAge,
		TokenType:  "Bearer",
	})
}

// LogoutOpaque godoc
// @Summary      Logout (post-cutover)
// @Description  Revoke the session identified by the access-token cookie
// @Description  and clear both auth cookies. Idempotent: returns 204 even
// @Description  if the cookie is missing or the session was already
// @Description  revoked.
// @Tags         auth
// @Produce      json
// @Success      204
// @Failure      500   {object}  handlers.JSONErrorBody "Revocation failed"
// @Router       /auth/v2/logout [post]
func (h *Handlers) LogoutOpaque(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.CookieAccessToken)
	if err == nil && cookie.Value != "" {
		hash := auth.HashToken(cookie.Value)
		if _, revokeErr := h.SessionRepo.RevokeSessionByAccessHash(r.Context(), hash); revokeErr != nil {
			WriteError(w, http.StatusInternalServerError, CodeInternalError, "Logout failed")
			return
		}
	}

	secure := h.Config.IsProduction() || r.Header.Get("X-Forwarded-Proto") == "https"
	middleware.ClearAuthCookies(w, secure)

	w.WriteHeader(http.StatusNoContent)
}

// MeOpaque godoc
// @Summary      Current user (post-cutover)
// @Description  Returns the user identified by the access-token cookie.
// @Description  The frontend hits this on every page load to validate the
// @Description  session and pull tier / locale / display fields.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  models.User
// @Failure      401  {object}  handlers.JSONErrorBody "No session"
// @Failure      403  {object}  handlers.JSONErrorBody "Soft-deleted account"
// @Router       /auth/v2/me [get]
func (h *Handlers) MeOpaque(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusUnauthorized, CodeUnauthenticated, "No active session")
		return
	}

	user, err := h.AuthRepo.GetUserByID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, CodeUnauthenticated, "User not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Lookup failed")
		return
	}

	if user.DeletedAt != nil {
		WriteError(w, http.StatusForbidden, CodeAccountDeleted, "Account has been deleted")
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

// RefreshOpaque issues a new access + refresh pair in exchange for the
// refresh cookie (Spec 017 §5.4 / §4.3 sliding session).
//
// Flow:
//  1. Read the itinera_refresh cookie and hash it.
//  2. Look the row up; if revoked_at is set → REUSE DETECTED: revoke the
//     entire family and return TOKEN_REUSE_DETECTED. This is the safety
//     mechanism against a stolen refresh cookie that a thief re-plays:
//     the legit user already used that row once, rotated to a new family,
//     so the cookie's hash now points at a row whose revoked_at is set.
//  3. If the row is fresh and not expired → mint new raw tokens, rotate
//     the row to a new family, set fresh cookies, return 200 with the
//     same shape as /auth/v2/login.
//
// @Summary      Refresh access + refresh tokens
// @Description  Exchanges a refresh cookie for a new access + refresh pair.
// @Description  Detects and revokes a token family on reuse.
// @Tags         auth
// @Produce      json
// @Success      200   {object}  handlers.opaqueLoginResponse
// @Failure      401   {object}  handlers.JSONErrorBody "No refresh cookie"
// @Failure      403   {object}  handlers.JSONErrorBody "Reuse detected"
// @Router       /auth/v2/refresh [post]
func (h *Handlers) RefreshOpaque(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.CookieRefreshToken)
	if err != nil || cookie.Value == "" {
		WriteError(w, http.StatusUnauthorized, CodeUnauthenticated, "No refresh cookie")
		return
	}

	rawRefresh := cookie.Value
	session, err := h.SessionRepo.FindSessionByRefreshTokenHash(r.Context(), auth.HashToken(rawRefresh))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, CodeUnauthenticated, "Refresh not recognised")
			return
		}
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Refresh lookup failed")
		return
	}

	// Reuse detection: row exists but was already revoked. This is the
	// canonical case of a stolen refresh cookie being replayed; the legit
	// user already rotated past it. Kill the family so the legitimate
	// owner's new session also dies, forcing a full re-login.
	if session.RevokedAt != nil {
		_ = h.SessionRepo.RevokeFamily(r.Context(), session.RefreshFamily)
		WriteError(w, http.StatusForbidden, CodeReuseDetected, "Refresh token reused; full session revoked")
		return
	}

	// Reject if the refresh row has aged past its absolute window. The
	// grace period the spec mentions is 30s, but only relevant to the
	// access leg; the refresh leg has its own expires_at.
	if expired, _ := sessionExpired(session); expired {
		WriteError(w, http.StatusUnauthorized, CodeSessionExpired, "Refresh token expired")
		return
	}

	// Mint fresh tokens and rotate the row to a new family.
	rawAccess, err := auth.NewSecureToken()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Token generation failed")
		return
	}
	rawRefreshNew, err := auth.NewSecureToken()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Token generation failed")
		return
	}
	newExpiry := time.Now().Add(time.Duration(refreshTokenMaxAge) * time.Second)

	if err := h.SessionRepo.RotateSession(r.Context(), session.ID,
		auth.HashToken(rawAccess), auth.HashToken(rawRefreshNew), newExpiry,
	); err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Session rotation failed")
		return
	}

	secure := h.Config.IsProduction() || r.Header.Get("X-Forwarded-Proto") == "https"
	middleware.SetAccessCookie(w, rawAccess, accessTokenMaxAge, secure)
	middleware.SetRefreshCookie(w, rawRefreshNew, refreshTokenMaxAge, secure)

	WriteJSON(w, http.StatusOK, opaqueLoginResponse{
		User:       nil,
		ExpiresInS: accessTokenMaxAge,
		TokenType:  "Bearer",
	})
}

// sessionExpired compares a parsed ExpiresAt string (ISO-8601) against
// now. It's a tiny helper rather than scanning RFC3339 inline in the
// handler — keeps the handler readable and the parsing rule in one place.
func sessionExpired(s *models.Session) (bool, error) {
	if s == nil || s.ExpiresAt == "" {
		return true, nil
	}
	t, err := time.Parse(time.RFC3339, s.ExpiresAt)
	if err != nil {
		return true, err
	}
	return t.Before(time.Now()), nil
}

// opaqueLoginResponse is the body of POST /auth/v2/login. `expires_in` and
// `token_type` mirror the OAuth2 token endpoint shape so the frontend can
// share code with existing OAuth client libraries if needed later.
type opaqueLoginResponse struct {
	User       any    `json:"user"`
	TokenType  string `json:"token_type"`
	ExpiresInS int    `json:"expires_in"`
}
