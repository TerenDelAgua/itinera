package handlers

import (
	"backend/internal/auth"
	"backend/internal/http/middleware"
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

// opaqueLoginResponse is the body of POST /auth/v2/login. `expires_in` and
// `token_type` mirror the OAuth2 token endpoint shape so the frontend can
// share code with existing OAuth client libraries if needed later.
type opaqueLoginResponse struct {
	User       any    `json:"user"`
	TokenType  string `json:"token_type"`
	ExpiresInS int    `json:"expires_in"`
}
