package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// TestContextTimeout_AbortsRetry tests that timeout cancels retry immediately
func TestContextTimeout_AbortsRetry(t *testing.T) {
	policy := NewPolicy(10, 50*time.Millisecond, 5*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	start := time.Now()

	err := policy.Do(ctx, func() error {
		attempts++
		return apperrors.NewRetryable("retryable error", nil)
	})

	elapsed := time.Since(start)

	// Should get DeadlineExceeded error
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	// Should have stopped early (not all 11 attempts: 1 + 10 retries)
	if attempts >= 6 {
		t.Errorf("expected early abort (< 6 attempts), got %d", attempts)
	}

	// Should have aborted within reasonable time (~100ms + small overhead)
	if elapsed > 200*time.Millisecond {
		t.Errorf("took too long to abort: %v", elapsed)
	}
}

// TestContextManualCancellation tests that manual cancel() stops retry
func TestContextManualCancellation(t *testing.T) {
	policy := NewPolicy(10, 20*time.Millisecond, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0

	// Cancel after first retry
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := policy.Do(ctx, func() error {
		attempts++
		return apperrors.NewRetryable("retryable error", nil)
	})

	// Should get Canceled error
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected Canceled, got %v", err)
	}

	// Should have made at least 1 attempt, but not all 11
	if attempts < 1 || attempts > 5 {
		t.Errorf("expected 1-5 attempts with cancellation, got %d", attempts)
	}
}

// TestContextManualCancellation_Immediate tests immediate cancellation
func TestContextManualCancellation_Immediate(t *testing.T) {
	policy := NewPolicy(5, 50*time.Millisecond, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before starting

	attempts := 0
	err := policy.Do(ctx, func() error {
		attempts++
		return apperrors.NewRetryable("retryable error", nil)
	})

	// First attempt executes, then context is checked
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected Canceled, got %v", err)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt before context check, got %d", attempts)
	}
}

// TestContextNormalRetry_NoCancel tests that normal retry works without cancellation
func TestContextNormalRetry_NoCancel(t *testing.T) {
	policy := NewPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
	ctx := context.Background()

	attempts := 0
	err := policy.Do(ctx, func() error {
		attempts++
		if attempts < 3 {
			return apperrors.NewRetryable("retryable error", nil)
		}
		return nil
	})

	// Should succeed without context error
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

// TestContextErrorWrapping tests that context errors are correctly returned
func TestContextErrorWrapping(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() (context.Context, context.CancelFunc)
		expectedErr error
	}{
		{
			name: "DeadlineExceeded",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			expectedErr: context.DeadlineExceeded,
		},
		{
			name: "Canceled",
			setupCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
			expectedErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := NewPolicy(5, 20*time.Millisecond, 1*time.Second)
			ctx, cancel := tt.setupCtx()
			defer cancel()

			err := policy.Do(ctx, func() error {
				return apperrors.NewRetryable("retryable", nil)
			})

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

// TestDoWithResult_ContextTimeout tests DoWithResult with timeout
func TestDoWithResult_ContextTimeout(t *testing.T) {
	policy := NewPolicy(10, 50*time.Millisecond, 5*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	result, err := DoWithResult(ctx, policy, func() (string, error) {
		attempts++
		return "", apperrors.NewRetryable("retryable error", nil)
	})

	// Should get DeadlineExceeded error
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	// Result should be zero value
	if result != "" {
		t.Errorf("expected empty result, got %s", result)
	}

	// Should have aborted early
	if attempts >= 6 {
		t.Errorf("expected early abort (< 6 attempts), got %d", attempts)
	}
}

// TestDoWithResult_ContextCancelDuringRetry tests cancellation during retry
func TestDoWithResult_ContextCancelDuringRetry(t *testing.T) {
	policy := NewPolicy(10, 20*time.Millisecond, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0

	// Cancel after a short delay
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	result, err := DoWithResult(ctx, policy, func() (int, error) {
		attempts++
		return 0, apperrors.NewRetryable("retryable error", nil)
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected Canceled, got %v", err)
	}

	if result != 0 {
		t.Errorf("expected zero result, got %d", result)
	}

	if attempts < 1 || attempts > 5 {
		t.Errorf("expected 1-5 attempts with cancellation, got %d", attempts)
	}
}

// TestDoWithResult_NormalRetryNoCancel tests normal operation without cancellation
func TestDoWithResult_NormalRetryNoCancel(t *testing.T) {
	policy := NewPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
	ctx := context.Background()

	attempts := 0
	result, err := DoWithResult(ctx, policy, func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", apperrors.NewRetryable("retryable error", nil)
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got %s", result)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

// TestContextCheckBeforeRetry verifies context is checked BEFORE retry attempt
func TestContextCheckBeforeRetry(t *testing.T) {
	policy := NewPolicy(5, 10*time.Millisecond, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	firstAttemptDone := make(chan bool, 1)

	err := policy.Do(ctx, func() error {
		attempts++
		if attempts == 1 {
			// Signal first attempt is done, then cancel
			firstAttemptDone <- true
		}
		return apperrors.NewRetryable("retryable error", nil)
	})

	// Cancel context after first attempt completes
	go func() {
		<-firstAttemptDone
		cancel()
	}()

	// Wait a bit for goroutine to execute
	time.Sleep(50 * time.Millisecond)

	// Re-run the test with controlled cancellation
	ctx2, cancel2 := context.WithCancel(context.Background())
	attempts2 := 0

	go func() {
		time.Sleep(15 * time.Millisecond) // After first attempt + backoff
		cancel2()
	}()

	err2 := policy.Do(ctx2, func() error {
		attempts2++
		return apperrors.NewRetryable("retryable error", nil)
	})

	if !errors.Is(err2, context.Canceled) {
		t.Errorf("expected Canceled, got %v", err2)
	}

	// Should have stopped before all retries
	if attempts2 > 3 {
		t.Errorf("expected early stop (≤3 attempts), got %d", attempts2)
	}

	_ = err // Suppress unused variable warning for first test run
}

// TestIntegrationRealTimeout is an integration test with realistic timeout scenario
func TestIntegrationRealTimeout(t *testing.T) {
	// Simulate a real-world scenario: API call with 500ms timeout
	policy := NewPolicy(5, 100*time.Millisecond, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	attempts := 0
	start := time.Now()

	err := policy.Do(ctx, func() error {
		attempts++
		// Simulate work that takes some time
		time.Sleep(10 * time.Millisecond)
		return apperrors.NewRetryable("API error", nil)
	})

	elapsed := time.Since(start)

	// Should timeout
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	// Should have completed within timeout + small overhead
	if elapsed > 600*time.Millisecond {
		t.Errorf("took too long: %v", elapsed)
	}

	// Should have made multiple attempts but not all
	if attempts < 2 || attempts > 5 {
		t.Errorf("expected 2-5 attempts in 500ms window, got %d", attempts)
	}

	t.Logf("Integration test: %d attempts in %v before timeout", attempts, elapsed)
}

// TestCheckContext tests the checkContext helper function
func TestCheckContext(t *testing.T) {
	t.Run("Active context returns nil", func(t *testing.T) {
		ctx := context.Background()
		err := checkContext(ctx)
		if err != nil {
			t.Errorf("expected nil for active context, got %v", err)
		}
	})

	t.Run("Cancelled context returns Canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := checkContext(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected Canceled, got %v", err)
		}
	})

	t.Run("Expired context returns DeadlineExceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond)
		err := checkContext(ctx)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected DeadlineExceeded, got %v", err)
		}
	})
}

// TestWaitWithContext tests the waitWithContext helper function
func TestWaitWithContext(t *testing.T) {
	t.Run("Normal wait completes", func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()
		err := waitWithContext(ctx, 50*time.Millisecond)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if elapsed < 40*time.Millisecond || elapsed > 80*time.Millisecond {
			t.Errorf("expected ~50ms wait, got %v", elapsed)
		}
	})

	t.Run("Context cancellation interrupts wait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		err := waitWithContext(ctx, 1*time.Second) // Long wait
		elapsed := time.Since(start)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected Canceled, got %v", err)
		}

		// Should have been interrupted early
		if elapsed > 100*time.Millisecond {
			t.Errorf("wait not interrupted quickly enough: %v", elapsed)
		}
	})

	t.Run("Timeout during wait", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := waitWithContext(ctx, 1*time.Second) // Long wait
		elapsed := time.Since(start)

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected DeadlineExceeded, got %v", err)
		}

		// Should timeout around 20ms
		if elapsed > 100*time.Millisecond {
			t.Errorf("timeout took too long: %v", elapsed)
		}
	})
}
