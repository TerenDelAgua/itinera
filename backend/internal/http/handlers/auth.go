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
// @Failure      400   {string}  string "Invalid request body"
// @Failure      409   {string}  string "Email already exists"
// @Router       /auth/register [post]
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.AuthRepo.CreateUser(r.Context(), input.Email, input.Password)
	if err != nil {
		http.Error(w, "Email already exists or DB error", http.StatusConflict)
		return
	}

	token, _ := h.generateToken(*user)
	h.setAuthCookie(w, r, token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{Token: token, User: *user})
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      object  true  "Login data (email, password)"
// @Success      200   {object}  handlers.TokenResponse
// @Failure      401   {string}  string "Invalid credentials"
// @Router       /auth/login [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.AuthRepo.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := h.generateToken(*user)
	h.setAuthCookie(w, r, token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{Token: token, User: *user})
}

func (h *Handlers) ClaimGuest(w http.ResponseWriter, r *http.Request) {
	userIdRaw := r.Context().Value(middleware.ContextKeyUserId{})
	userId, ok := userIdRaw.(uuid.UUID)

	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	sessionIdRaw := r.Context().Value(middleware.ContextKeySessionId{})
	sessionId, _ := sessionIdRaw.(string)

	claimed, err := h.AuthRepo.ClaimGuestTrips(r.Context(), sessionId, userId)
	if err != nil {
		http.Error(w, "Migration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"claimed_trips_count": claimed,
		"message":             fmt.Sprintf("Hemos añadido %d viajes a tu cuenta", claimed),
	})
}
