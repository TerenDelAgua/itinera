package database

// This file declares the storage interfaces consumed by the HTTP layer and
// services. Concrete implementations (TripRepository, PlaceRepository, etc.)
// satisfy these interfaces structurally — Go's type system checks satisfaction
// at compile time wherever the interface is used as a parameter or field
// type, so we do not need an `implements` declaration.
//
// The interfaces deliberately expose only the methods called from outside the
// `database` package. Keeping them narrow:
//
//   - reduces the surface area that needs to be mocked in unit tests;
//   - forces business logic that does not belong in a repository (e.g. cross-
//     entity operations) to live in the services or a Unit-of-Work instead.

import (
	"context"
	"time"

	"backend/internal/models"

	"github.com/google/uuid"
)

// ── Trips ────────────────────────────────────────────────────────────────────

// TripStore is the read/write contract for trips.
type TripStore interface {
	CreateTrip(ctx context.Context, userID *uuid.UUID, sessionID *string, tripData models.Trip) (*models.Trip, error)
	GetTrip(ctx context.Context, id string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	GetTripById(ctx context.Context, id uuid.UUID) (*models.Trip, error)
	ListTrips(ctx context.Context, userID *uuid.UUID, sessionID *string) ([]models.Trip, error)
	UpdateTrip(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string, updates map[string]any) (*models.Trip, error)
	DeleteTrip(ctx context.Context, id string, userID *uuid.UUID, sessionID *string) error

	// ForkTrip clones a trip (and its dependent rows) within a single
	// transaction; both source and destination share the same user/session
	// ownership context.
	ForkTrip(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)

	// GetFork returns the caller's existing fork of a public demo, if any.
	// Used by the trip-context middleware to detect ghost-mode edits.
	GetFork(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)

	// GetTripMeta returns (isOwner, isPublicDemo, err) used by the events
	// handler to scope analytics without exposing the full repository.
	GetTripMeta(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error)

	// Public demo content for the landing page.
	ListPublicDemos(ctx context.Context, limit int) ([]models.Trip, error)
	GetPublicStats(ctx context.Context) (int, error)
}

// TripContextStore is the read-only + fork accessor used by the trip-context
// middleware to detect when a request hits a public demo and must be
// redirected to the caller's per-session fork. It exposes the same surface
// the middleware needs and nothing more, so the middleware cannot
// accidentally start mutating trips directly.
type TripContextStore interface {
	GetFork(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	ForkTrip(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	GetTripMeta(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error)
}

// ── Places ───────────────────────────────────────────────────────────────────

// PlaceStore is the read/write contract for places.
type PlaceStore interface {
	GetPlace(ctx context.Context, placeID uuid.UUID) (*models.Place, error)
	ListPlacesByTrip(ctx context.Context, tripID uuid.UUID) ([]models.Place, error)
	CreatePlace(ctx context.Context, tripID uuid.UUID, p models.Place) (*models.Place, error)
	UpdatePlace(ctx context.Context, placeID uuid.UUID, updates map[string]any) (*models.Place, error)
	DeletePlace(ctx context.Context, placeID uuid.UUID) error
}

// ── Activities ───────────────────────────────────────────────────────────────

// ActivityStore is the read/write contract for activities.
type ActivityStore interface {
	GetByTripId(ctx context.Context, tripID uuid.UUID) ([]models.Activity, error)
	GetByPlaceId(ctx context.Context, placeID uuid.UUID) ([]models.Activity, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.Activity, error)
	CreateActivity(ctx context.Context, a *models.Activity) error
	UpdateActivity(ctx context.Context, a *models.Activity) error
	DeleteActivity(ctx context.Context, id uuid.UUID) error
}

// ── Expenses ─────────────────────────────────────────────────────────────────

// ExpenseStore is the read/write contract for expenses and their categories.
type ExpenseStore interface {
	GetCategories(ctx context.Context) ([]models.ExpenseCategory, error)
	CreateExpense(ctx context.Context, tripID *uuid.UUID, placeID *uuid.UUID, exp models.Expense) (*models.Expense, error)
	GetExpensesByTrip(ctx context.Context, tripID uuid.UUID) ([]models.Expense, error)
	GetExpensesSummary(ctx context.Context, tripID uuid.UUID) (*models.TripExpenseSummary, error)
	ListGlobalExpenses(ctx context.Context, tripID uuid.UUID) ([]models.Expense, error)
	ListPlaceExpenses(ctx context.Context, placeID uuid.UUID) ([]models.Expense, error)
	UpdateExpense(ctx context.Context, id uuid.UUID, exp models.Expense) (*models.Expense, error)
	DeleteExpense(ctx context.Context, id uuid.UUID) error
	GetPlaceExpensesSummary(ctx context.Context, placeID uuid.UUID) ([]models.CategorySummary, error)
}

// ── Auth ─────────────────────────────────────────────────────────────────────

// AuthStore is the read/write contract for users and the guest-to-user trip
// migration.
type AuthStore interface {
	CreateUser(ctx context.Context, email, password string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpgradeTrips(ctx context.Context, sessionID string, userID uuid.UUID) error
}

// ── Events ───────────────────────────────────────────────────────────────────

// EventStore persists analytics events. The implementation is expected to be
// fire-and-forget safe: errors are logged but never bubble up to the caller.
type EventStore interface {
	Insert(ctx context.Context, eventType, sessionID string, userID *string, tripID *string, metadataJSON []byte, createdAt time.Time) error
}

// ── Rate Limiting ────────────────────────────────────────────────────────────

// RateLimitStore is a fixed-window counter shared across all backend
// replicas. The implementation MUST be safe for concurrent use (Postgres'
// UPSERT satisfies that).
//
// CheckAndIncrement atomically increments the counter for the given key
// and returns true when the caller is still within the limit, false when
// the limit has been reached. The window resets automatically: if the
// current row's window_start is older than the supplied window, the
// counter is reset to 1 before the comparison.
type RateLimitStore interface {
	CheckAndIncrement(ctx context.Context, key string, window time.Duration, max int) (allowed bool, err error)
}
