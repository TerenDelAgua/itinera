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
	ForkTrip(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	DeleteTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) error
}

type TripContextStore interface {
	GetFork(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	GetTripMeta(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (isOwner bool, isDemo bool, err error)
	ForkTrip(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
}
