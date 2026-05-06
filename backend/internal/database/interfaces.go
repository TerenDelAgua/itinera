package database

import (
	"context"

	"github.com/google/uuid"
	"backend/internal/models"
)

type TripStore interface {
	CreateTrip(ctx context.Context, userId *uuid.UUID, sessionId *string, tripData models.Trip) (*models.Trip, error)
	ListTrips(ctx context.Context, userId *uuid.UUID, sessionId *string) ([]models.Trip, error)
	GetTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) (*models.Trip, error)
	GetTripById(ctx context.Context, id uuid.UUID) (*models.Trip, error)
	UpdateTrip(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string, updates map[string]any) (*models.Trip, error)
	DeleteTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) error
}
