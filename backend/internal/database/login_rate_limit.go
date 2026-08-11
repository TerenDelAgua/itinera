package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LoginRateLimitRepository composes two fixed windows on top of the
// existing rate_limits table:
//
//   - burst (1 minute, 5 max): catches the trivial brute-force case.
//   - long  (1 hour,   10 max): catches the slow trickle from a botnet.
//
// A request is allowed only when BOTH windows have headroom. Crossing
// the long window puts the IP in a 1h cooldown; once the window expires,
// the next request becomes the first attempt in a fresh window and the
// caller is allowed again.
//
// Spec 017 §4.6 mandates this two-window shape; the simpler RateLimitRepository
// is kept for events (fire-and-forget) where a single window is enough.
type LoginRateLimitRepository struct {
	Inner     *RateLimitRepository
	BurstMax  int
	BurstSpan time.Duration
	LongMax   int
	LongSpan  time.Duration
}

// NewLoginRateLimitRepository defaults to the Spec §4.6 values. Override
// the fields directly (e.g. in tests) to exercise edge cases cheaply.
func NewLoginRateLimitRepository(pool *pgxpool.Pool) *LoginRateLimitRepository {
	return &LoginRateLimitRepository{
		Inner:     NewRateLimitRepository(pool),
		BurstMax:  5,
		BurstSpan: 1 * time.Minute,
		LongMax:   10,
		LongSpan:  1 * time.Hour,
	}
}

// IP returns the keys used for a given IP. Exported so the tests (and any
// future admin endpoints) can introspect or reset specific keys.
func (r *LoginRateLimitRepository) IP(ip string) (burst, long_ string) {
	return "login_burst:ip:" + ip, "login_long:ip:" + ip
}

// record returns (allowed, blocked, err) where:
//   - allowed = the call may proceed
//   - blocked = the long window has filled up; the IP is in cooldown
//
// `blocked` is a stronger guarantee than `!allowed`: even legitimate
// credentials will be rejected. The handler turns this into a 429 with
// code RATE_LIMITED.
func (r *LoginRateLimitRepository) record(ctx context.Context, burstKey, longKey string) (allowed, blocked bool, err error) {
	burstOK, err := r.Inner.CheckAndIncrement(ctx, burstKey, r.BurstSpan, r.BurstMax)
	if err != nil {
		return false, false, fmt.Errorf("login burst: %w", err)
	}

	longOK, err := r.Inner.CheckAndIncrement(ctx, longKey, r.LongSpan, r.LongMax)
	if err != nil {
		return false, false, fmt.Errorf("login long: %w", err)
	}

	if !longOK {
		return false, true, nil
	}
	if !burstOK {
		return false, false, nil
	}
	return true, false, nil
}

// IsBlocked answers whether the IP is currently inside its 1h cooldown
// WITHOUT bumping any counter. Useful for "am I allowed to attempt a
// password reset email yet?" lookups without consuming budget.
func (r *LoginRateLimitRepository) IsBlocked(ctx context.Context, ip string) (bool, error) {
	_, longKey := r.IP(ip)
	ok, err := r.Inner.CheckAndIncrement(ctx, longKey, r.LongSpan, r.LongMax-1)
	if err != nil {
		return false, err
	}
	return !ok, nil
}

// RecordFailure consumes a single attempt (so a successful login can be
// followed by RecordSuccess to roll the counter back). Returns the same
// (allowed, blocked, err) tuple as the private helper.
func (r *LoginRateLimitRepository) RecordFailure(ctx context.Context, ip string) (allowed, blocked bool, err error) {
	burstKey, longKey := r.IP(ip)
	return r.record(ctx, burstKey, longKey)
}

// ResetOnSuccess rolls the counters for the IP back to zero. Called from
// /auth/v2/login after a successful credential match so a legitimate
// user who mistypes 4 times and then succeeds isn't punished.
func (r *LoginRateLimitRepository) ResetOnSuccess(ctx context.Context, ip string) error {
	burstKey, longKey := r.IP(ip)
	if _, err := r.Inner.Pool.Exec(ctx,
		`DELETE FROM public.rate_limits WHERE rate_key IN ($1, $2)`,
		burstKey, longKey,
	); err != nil {
		return fmt.Errorf("rate limit reset: %w", err)
	}
	return nil
}