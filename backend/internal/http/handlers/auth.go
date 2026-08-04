package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const authCookieName = "auth_token"
const authCookieMaxAge = 72 * 60 * 60 // matches JWT exp (72h)

type TokenResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *Handlers) generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.Config.JWTSecret))
}

// setAuthCookie writes the JWT as an HttpOnly cookie so the browser sends it
// automatically with every request via credentials: 'include'. This removes
// the need to store the token in localStorage (XSS-readable) on the client.
// SameSite is set to None in secure contexts (production / TLS-terminated
// proxies) so the cookie crosses the Vercel <-> Railway origin boundary,
// and Lax otherwise for local development.
func (h *Handlers) setAuthCookie(w http.ResponseWriter, r *http.Request, token string) {
	isSecureContext := h.Config.IsProduction() || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if isSecureContext {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   authCookieMaxAge,
		HttpOnly: true,
		Secure:   isSecureContext,
		SameSite: sameSite,
	})
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new account and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      object  true  "Registration data (email, password)"
// @Success      200   {object}  handlers.TokenResponse
// @Failure      400   {object}  handlers.JSONErrorBody "Invalid request body"
// @Failure      409   {object}  handlers.JSONErrorBody "Email already exists"
// @Router       /auth/register [post]
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
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
			map[string]any{
				"email":    "REQUIRED",
				"password": "REQUIRED",
			})
		return
	}
	if input.Locale == "" {
		input.Locale = "en"
	}

	user, err := h.AuthRepo.CreateUser(r.Context(), input.Email, input.Password, input.Locale)
	if err != nil {
		// The partial unique index is case-insensitive, so
		// INSERT collisions go to a 409. The body NEVER echoes err.Error().
		WriteError(w, http.StatusConflict, CodeEmailAlreadyExists, "An account with this email already exists")
		return
	}

	token, _ := h.generateToken(*user)
	h.setAuthCookie(w, r, token)
	WriteJSON(w, http.StatusOK, TokenResponse{Token: token, User: *user})
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      object  true  "Login data (email, password)"
// @Success      200   {object}  handlers.TokenResponse
// @Failure      401   {object}  handlers.JSONErrorBody "Invalid credentials"
// @Router       /auth/login [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, CodeValidationError, "Invalid request body")
		return
	}

	user, err := h.AuthRepo.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		// Anti-enumeration: same response if the email doesn't exist OR if
		// the password is wrong. The frontend uses the same "invalid
		// credentials" copy so an attacker can't enumerate.
		WriteError(w, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		WriteError(w, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials")
		return
	}

	token, _ := h.generateToken(*user)
	h.setAuthCookie(w, r, token)
	WriteJSON(w, http.StatusOK, TokenResponse{Token: token, User: *user})
}

func (h *Handlers) ClaimGuest(w http.ResponseWriter, r *http.Request) {
	userIdRaw := r.Context().Value(middleware.ContextKeyUserId{})
	userId, ok := userIdRaw.(uuid.UUID)

	if !ok {
		WriteError(w, http.StatusUnauthorized, CodeUnauthenticated, "Authentication required")
		return
	}

	sessionIdRaw := r.Context().Value(middleware.ContextKeySessionId{})
	sessionId, _ := sessionIdRaw.(string)

	claimed, err := h.AuthRepo.ClaimGuestTrips(r.Context(), sessionId, userId)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, CodeInternalError, "Migration failed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"claimed_trips_count": claimed,
		"message":             fmt.Sprintf("Hemos añadido %d viajes a tu cuenta", claimed),
	})
}
