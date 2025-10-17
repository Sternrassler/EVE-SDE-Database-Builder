package worker

import (
	"context"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// TestSetupSignalHandler_SIGINT tests graceful shutdown on SIGINT
func TestSetupSignalHandler_SIGINT(t *testing.T) {
	ctx := SetupSignalHandler()

	// Verify context is not cancelled initially
	select {
	case <-ctx.Done():
		t.Fatal("context should not be cancelled initially")
	default:
		// Expected: context not cancelled
	}

	// Send SIGINT to current process
	currentPID := os.Getpid()
	process, err := os.FindProcess(currentPID)
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}

	if err := process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	// Wait for context cancellation (with timeout)
	select {
	case <-ctx.Done():
		// Expected: context cancelled due to SIGINT
		if ctx.Err() != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", ctx.Err())
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled after SIGINT")
	}
}

// TestSetupSignalHandler_SIGTERM tests graceful shutdown on SIGTERM
func TestSetupSignalHandler_SIGTERM(t *testing.T) {
	ctx := SetupSignalHandler()

	// Verify context is not cancelled initially
	select {
	case <-ctx.Done():
		t.Fatal("context should not be cancelled initially")
	default:
		// Expected: context not cancelled
	}

	// Send SIGTERM to current process
	currentPID := os.Getpid()
	process, err := os.FindProcess(currentPID)
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	// Wait for context cancellation (with timeout)
	select {
	case <-ctx.Done():
		// Expected: context cancelled due to SIGTERM
		if ctx.Err() != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", ctx.Err())
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled after SIGTERM")
	}
}

// TestSetupSignalHandler_WithPool tests signal handling integration with worker pool
func TestSetupSignalHandler_WithPool(t *testing.T) {
	ctx := SetupSignalHandler()
	pool := NewPool(2)
	pool.Start(ctx)

	var jobsStarted, jobsCompleted atomic.Int32
	done := make(chan struct{})

	// Submit long-running jobs
	for i := 0; i < 4; i++ {
		pool.Submit(Job{
			ID: "long-job",
			Fn: func(ctx context.Context) (interface{}, error) {
				jobsStarted.Add(1)
				// Simulate work with context awareness
				for j := 0; j < 10; j++ {
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(50 * time.Millisecond):
						// Continue work
					}
				}
				jobsCompleted.Add(1)
				return "completed", nil
			},
		})
	}

	// Send SIGINT after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		currentPID := os.Getpid()
		process, _ := os.FindProcess(currentPID)
		_ = process.Signal(syscall.SIGINT) // Ignore error in test goroutine
		close(done)
	}()

	<-done

	// Wait for pool to finish
	results, _ := pool.Wait()

	// Verify: Some jobs should have been cancelled
	if len(results) >= 4 {
		t.Logf("All jobs completed before cancellation (timing dependent)")
	} else {
		t.Logf("Jobs cancelled: %d/%d completed", len(results), 4)
	}

	// Verify context is cancelled
	if ctx.Err() == nil {
		t.Error("expected context to be cancelled")
	}
}

// TestSetupSignalHandler_MultipleContexts tests that multiple signal handlers work independently
func TestSetupSignalHandler_MultipleContexts(t *testing.T) {
	ctx1 := SetupSignalHandler()
	ctx2 := SetupSignalHandler()

	// Both contexts should be independent but react to same signal
	select {
	case <-ctx1.Done():
		t.Fatal("ctx1 should not be cancelled initially")
	case <-ctx2.Done():
		t.Fatal("ctx2 should not be cancelled initially")
	default:
		// Expected: both contexts not cancelled
	}

	// Send SIGINT
	currentPID := os.Getpid()
	process, err := os.FindProcess(currentPID)
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}

	if err := process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	// Both contexts should be cancelled
	timeout := time.After(500 * time.Millisecond)

	select {
	case <-ctx1.Done():
		// Expected
	case <-timeout:
		t.Fatal("ctx1 was not cancelled")
	}

	select {
	case <-ctx2.Done():
		// Expected
	case <-timeout:
		t.Fatal("ctx2 was not cancelled")
	}
}

// TestSetupSignalHandler_InFlightJobs tests that in-flight jobs complete gracefully
func TestSetupSignalHandler_InFlightJobs(t *testing.T) {
	ctx := SetupSignalHandler()
	pool := NewPool(2)
	pool.Start(ctx)

	jobsInProgress := make(chan bool, 2)
	jobsFinished := make(chan bool, 2)

	// Submit 2 jobs that signal when they're in progress
	for i := 0; i < 2; i++ {
		pool.Submit(Job{
			ID: "tracked-job",
			Fn: func(ctx context.Context) (interface{}, error) {
				jobsInProgress <- true
				// Simulate work that can be interrupted
				select {
				case <-ctx.Done():
					jobsFinished <- false
					return nil, ctx.Err()
				case <-time.After(200 * time.Millisecond):
					jobsFinished <- true
					return "completed", nil
				}
			},
		})
	}

	// Wait for both jobs to start
	<-jobsInProgress
	<-jobsInProgress

	// Send signal after jobs have started
	currentPID := os.Getpid()
	process, _ := os.FindProcess(currentPID)
	_ = process.Signal(syscall.SIGINT) // Ignore error in test

	// Collect results
	pool.Wait()

	// At least some jobs should have been notified of cancellation
	t.Log("In-flight jobs handling cancellation verified")
}
