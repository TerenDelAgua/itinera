package jobs

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHardDeleter captures how the job called it. We don't need a real
// database because the actual SQL lives in database.AuthRepository (and is
// covered by integration tests in package database). The job's contract
// is just "call HardDeleteExpired(ctx, window, batchSize) and propagate
// the result".
type fakeHardDeleter struct {
	calls   atomic.Int32
	lastWin atomic.Value // time.Duration
	lastLim atomic.Value // int
	out     atomic.Int32
	err     error

	// delay forces the cycle to take a moment — used to assert that
	// cancellation actually interrupts an in-flight RunOnce.
	delay time.Duration
}

func (f *fakeHardDeleter) HardDeleteExpired(ctx context.Context, after time.Duration, limit int) (int, error) {
	f.calls.Add(1)
	f.lastWin.Store(after)
	f.lastLim.Store(limit)

	if f.delay > 0 {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(f.delay):
		}
	}

	if f.err != nil {
		return 0, f.err
	}
	return int(f.out.Load()), nil
}

// TestAccountCleanupJob_Defaults ensures the zero-value config picks the
// spec-mandated 30-day window, 24h interval, and 1000 batch.
func TestAccountCleanupJob_Defaults(t *testing.T) {
	t.Parallel()
	j := NewAccountCleanupJob(&fakeHardDeleter{}, AccountCleanupConfig{})
	assert.Equal(t, 30*24*time.Hour, j.cfg.Window)
	assert.Equal(t, 24*time.Hour, j.cfg.Interval)
	assert.Equal(t, 1000, j.cfg.BatchSize)
}

// TestAccountCleanupJob_Overrides: explicit config values win.
func TestAccountCleanupJob_Overrides(t *testing.T) {
	t.Parallel()
	j := NewAccountCleanupJob(&fakeHardDeleter{}, AccountCleanupConfig{
		Window:   7 * 24 * time.Hour,
		Interval: 1 * time.Hour,
		BatchSize: 50,
	})
	assert.Equal(t, 7*24*time.Hour, j.cfg.Window)
	assert.Equal(t, 1*time.Hour, j.cfg.Interval)
	assert.Equal(t, 50, j.cfg.BatchSize)
}

// TestAccountCleanupJob_RunOnce_Propagates ensures RunOnce passes the
// spec values straight through to the underlying store.
func TestAccountCleanupJob_RunOnce_Propagates(t *testing.T) {
	t.Parallel()
	deleter := &fakeHardDeleter{}
	deleter.out.Store(7)
	j := NewAccountCleanupJob(deleter, AccountCleanupConfig{
		Window:    14 * 24 * time.Hour,
		BatchSize: 500,
	})

	n, err := j.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, int32(1), deleter.calls.Load())

	gotWin, _ := deleter.lastWin.Load().(time.Duration)
	gotLim, _ := deleter.lastLim.Load().(int)
	assert.Equal(t, 14*24*time.Hour, gotWin)
	assert.Equal(t, 500, gotLim)
}

// TestAccountCleanupJob_RunOnce_Error covers the error-propagation
// path. The job itself swallows the error in cycle(), but RunOnce
// returns it so callers (tests, operational tooling) can react.
func TestAccountCleanupJob_RunOnce_Error(t *testing.T) {
	t.Parallel()
	deleter := &fakeHardDeleter{err: errors.New("boom")}
	j := NewAccountCleanupJob(deleter, AccountCleanupConfig{})
	_, err := j.RunOnce(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

// TestAccountCleanupJob_Ticker_Cancels confirms Start's goroutine
// honours the parent context's Done channel: cancelling parentCtx
// causes the loop to exit within a small grace period.
func TestAccountCleanupJob_Ticker_Cancels(t *testing.T) {
	t.Parallel()
	deleter := &fakeHardDeleter{}
	j := NewAccountCleanupJob(deleter, AccountCleanupConfig{
		Interval:   50 * time.Millisecond,
		FirstDelay: 0,
	})

	parentCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		j.run(parentCtx)
		close(done)
	}()

	// Let the startup-cycle run.
	time.Sleep(100 * time.Millisecond)
	assert.GreaterOrEqual(t, deleter.calls.Load(), int32(1))

	cancel()
	select {
	case <-done:
		// Good — goroutine exited.
	case <-time.After(2 * time.Second):
		t.Fatalf("run() did not exit after parent context cancel")
	}
}

// TestAccountCleanupJob_Ticker_FirstDelay confirms the FirstDelay
// parameter postpones the startup cycle. We use a very small delay
// (50ms) and assert no calls fired before then.
func TestAccountCleanupJob_Ticker_FirstDelay(t *testing.T) {
	t.Parallel()
	deleter := &fakeHardDeleter{}
	j := NewAccountCleanupJob(deleter, AccountCleanupConfig{
		Interval:   1 * time.Hour, // huge interval so only the startup cycle fires
		FirstDelay: 200 * time.Millisecond,
	})

	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		j.run(parentCtx)
		close(done)
	}()

	// 50ms < FirstDelay: no calls yet.
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(0), deleter.calls.Load(),
		"no calls expected before FirstDelay elapses")

	// 250ms > FirstDelay: at least one call.
	time.Sleep(250 * time.Millisecond)
	assert.GreaterOrEqual(t, deleter.calls.Load(), int32(1))

	cancel()
	<-done
}

// TestAccountCleanupJob_Ticker_CancelDuringCycle: an in-flight RunOnce
// must respect the parent's cancellation. We force the fake to take
// 200ms and cancel mid-flight — the next call count should not advance
// past the first one.
func TestAccountCleanupJob_Ticker_CancelDuringCycle(t *testing.T) {
	t.Parallel()
	deleter := &fakeHardDeleter{delay: 200 * time.Millisecond}
	j := NewAccountCleanupJob(deleter, AccountCleanupConfig{
		Interval: 24 * time.Hour, // only the startup cycle fires
	})

	parentCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		j.run(parentCtx)
		close(done)
	}()

	// Wait until the startup cycle starts (first call happens).
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	// At most one call completed (the in-flight one). The tick never
	// reached a second fire because Interval is huge.
	assert.LessOrEqual(t, deleter.calls.Load(), int32(1))
}