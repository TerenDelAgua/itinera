// Package jobs contains the long-running background tasks owned by the
// Itinera backend. Each job is independent, takes its own context, and
// must respect the spec's GDPR retention rules (Spec 017 §5.10).
//
// Job lifecycle:
//
//  1. main.go constructs the job with its dependencies (AuthStore).
//  2. main.go calls Start(ctx) on a goroutine.
//  3. The job's Run loop fires on every interval until the parent
//     context is cancelled (graceful shutdown via signal handler).
//  4. Each cycle is wrapped in its own context derived from the parent
//     so a slow query cannot leak across shutdown.
package jobs

import (
	"context"
	"errors"
	"log"
	"time"
)

// AccountCleanupConfig tunes the AccountCleanupJob. Zero values are
// "sensible production defaults" — callers can override any of them.
type AccountCleanupConfig struct {
	// Window is the grace period between soft-delete (deleted_at) and
	// hard-delete. Spec 017 §5.10 mandates 30 days.
	Window time.Duration

	// Interval is the period between Run cycles. Default: 24h, so we
	// only need to catch a single day's worth of expired users in each
	// pass — a tiny workload that fits inside the bcrypt cost budget.
	Interval time.Duration

	// BatchSize bounds the number of users anonymised per cycle. Spec
	// §5.10 mandates ≤1000 to keep the worst-case bcrypt cost < 100s.
	BatchSize int

	// FirstDelay lets the caller postpone the first Run (e.g. waiting
	// for migrations to complete). Default: 0 (run immediately).
	FirstDelay time.Duration
}

func (c AccountCleanupConfig) withDefaults() AccountCleanupConfig {
	if c.Window == 0 {
		c.Window = 30 * 24 * time.Hour
	}
	if c.Interval == 0 {
		c.Interval = 24 * time.Hour
	}
	if c.BatchSize == 0 {
		c.BatchSize = 1000
	}
	return c
}

// AccountCleanupJob hard-deletes users whose soft-delete timestamp is
// older than the configured Window (Spec 017 §5.10).
//
// The job's only dependency is the AuthStore. We deliberately keep it
// narrow — the work itself lives in AuthStore.HardDeleteExpired — so
// the job is responsible for timing, context plumbing, and observability,
// and the storage layer is responsible for the SQL.
type AccountCleanupJob struct {
	auth  HardDeleter
	cfg   AccountCleanupConfig
	clock func() time.Time // injectable for tests; defaults to time.Now
}

// HardDeleter is the slice of database.AuthStore this job needs. Tests
// can pass a fake without depending on pgx.
type HardDeleter interface {
	HardDeleteExpired(ctx context.Context, after time.Duration, limit int) (int, error)
}

// NewAccountCleanupJob wires the job. The caller is expected to have a
// fully-initialised AuthStore ready to take queries.
func NewAccountCleanupJob(auth HardDeleter, cfg AccountCleanupConfig) *AccountCleanupJob {
	return &AccountCleanupJob{
		auth:  auth,
		cfg:   cfg.withDefaults(),
		clock: time.Now,
	}
}

// Start launches the ticker loop on a new goroutine and returns
// immediately. When parentCtx is cancelled the loop exits cleanly,
// allowing the in-flight cycle (if any) to abort via its derived
// context. Always call Stop (or cancel the parent) for graceful
// shutdown — otherwise the goroutine leaks for the life of the process.
//
// Start is intentionally non-blocking. Errors during a cycle are logged
// and the next cycle still fires — we never let one bad run stall the
// whole job forever (Spec §5.10 "count is 0 or fails → log warn,
// next tick").
func (j *AccountCleanupJob) Start(parentCtx context.Context) {
	go j.run(parentCtx)
}

// RunOnce executes a single cycle. Exposed so tests (and operational
// tooling) can drive the job deterministically without a ticker.
func (j *AccountCleanupJob) RunOnce(ctx context.Context) (int, error) {
	return j.auth.HardDeleteExpired(ctx, j.cfg.Window, j.cfg.BatchSize)
}

func (j *AccountCleanupJob) run(parentCtx context.Context) {
	if j.cfg.FirstDelay > 0 {
		select {
		case <-parentCtx.Done():
			return
		case <-time.After(j.cfg.FirstDelay):
		}
	}

	t := time.NewTicker(j.cfg.Interval)
	defer t.Stop()

	// Fire once at startup so a freshly-deployed backend can catch up on
	// any users that should have been anonymised while the previous
	// instance was down.
	j.cycle(parentCtx)

	for {
		select {
		case <-parentCtx.Done():
			log.Printf("[account-cleanup] parent context cancelled, exiting (window=%s)", j.cfg.Window)
			return
		case <-t.C:
			j.cycle(parentCtx)
		}
	}
}

func (j *AccountCleanupJob) cycle(parentCtx context.Context) {
	// Bound the cycle independently of the parent context. If shutdown
	// takes longer than 60s we give up — the next instance picks up the
	// rows on its first-cycle sweep.
	cycleCtx, cancel := context.WithTimeout(parentCtx, 60*time.Second)
	defer cancel()

	start := j.clock()
	n, err := j.RunOnce(cycleCtx)
	duration := j.clock().Sub(start)
	if err != nil {
		// Err is typically context.Canceled during shutdown — that's
		// not an error condition, just an early exit.
		if !errors.Is(err, context.Canceled) && !errors.Is(parentCtx.Err(), context.Canceled) {
			log.Printf("[account-cleanup] cycle error after %s: %v", duration, err)
		}
		return
	}
	log.Printf("[account-cleanup] cycle ok: %d anonymised in %s", n, duration)
}
