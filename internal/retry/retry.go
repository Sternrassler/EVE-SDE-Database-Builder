// Package retry provides exponential backoff retry logic for transient errors.
package retry

import (
	"context"
	"math/rand"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// Policy defines the retry behavior configuration
type Policy struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// RetryFunc is a function that can be retried
type RetryFunc func() error

// NewPolicy creates a new retry policy with the specified parameters
func NewPolicy(maxRetries int, initialDelay, maxDelay time.Duration) *Policy {
	return &Policy{
		MaxRetries:   maxRetries,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// DefaultPolicy returns a policy with sensible defaults (3 retries, 100ms initial, 5s max)
func DefaultPolicy() *Policy {
	return NewPolicy(3, 100*time.Millisecond, 5*time.Second)
}

// Do executes the given function with retry logic according to the policy
// Only errors marked as ErrorTypeRetryable will be retried
func (p *Policy) Do(ctx context.Context, fn RetryFunc) error {
	var lastErr error

	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// Check context before attempting (except first attempt)
		if attempt > 0 {
			if err := checkContext(ctx); err != nil {
				return err
			}
		}

		// Execute the function
		err := fn()

		// Success - return immediately
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !errors.IsRetryable(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == p.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt, p)

		// Wait for backoff duration or context cancellation
		if err := waitWithContext(ctx, delay); err != nil {
			return err
		}
	}

	return lastErr
}

// DoWithResult executes a function that returns a value and an error
// with retry logic according to the policy
func DoWithResult[T any](ctx context.Context, policy *Policy, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		// Check context before attempting (except first attempt)
		if attempt > 0 {
			if err := checkContext(ctx); err != nil {
				return result, err
			}
		}

		// Execute the function
		res, err := fn()

		// Success - return immediately
		if err == nil {
			return res, nil
		}

		lastErr = err

		// Check if error is retryable
		if !errors.IsRetryable(err) {
			return result, err
		}

		// Don't sleep after the last attempt
		if attempt == policy.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt, policy)

		// Wait for backoff duration or context cancellation
		if err := waitWithContext(ctx, delay); err != nil {
			return result, err
		}
	}

	return result, lastErr
}

// calculateBackoff computes the delay for a given attempt using exponential backoff
func calculateBackoff(attempt int, policy *Policy) time.Duration {
	// Calculate exponential backoff: initialDelay * multiplier^attempt
	delay := float64(policy.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= policy.Multiplier
	}

	// Cap at MaxDelay
	if delay > float64(policy.MaxDelay) {
		delay = float64(policy.MaxDelay)
	}

	// Apply jitter if enabled (±10%)
	if policy.Jitter {
		jitterRange := delay * 0.1
		jitter := (rand.Float64() * 2 * jitterRange) - jitterRange
		delay += jitter

		// Ensure delay is positive and doesn't exceed max
		if delay < 0 {
			delay = 0
		}
		if delay > float64(policy.MaxDelay) {
			delay = float64(policy.MaxDelay)
		}
	}

	return time.Duration(delay)
}
