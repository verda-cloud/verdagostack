// Package retry provides exponential backoff, polling, and periodic execution
// helpers using only the Go standard library. See README.md for examples.
package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"
)

// BackoffConfig controls exponential backoff parameters.
type BackoffConfig struct {
	Steps    int           // Maximum number of retry attempts.
	Duration time.Duration // Initial backoff duration.
	Factor   float64       // Multiplier applied to duration after each step.
	Jitter   float64       // Randomization factor (0 = none, 1 = ±100%).
}

// DefaultBackoff returns a reasonable default backoff configuration.
func DefaultBackoff() BackoffConfig {
	return BackoffConfig{
		Steps:    10,
		Duration: 5 * time.Second,
		Factor:   1.25,
		Jitter:   1.0,
	}
}

// ConditionFunc returns true when the condition is met, or an error to abort.
type ConditionFunc func() (done bool, err error)

// Retry calls fn with exponential backoff until it returns true, an error,
// or all attempts are exhausted.
func Retry(fn ConditionFunc, cfg BackoffConfig) error {
	d := cfg.Duration
	for i := range cfg.Steps {
		ok, err := fn()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if i == cfg.Steps-1 {
			break
		}
		sleep := d
		if cfg.Jitter > 0 {
			sleep = jitter(d, cfg.Jitter)
		}
		time.Sleep(sleep)
		d = time.Duration(float64(d) * cfg.Factor)
	}
	return fmt.Errorf("retry: timed out after %d attempts", cfg.Steps)
}

// Poll tries fn every interval until it returns true, an error, or the
// timeout elapses. The first check happens after one interval.
func Poll(ctx context.Context, interval, timeout time.Duration, fn ConditionFunc) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry: poll timed out: %w", ctx.Err())
		case <-ticker.C:
			ok, err := fn()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

// PollImmediate is like Poll but checks the condition immediately before
// waiting for the first interval.
func PollImmediate(ctx context.Context, interval, timeout time.Duration, fn ConditionFunc) error {
	ok, err := fn()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return Poll(ctx, interval, timeout, fn)
}

// RunImmediatelyThenPeriod executes fn once synchronously and returns an error
// if it fails. On success, it starts a background goroutine that re-runs fn
// every period until ctx is cancelled. Errors in periodic runs are ignored.
func RunImmediatelyThenPeriod(ctx context.Context, fn func(context.Context) error, period time.Duration) error {
	if err := fn(ctx); err != nil {
		return fmt.Errorf("retry: initial execution failed: %w", err)
	}

	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = fn(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func jitter(d time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0 {
		return d
	}
	delta := maxFactor * float64(d)
	return d + time.Duration(math.Floor(delta*rand.Float64())) //nolint:gosec // G404: crypto/rand not needed for jitter timing
}
