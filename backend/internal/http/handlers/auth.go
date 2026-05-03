package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{Token: token, User: *user})
}

func (h *Handlers) UpgradeSession(w http.ResponseWriter, r *http.Request) {
	userIdRaw := r.Context().Value(middleware.ContextKeyUserId{})
	userId, ok := userIdRaw.(uuid.UUID)

	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	sessionIdRaw := r.Context().Value(middleware.ContextKeySessionId{})
	sessionId, _ := sessionIdRaw.(string)

	if err := h.AuthRepo.UpgradeTrips(r.Context(), sessionId, userId); err != nil {
		http.Error(w, "Migration failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Guest trips migrated successfully"}`))
}
