package services

import (
	"context"

	"github.com/google/uuid"
	"backend/internal/database"
	"backend/internal/models"
)

type TripService struct {
	store database.TripStore
}

func NewTripService(store database.TripStore) *TripService {
	return &TripService{store: store}
}

func (s *TripService) CreateTrip(ctx context.Context, userId *uuid.UUID, sessionId *string, tripData models.Trip) (*models.Trip, error) {
	return s.store.CreateTrip(ctx, userId, sessionId, tripData)
}

func (s *TripService) ListTrips(ctx context.Context, userId *uuid.UUID, sessionId *string) ([]models.Trip, error) {
	return s.store.ListTrips(ctx, userId, sessionId)
}

func (s *TripService) GetTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) (*models.Trip, error) {
	return s.store.GetTrip(ctx, id, userId, sessionId)
}

func (s *TripService) UpdateTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string, updates map[string]any) (*models.Trip, error) {
	return s.store.UpdateTrip(ctx, id, userId, sessionId, updates)
}

func (s *TripService) DeleteTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) error {
	return s.store.DeleteTrip(ctx, id, userId, sessionId)
}

func (s *TripService) ListPublicDemos(ctx context.Context, limit int) ([]models.Trip, error) {
	return s.store.ListPublicDemos(ctx, limit)
}

func (s *TripService) GetPublicStats(ctx context.Context) (int, error) {
	return s.store.GetPublicStats(ctx)
}
