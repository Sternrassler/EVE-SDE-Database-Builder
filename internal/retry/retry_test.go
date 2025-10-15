package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// TestNewPolicy tests the creation of a new retry policy
func TestNewPolicy(t *testing.T) {
	policy := NewPolicy(5, 50*time.Millisecond, 10*time.Second)

	if policy.MaxRetries != 5 {
		t.Errorf("expected MaxRetries=5, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 50*time.Millisecond {
		t.Errorf("expected InitialDelay=50ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 10*time.Second {
		t.Errorf("expected MaxDelay=10s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", policy.Multiplier)
	}

	if !policy.Jitter {
		t.Error("expected Jitter=true by default")
	}
}

// TestDefaultPolicy tests the default policy values
func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()

	if policy.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected InitialDelay=100ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 5*time.Second {
		t.Errorf("expected MaxDelay=5s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", policy.Multiplier)
	}

	if !policy.Jitter {
		t.Error("expected Jitter=true")
	}
}

// TestDo_Success tests successful execution without retries
func TestDo_Success(t *testing.T) {
	policy := DefaultPolicy()
	attempts := 0

	err := policy.Do(context.Background(), func() error {
		attempts++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

// TestDo_SuccessAfterRetries tests successful execution after 2 failures
func TestDo_SuccessAfterRetries(t *testing.T) {
	policy := NewPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
	attempts := 0

	err := policy.Do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return apperrors.NewRetryable("transient error", nil)
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

// TestDo_MaxRetriesReached tests that retry stops after max attempts
func TestDo_MaxRetriesReached(t *testing.T) {
	policy := NewPolicy(2, 10*time.Millisecond, 100*time.Millisecond)
	attempts := 0

	err := policy.Do(context.Background(), func() error {
		attempts++
		return apperrors.NewRetryable("persistent error", nil)
	})

	if err == nil {
		t.Error("expected error after max retries")
	}

	// MaxRetries=2 means initial attempt + 2 retries = 3 total attempts
	if attempts != 3 {
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
}

// TestDo_NonRetryableError tests that non-retryable errors are not retried
func TestDo_NonRetryableError(t *testing.T) {
	policy := DefaultPolicy()
	attempts := 0

	err := policy.Do(context.Background(), func() error {
		attempts++
		return apperrors.NewFatal("fatal error", nil)
	})

	if err == nil {
		t.Error("expected fatal error")
	}

	if !apperrors.IsFatal(err) {
		t.Error("expected fatal error type")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

// TestDo_StandardError tests that standard errors are not retried
func TestDo_StandardError(t *testing.T) {
	policy := DefaultPolicy()
	attempts := 0

	err := policy.Do(context.Background(), func() error {
		attempts++
		return errors.New("standard error")
	})

	if err == nil {
		t.Error("expected error")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt for standard error, got %d", attempts)
	}
}

// TestDo_ContextCancellation tests that context cancellation stops retries
func TestDo_ContextCancellation(t *testing.T) {
	policy := NewPolicy(5, 50*time.Millisecond, 1*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	err := policy.Do(ctx, func() error {
		attempts++
		return apperrors.NewRetryable("retryable error", nil)
	})

	if err == nil {
		t.Error("expected context error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	// Should have made at least one attempt, but not all 6 (1 + 5 retries)
	if attempts < 1 || attempts > 3 {
		t.Errorf("expected 1-3 attempts before timeout, got %d", attempts)
	}
}

// TestDoWithResult_Success tests successful execution with result
func TestDoWithResult_Success(t *testing.T) {
	policy := DefaultPolicy()
	attempts := 0

	result, err := DoWithResult(context.Background(), policy, func() (int, error) {
		attempts++
		return 42, nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != 42 {
		t.Errorf("expected result=42, got %d", result)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

// TestDoWithResult_SuccessAfterRetries tests successful result after retries
func TestDoWithResult_SuccessAfterRetries(t *testing.T) {
	policy := NewPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
	attempts := 0

	result, err := DoWithResult(context.Background(), policy, func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", apperrors.NewRetryable("transient error", nil)
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected result='success', got '%s'", result)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

// TestDoWithResult_NonRetryableError tests non-retryable error with result
func TestDoWithResult_NonRetryableError(t *testing.T) {
	policy := DefaultPolicy()
	attempts := 0

	result, err := DoWithResult(context.Background(), policy, func() (int, error) {
		attempts++
		return 0, apperrors.NewValidation("validation error", nil)
	})

	if err == nil {
		t.Error("expected validation error")
	}

	if result != 0 {
		t.Errorf("expected zero result on error, got %d", result)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

// TestCalculateBackoff tests backoff delay calculation
func TestCalculateBackoff(t *testing.T) {
	policy := &Policy{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},   // 100ms * 2^0
		{1, 200 * time.Millisecond},   // 100ms * 2^1
		{2, 400 * time.Millisecond},   // 100ms * 2^2
		{3, 800 * time.Millisecond},   // 100ms * 2^3
		{4, 1600 * time.Millisecond},  // 100ms * 2^4
		{5, 3200 * time.Millisecond},  // 100ms * 2^5
		{6, 5000 * time.Millisecond},  // Capped at MaxDelay
		{10, 5000 * time.Millisecond}, // Capped at MaxDelay
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			delay := calculateBackoff(tt.attempt, policy)
			if delay != tt.expected {
				t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, delay)
			}
		})
	}
}

// TestCalculateBackoff_WithJitter tests backoff with jitter enabled
func TestCalculateBackoff_WithJitter(t *testing.T) {
	policy := &Policy{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}

	// Run multiple times to check jitter variation
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = calculateBackoff(1, policy)
	}

	// Check that delay is within reasonable range (100ms-300ms for attempt 1)
	// Base delay for attempt 1 is 200ms, jitter ±10% = 180-220ms
	for i, delay := range delays {
		if delay < 150*time.Millisecond || delay > 250*time.Millisecond {
			t.Errorf("delay %d out of expected range (150-250ms): %v", i, delay)
		}
	}

	// Check that not all delays are identical (jitter is working)
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}

	// It's theoretically possible all are the same, but extremely unlikely
	if allSame {
		t.Log("Warning: all jittered delays are identical (unlikely but possible)")
	}
}

// TestCalculateBackoff_MaxDelayRespected tests that MaxDelay is never exceeded
func TestCalculateBackoff_MaxDelayRespected(t *testing.T) {
	policy := &Policy{
		MaxRetries:   10,
		InitialDelay: 1 * time.Second,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}

	// Test many attempts to ensure jitter never exceeds max
	for attempt := 0; attempt < 20; attempt++ {
		delay := calculateBackoff(attempt, policy)
		if delay > policy.MaxDelay {
			t.Errorf("attempt %d: delay %v exceeds MaxDelay %v", attempt, delay, policy.MaxDelay)
		}
	}
}

// TestDo_Timing tests that retries actually wait
func TestDo_Timing(t *testing.T) {
	policy := NewPolicy(2, 50*time.Millisecond, 200*time.Millisecond)
	policy.Jitter = false // Disable jitter for predictable timing

	attempts := 0
	start := time.Now()

	err := policy.Do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return apperrors.NewRetryable("retryable", nil)
		}
		return nil
	})

	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should have waited 50ms + 100ms = 150ms minimum
	// Allow some tolerance for execution time
	minExpected := 140 * time.Millisecond
	maxExpected := 250 * time.Millisecond

	if elapsed < minExpected || elapsed > maxExpected {
		t.Errorf("expected elapsed time between %v and %v, got %v", minExpected, maxExpected, elapsed)
	}
}

// TestDo_ContextCancelImmediate tests immediate context cancellation
func TestDo_ContextCancelImmediate(t *testing.T) {
	policy := DefaultPolicy()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	attempts := 0
	err := policy.Do(ctx, func() error {
		attempts++
		return apperrors.NewRetryable("retryable", nil)
	})

	// First attempt should execute, then context check should fail
	if err != context.Canceled {
		t.Errorf("expected Canceled error, got %v", err)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt before context check, got %d", attempts)
	}
}

// TestDoWithResult_ContextCancellation tests context cancellation with result
func TestDoWithResult_ContextCancellation(t *testing.T) {
	policy := NewPolicy(5, 50*time.Millisecond, 1*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	result, err := DoWithResult(ctx, policy, func() (int, error) {
		attempts++
		return 0, apperrors.NewRetryable("retryable error", nil)
	})

	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	if result != 0 {
		t.Errorf("expected zero result on error, got %d", result)
	}

	// Should have made at least one attempt, but not all
	if attempts < 1 || attempts > 3 {
		t.Errorf("expected 1-3 attempts before timeout, got %d", attempts)
	}
}

// TestDifferentErrorTypes tests handling of different error types
func TestDifferentErrorTypes(t *testing.T) {
	policy := NewPolicy(2, 10*time.Millisecond, 100*time.Millisecond)

	tests := []struct {
		name             string
		errorFunc        func() error
		expectedRetry    bool
		expectedAttempts int
	}{
		{
			name: "Retryable Error",
			errorFunc: func() error {
				return apperrors.NewRetryable("retry me", nil)
			},
			expectedRetry:    true,
			expectedAttempts: 3, // 1 initial + 2 retries
		},
		{
			name: "Fatal Error",
			errorFunc: func() error {
				return apperrors.NewFatal("fatal", nil)
			},
			expectedRetry:    false,
			expectedAttempts: 1,
		},
		{
			name: "Validation Error",
			errorFunc: func() error {
				return apperrors.NewValidation("invalid", nil)
			},
			expectedRetry:    false,
			expectedAttempts: 1,
		},
		{
			name: "Skippable Error",
			errorFunc: func() error {
				return apperrors.NewSkippable("skip", nil)
			},
			expectedRetry:    false,
			expectedAttempts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0
			err := policy.Do(context.Background(), func() error {
				attempts++
				return tt.errorFunc()
			})

			if err == nil {
				t.Error("expected error, got nil")
			}

			if attempts != tt.expectedAttempts {
				t.Errorf("expected %d attempts, got %d", tt.expectedAttempts, attempts)
			}
		})
	}
}

// BenchmarkRetry_Success benchmarks successful retry without delays
func BenchmarkRetry_Success(b *testing.B) {
policy := NewPolicy(3, 1*time.Nanosecond, 10*time.Nanosecond)
policy.Jitter = false

b.ResetTimer()
for i := 0; i < b.N; i++ {
_ = policy.Do(context.Background(), func() error {
return nil
})
}
}

// BenchmarkRetry_OneRetry benchmarks retry with one failure
func BenchmarkRetry_OneRetry(b *testing.B) {
policy := NewPolicy(3, 1*time.Nanosecond, 10*time.Nanosecond)
policy.Jitter = false

b.ResetTimer()
for i := 0; i < b.N; i++ {
attempts := 0
_ = policy.Do(context.Background(), func() error {
attempts++
if attempts == 1 {
return apperrors.NewRetryable("retry", nil)
}
return nil
})
}
}

// BenchmarkRetry_NonRetryable benchmarks non-retryable error path
func BenchmarkRetry_NonRetryable(b *testing.B) {
policy := DefaultPolicy()

b.ResetTimer()
for i := 0; i < b.N; i++ {
_ = policy.Do(context.Background(), func() error {
return apperrors.NewFatal("fatal", nil)
})
}
}

// BenchmarkCalculateBackoff benchmarks backoff calculation
func BenchmarkCalculateBackoff(b *testing.B) {
policy := DefaultPolicy()

b.ResetTimer()
for i := 0; i < b.N; i++ {
_ = calculateBackoff(5, policy)
}
}

// BenchmarkDoWithResult_Success benchmarks successful generic retry
func BenchmarkDoWithResult_Success(b *testing.B) {
policy := NewPolicy(3, 1*time.Nanosecond, 10*time.Nanosecond)
policy.Jitter = false

b.ResetTimer()
for i := 0; i < b.N; i++ {
_, _ = DoWithResult(context.Background(), policy, func() (int, error) {
return 42, nil
})
}
}
