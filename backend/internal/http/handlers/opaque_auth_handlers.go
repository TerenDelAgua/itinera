package handlers

import (
	"backend/internal/auth"
	"backend/internal/http/middleware"
	"backend/internal/models"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// opaque_auth_handlers.go implements the post-cutover auth endpoints
// They mirror the JWT-cookie handlers in auth.go but use
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
		// Anti-enumeration: same code + message on wrong password so the
		// caller can't tell "user doesn't exist" from "user exists,
		// wrong password".
		WriteError(w, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials")
		return
	}

	// Generate the access + refresh tokens. The server stores ONLY the
	// SHA-256 hashes; the raw values ride HttpOnly cookies.
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

// RegisterOpaque is the post-cutover signup endpoint. It mirrors the JWT
// path (handlers/auth.go: Register) but ends by minting opaque tokens and
// setting the HttpOnly cookies, so the user is signed in on return.
//
// @Summary      Register a new user (post-cutover)
// @Description  Creates a new account and signs the user in.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      object  true  "Registration data (email, password, locale)"
// @Success      200   {object}  handlers.opaqueLoginResponse
// @Failure      400   {object}  handlers.JSONErrorBody "Validation error"
// @Failure      409   {object}  handlers.JSONErrorBody "Email already exists"
// @Router       /auth/v2/register [post]
func (h *Handlers) RegisterOpaque(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Locale   string `json:"locale"`
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
	if !isStrongEnoughPassword(input.Password) {
		WriteErrorWithFields(w, http.StatusBadRequest, CodeWeakPassword,
			"Password must be at least 8 characters",
			map[string]any{"password": "TOO_SHORT"})
		return
	}
	if input.Locale == "" {
		input.Locale = "en"
	}

	user, err := h.AuthRepo.CreateUser(r.Context(), input.Email, input.Password, input.Locale)
	if err != nil {
		WriteError(w, http.StatusConflict, CodeEmailAlreadyExists, "An account with this email already exists")
		return
	}

	// Auto-login: mint tokens, create session, set cookies. We deliberately
	// share the path with LoginOpaque so future changes (extra claims,
	// analytics events) only touch one place.
	if err := h.activateSession(w, r, *user); err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Session creation failed")
		return
	}

	// Welcome email fires AFTER the session is set so a backend hiccup
	// doesn't cost the user their signup.
	if h.EmailSender != nil {
		_ = h.EmailSender.SendWelcome(r.Context(), *user, user.Locale)
	}

	WriteJSON(w, http.StatusOK, opaqueLoginResponse{
		User:       user,
		ExpiresInS: accessTokenMaxAge,
		TokenType:  "Bearer",
	})
}

// ForgotOpaque starts the password reset flow. Anti-enumeration + IP rate
// limiting apply; the same copy is returned whether the email exists or
// not (Spec 017 §5.5).
//
// @Summary      Request a password reset code
// @Description  Sends a 6-digit code to the email if the account exists.
// @Description  Same response shape regardless of whether the email is
// @Description  registered — anti-enumeration.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      object  true  "{email}"
// @Success      202   {object}  handlers.forgotResponse
// @Failure      400   {object}  handlers.JSONErrorBody "Validation error"
// @Failure      429   {object}  handlers.JSONErrorBody "Too many attempts"
// @Router       /auth/v2/forgot [post]
func (h *Handlers) ForgotOpaque(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email  string `json:"email"`
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, CodeValidationError, "Invalid request body")
		return
	}
	if input.Email == "" {
		WriteErrorWithFields(w, http.StatusBadRequest, CodeValidationError,
			"Email is required", map[string]any{"email": "REQUIRED"})
		return
	}
	if input.Locale == "" {
		input.Locale = "en"
	}

	// Rate-limit by IP BEFORE we touch the user table so an attacker
	// can't probe emails through the reset endpoint.
	if h.LoginRateLimitRepo != nil {
		ip := clientIP(r)
		_, blocked, err := h.LoginRateLimitRepo.RecordFailure(r.Context(), ip)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, CodeInternalError, "Rate limit lookup failed")
			return
		}
		if blocked {
			WriteError(w, http.StatusTooManyRequests, CodeRateLimited,
				"Too many reset attempts; try again later")
			return
		}
	}

	user, err := h.AuthRepo.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		// Either user doesn't exist OR user is soft-deleted. We mirror the
		// success response to avoid leaking which it is.
		WriteJSON(w, http.StatusAccepted, forgotResponse{Message: forgotAck})
		return
	}
	if user.DeletedAt != nil {
		WriteJSON(w, http.StatusAccepted, forgotResponse{Message: forgotAck})
		return
	}

	code, err := generateSixDigitCode()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Code generation failed")
		return
	}

	// Invalidate any previous active code, then write the new one. The
	// unique constraint on token_hash + DB-side `MarkPreviousAsUsed`
	// guarantees a single live code per user.
	if err := h.ResetRepo.MarkPreviousAsUsed(r.Context(), user.ID); err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Reset code invalidation failed")
		return
	}
	hash := auth.HashToken(code)
	ipAddr := r.RemoteAddr
	if err := h.ResetRepo.Create(r.Context(), user.ID, hash, time.Now().Add(1*time.Hour), &ipAddr); err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Reset code persistence failed")
		return
	}

	if h.EmailSender != nil {
		_ = h.EmailSender.SendPasswordReset(r.Context(), *user, code, input.Locale)
	}

	WriteJSON(w, http.StatusAccepted, forgotResponse{Message: forgotAck})
}

// ResetOpaque consumes the 6-digit code, updates the password, and
// revokes every active session so a stolen cookie cannot survive a reset.
//
// @Summary      Reset password with the 6-digit code
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      object  true  "{email, code, new_password}"
// @Success      204
// @Failure      400   {object}  handlers.JSONErrorBody "Validation error"
// @Failure      401   {object}  handlers.JSONErrorBody "Invalid / expired / locked code"
// @Router       /auth/v2/reset [post]
func (h *Handlers) ResetOpaque(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, CodeValidationError, "Invalid request body")
		return
	}
	if input.Email == "" || input.Code == "" || input.NewPassword == "" {
		WriteError(w, http.StatusBadRequest, CodeValidationError, "All fields are required")
		return
	}
	if !isStrongEnoughPassword(input.NewPassword) {
		WriteErrorWithFields(w, http.StatusBadRequest, CodeWeakPassword,
			"Password must be at least 8 characters",
			map[string]any{"new_password": "TOO_SHORT"})
		return
	}

	// Lookup the user first; the reset row is keyed by user_id so the
	// hash alone wouldn't disambiguate which account to update.
	user, err := h.AuthRepo.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, CodeInvalidToken, "Invalid reset code")
		return
	}

	codeHash := auth.HashToken(input.Code)
	token, err := h.ResetRepo.FindActiveByHash(r.Context(), codeHash)
	if err != nil || token.UserID != user.ID {
		WriteError(w, http.StatusUnauthorized, CodeInvalidToken, "Invalid reset code")
		return
	}

	attempts, locked, err := h.ResetRepo.RecordFailedAttempt(r.Context(), token.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Reset attempt tracking failed")
		return
	}
	if locked {
		WriteError(w, http.StatusUnauthorized, CodeLockedToken, "Too many invalid attempts; request a new code")
		return
	}

	if err := h.AuthRepo.UpdateUserPassword(r.Context(), user.ID, input.NewPassword); err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Password update failed")
		return
	}
	if err := h.ResetRepo.MarkUsed(r.Context(), token.ID); err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Reset code closure failed")
		return
	}
	// Revoke every active session for the user. RevokeAllSessionsForUser
	// returns an error only if the DB is unhealthy, so we surface it
	// directly.
	if err := h.SessionRepo.RevokeAllSessionsForUser(r.Context(), user.ID); err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Session revocation failed")
		return
	}

	// `attempts` is recorded so future log/audit endpoints can read the
	// row count. We intentionally ignore it on the happy path: the
	// caller just wants 204.
	_ = attempts

	secure := h.Config.IsProduction() || r.Header.Get("X-Forwarded-Proto") == "https"
	middleware.ClearAuthCookies(w, secure)

	w.WriteHeader(http.StatusNoContent)
}

// ---------- Helpers shared by the v2 tree -------------------------------

// activateSession is the login/register common path. It mints tokens,
// creates the session row, and sets the cookies. Returns the error so the
// caller can decide the HTTP shape.
func (h *Handlers) activateSession(w http.ResponseWriter, r *http.Request, user models.User) error {
	rawAccess, err := auth.NewSecureToken()
	if err != nil {
		return err
	}
	rawRefresh, err := auth.NewSecureToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(time.Duration(refreshTokenMaxAge) * time.Second)
	ua := r.UserAgent()
	ip := r.RemoteAddr
	_, err = h.SessionRepo.CreateSession(r.Context(), user.ID,
		auth.HashToken(rawAccess), auth.HashToken(rawRefresh),
		expiresAt, &ua, &ip,
	)
	if err != nil {
		return err
	}

	secure := h.Config.IsProduction() || r.Header.Get("X-Forwarded-Proto") == "https"
	middleware.SetAccessCookie(w, rawAccess, accessTokenMaxAge, secure)
	middleware.SetRefreshCookie(w, rawRefresh, refreshTokenMaxAge, secure)
	return nil
}

// generateSixDigitCode returns a cryptographically random 6-digit decimal
// string suitable for the password-reset email. 10^6 combinations is
// 1M; combined with 5 attempts per code and 1h expiry, the per-code
// brute-force success probability is 5/1e6 = 0.0005% (Spec 017 §5.7).
func generateSixDigitCode() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// Mask to 20 bits — fits within "000000"..999999 without bias from
	// the 12 bits we drop on the left of the uint32.
	n := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1000000
	return fmt.Sprintf("%06d", n), nil
}

// isStrongEnoughPassword is the Spec 017 §5.1 password check: minimum
// 8 characters. We deliberately don't add a complexity rule here —
// length > complexity. Future enhancement: zxcvbn.
func isStrongEnoughPassword(p string) bool {
	return len(p) >= 8
}

// clientIP extracts a stable IP from the request, preferring
// X-Forwarded-For when the proxy is trusted. We DON'T trust it from the
// client (any browser can set the header); for now this is best-effort.
// A future iteration can plumb a TrustedProxies config.
func clientIP(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		if i := strings.IndexByte(h, ','); i >= 0 {
			return strings.TrimSpace(h[:i])
		}
		return strings.TrimSpace(h)
	}
	return r.RemoteAddr
}

// forgotResponse is the single response shape /auth/v2/forgot returns
// regardless of whether the email is registered. The Message is localisable
// in a future PR (Spec 017 §8.4).
type forgotResponse struct {
	Message string `json:"message"`
}

const forgotAck = "If the email exists, a code has been sent."
