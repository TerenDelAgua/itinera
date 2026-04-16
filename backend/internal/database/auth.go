package database

import (
	"backend/internal/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (db *DB) CreateUser(ctx context.Context, email, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user models.User
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email
	`

	err = db.Pool.QueryRow(ctx, query, email, string(hash)).Scan(&user.ID, &user.Email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
	SELECT id, email, password_hash
	FROM users
	WHERE email = $1
	`
	err := db.Pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (db *DB) UpgradeTrips(ctx context.Context, sessionId string, userID uuid.UUID) error {
	query := `
		UPDATE trips
		SET user_id = $1, session_id = NULL
		WHERE session_id = $2 AND user_id IS NULL
	`
	_, err := db.Pool.Exec(ctx, query, userID, sessionId)

	return err
}
