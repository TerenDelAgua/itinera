package services

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripService struct {
	Pool           *pgxpool.Pool
	TripsRepo      database.TripStore
	PlacesRepo     database.PlaceStore
	ActivitiesRepo database.ActivityStore
	ExpensesRepo   database.ExpenseStore
	EventsRepo     database.EventStore
}

func NewTripService(
	pool *pgxpool.Pool,
	tripsRepo database.TripStore,
	placesRepo database.PlaceStore,
	activitiesRepo database.ActivityStore,
	expensesRepo database.ExpenseStore,
	eventsRepo database.EventStore,
) *TripService {
	return &TripService{
		Pool:           pool,
		TripsRepo:      tripsRepo,
		PlacesRepo:     placesRepo,
		ActivitiesRepo: activitiesRepo,
		ExpensesRepo:   expensesRepo,
		EventsRepo:     eventsRepo,
	}

}

// `trigger` loggued in events.metadata.trigger:
//   - "ghost_mode_write" for public demos forks
//   - "share_forked" for shared trips forks
func (s *TripService) CloneTrip(
	ctx context.Context,
	origID uuid.UUID,
	userID *uuid.UUID,
	sessionID *string,
	isInternal bool,
	trigger string,
) (*models.Trip, error) {
	if userID == nil && sessionID == nil {
		return nil, fmt.Errorf("identity required to clone trip")
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	origTrip, err := s.TripsRepo.GetBaseForClone(ctx, tx, origID)
	if err != nil {
		return nil, fmt.Errorf("failed to load original trip: %w", err)
	}

	newTripID := uuid.New()
	if err := s.TripsRepo.InsertFork(
		ctx, tx, newTripID, userID, sessionID, isInternal, origID, origTrip,
	); err != nil {
		return nil, fmt.Errorf("failed to insert forked trip: %w", err)
	}

	placeMap, err := s.PlacesRepo.CloneByTripID(ctx, tx, origID, newTripID)
	if err != nil {
		return nil, fmt.Errorf("failed to clone places: %w", err)
	}

	if err := s.ActivitiesRepo.CloneByTripID(
		ctx, tx, origID, newTripID, placeMap,
	); err != nil {
		return nil, fmt.Errorf("failed to clone activities: %w", err)
	}

	if err := s.ExpensesRepo.CloneByTripID(
		ctx, tx, origID, newTripID, userID, placeMap,
	); err != nil {
		return nil, fmt.Errorf("failed to clone expenses: %w", err)
	}

	metadata := fmt.Sprintf(`{"forked_from":"%s","trigger":"%s"}`, origID, trigger)

	// Commit BEFORE logging the event. The events_trip_id_fkey foreign key
	// would otherwise reject the insert: the trip row exists inside this
	// transaction, but a separate connection from the pool (which is what
	// EventRepository.Insert uses) cannot see uncommitted inserts.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit clone transaction: %w", err)
	}

	// Fire-and-forget analytics: best-effort, after the trip is durable.
	var sessIDStr string
	if sessionID != nil {
		sessIDStr = *sessionID
	}
	var userIDStr, tripIDStr *string
	if userID != nil {
		s := userID.String()
		userIDStr = &s
	}
	t := newTripID.String()
	tripIDStr = &t
	if err := s.EventsRepo.Insert(
		ctx, "trip_forked", sessIDStr, userIDStr, tripIDStr,
		[]byte(metadata), time.Now(),
	); err != nil {
		// Analytics must never block the user-visible flow. Log and continue.
		fmt.Printf("warn: fork event log failed: %v\n", err)
	}

	return s.TripsRepo.GetTripById(ctx, newTripID)

}

func (s *TripService) ForkFromDemo(
	ctx context.Context,
	demoTripID uuid.UUID,
	userID *uuid.UUID,
	sessionID *string,
) (*models.Trip, error) {
	isDemo, err := s.TripsRepo.IsPublicDemo(ctx, demoTripID)
	if err != nil {
		return nil, err
	}

	if !isDemo {
		return nil, fmt.Errorf("trip %s is not a public demo", demoTripID)
	}
	return s.CloneTrip(ctx, demoTripID, userID, sessionID, false, "ghost_mode_write")
}

func (s *TripService) ForkFromShareToken(
	ctx context.Context,
	token string,
	userID *uuid.UUID,
	sessionID *string,
	isInternal bool,
) (*models.Trip, error) {
	origID, err := s.TripsRepo.GetActiveShareTripID(ctx, token)
	if err != nil {
		return nil, err //pgx.ErrNoRows if not exists /expired / disabled
	}
	return s.CloneTrip(ctx, origID, userID, sessionID, isInternal, "share_forked")
}

func (s *TripService) GetFork(
	ctx context.Context,
	forkedFrom string,
	userID *uuid.UUID,
	sessionID *string,
) (*models.Trip, error) {
	return s.TripsRepo.GetFork(ctx, forkedFrom, userID, sessionID)
}

func (s *TripService) GetTripMeta(
	ctx context.Context,
	tripID string,
	userID *uuid.UUID,
	sessionID *string,
) (bool, bool, error) {
	return s.TripsRepo.GetTripMeta(ctx, tripID, userID, sessionID)
}
