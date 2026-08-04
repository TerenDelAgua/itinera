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
	"github.com/jackc/pgx/v5"
)

// ── Trips ────────────────────────────────────────────────────────────────────

// TripStore is the read/write contract for trips.
type TripStore interface {
	CreateTrip(ctx context.Context, userID *uuid.UUID, sessionID *string, isInternal bool, tripData models.Trip) (*models.Trip, error)
	GetTrip(ctx context.Context, id string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	GetTripById(ctx context.Context, id uuid.UUID) (*models.Trip, error)
	ListTrips(ctx context.Context, userID *uuid.UUID, sessionID *string) ([]models.Trip, error)
	UpdateTrip(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string, updates map[string]any) (*models.Trip, error)
	DeleteTrip(ctx context.Context, id string, userID *uuid.UUID, sessionID *string) error

	// GetFork returns the caller's existing fork of a public demo, if any.
	// Used by the trip-context middleware to detect ghost-mode edits.
	GetFork(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)

	// GetTripMeta returns (isOwner, isPublicDemo, err) used by the events
	// handler to scope analytics without exposing the full repository.
	GetTripMeta(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error)

	// Public demo content for the landing page.
	ListPublicDemos(ctx context.Context, limit int) ([]models.Trip, error)
	GetPublicStats(ctx context.Context) (int, error)

	//Share trip
	EnableShare(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (token string, expiresAt *time.Time, err error)
	DisableShare(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) error
	GetByShareToken(ctx context.Context, token string) (*models.Trip, error)
	CountActiveSharesByUser(ctx context.Context, userID *uuid.UUID, sessionID *string) (int, error)
	GetActiveShareTripID(ctx context.Context, token string) (uuid.UUID, error)
	IsPublicDemo(ctx context.Context, tripID uuid.UUID) (bool, error)

	//Clone helpers
	GetBaseForClone(ctx context.Context, tx pgx.Tx, origID uuid.UUID) (*models.Trip, error)
	InsertFork(ctx context.Context, tx pgx.Tx, newTripID uuid.UUID, userId *uuid.UUID, sessionID *string, isInternal bool, origID uuid.UUID, base *models.Trip) error
}

type TripForker interface {
	GetFork(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	ForkFromDemo(ctx context.Context, demoTripID uuid.UUID, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
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
	//Clone helper
	CloneByTripID(ctx context.Context, tx pgx.Tx, origTripID, newTripID uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
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

	//Clone helper
	CloneByTripID(ctx context.Context, tx pgx.Tx, origTripID, newTripID uuid.UUID, placeMap map[uuid.UUID]uuid.UUID) error
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

	//Clone helper
	CloneByTripID(ctx context.Context, tx pgx.Tx, origTripID, newTripID uuid.UUID, userID *uuid.UUID, placeMap map[uuid.UUID]uuid.UUID) error
}

// ── Auth ─────────────────────────────────────────────────────────────────────

// AuthStore is the read/write contract for users and the guest-to-user trip
// migration. Each method maps 1:1 to a Spec 017 use case (see comments).
//
// Conventions:
//   - Every method that accepts email goes through normalizeEmail internally,
//     so casing and stray whitespace never escape the repo layer.
//   - All "Get" methods return pgx.ErrNoRows when the row is missing; the
//     caller decides the HTTP shape (404 vs 401 vs 403) without leaking the
//     distinction in the API surface.
type AuthStore interface {
	// CreateUser persists a new account with a normalised email and the
	// user's chosen locale. terms_accepted_at is set explicitly by the
	// handler so the check that rejects terms_accepted=false stays in the
	// HTTP layer (and can return 400 TERMS_NOT_ACCEPTED).
	CreateUser(ctx context.Context, email, password, locale string) (*models.User, error)

	// GetUserByEmail resolves the login flow and the forgot-password flow.
	// Soft-deleted users (deleted_at IS NOT NULL) are returned; the login
	// handler distinguishes them via user.DeletedAt and returns 403.
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// GetUserByID is used by /auth/me and the future access-token middleware.
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// ClaimGuestTrips moves the guest's trips into the user's account.
	// session_id is preserved on the trip (the cookie is still usable).
	ClaimGuestTrips(ctx context.Context, sessionID string, userID uuid.UUID) (claimed int, err error)

	// SoftDeleteUser marks the account as soft-deleted (deleted_at = now()).
	// Returns the email so callers can clear in-memory caches.
	SoftDeleteUser(ctx context.Context, userID uuid.UUID) (email string, err error)

	// MarkUserAsHardDeleted anonymises a soft-deleted account after the
	// 30-day retention window (Spec 017 §5.10). The row stays so analytics
	// FK references remain valid.
	MarkUserAsHardDeleted(ctx context.Context, userID uuid.UUID, randomStringHash string) error

	// UpdateUserPassword writes a fresh bcrypt hash. Used by reset-password
	// after the 6-digit code check passes.
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
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

// ── Analytics ────────────────────────────────────────────────────────────────

// AnalyticsStore is the read contract for the admin-only analytics module.
// Every method reads from v_real_sessions / v_real_trips / v_user_engagement
// (or from events with the real-sessions filter applied), so the
// "internal = excluded" invariant is centralised in the SQL views.
type AnalyticsStore interface {
	GetAnalyticsOverview(ctx context.Context) (models.AnalyticsOverview, error)
	GetAnalyticsActivation(ctx context.Context) (models.AnalyticsActivation, error)
	GetAnalyticsEngagement(ctx context.Context) (models.AnalyticsEngagement, error)
	GetAnalyticsRetention(ctx context.Context, days int) (models.AnalyticsRetention, error)
	GetAnalyticsFunnel(ctx context.Context) (models.AnalyticsFunnel, error)
	GetAnalyticsPowerUsers(ctx context.Context, limit int) (models.AnalyticsPowerUsers, error)

	// Internal-vs-real session/trip counts (Spec 015 v1).
	// Returns (internalTrips, realTrips, internalSessions, realSessions).
	GetAnalyticsSessions(ctx context.Context) (int, int, int, int, error)
}

// ── Sessions (Spec 017 §3.1 #2) ──────────────────────────────────────────────

// SessionStore owns the `sessions` table introduced by migration 017. Only
// hashed tokens are stored here; the raw tokens are kept in HttpOnly cookies
// (Spec 017 §4.5 constant-time compare).
type SessionStore interface {
	CreateSession(ctx context.Context, userID uuid.UUID, accessHash, refreshHash string, expiresAt time.Time, userAgent, ipAddress *string) (*models.Session, error)
	FindSessionByAccessTokenHash(ctx context.Context, accessHash string) (*models.Session, error)
	RotateSession(ctx context.Context, sessionID uuid.UUID, newAccessHash, newRefreshHash string, newExpiry time.Time) error
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	RevokeAllSessionsForUser(ctx context.Context, userID uuid.UUID) error
	CountActiveSessionsForUser(ctx context.Context, userID uuid.UUID) (int, error)
	CleanupExpiredSessions(ctx context.Context) (int64, error)
}

// ── PasswordReset (Spec 017 §3.1 #3) ──────────────────────────────────────────

// PasswordResetStore owns the `password_reset_tokens` table. It enforces the
// 5-attempts lockout and the "one active code per user" invariant described
// in Spec 017 §5.7; rate limiting at the IP level lives elsewhere and must
// happen BEFORE this store is called (see Spec 017 §4.6).
type PasswordResetStore interface {
	Create(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time, ip *string) error
	MarkPreviousAsUsed(ctx context.Context, userID uuid.UUID) error
	FindActiveByHash(ctx context.Context, hash string) (*models.PasswordResetToken, error)
	FindActiveByUser(ctx context.Context, userID uuid.UUID) (*models.PasswordResetToken, error)
	RecordFailedAttempt(ctx context.Context, tokenID uuid.UUID) (attempts int, locked bool, err error)
	MarkUsed(ctx context.Context, tokenID uuid.UUID) error
	HardDeleteOldTokens(ctx context.Context) (int64, error)
}
