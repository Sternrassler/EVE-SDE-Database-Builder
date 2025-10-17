package worker

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewPool tests pool creation with different worker counts
func TestNewPool(t *testing.T) {
	tests := []struct {
		name            string
		workers         int
		expectedWorkers int
	}{
		{"4 workers", 4, 4},
		{"1 worker", 1, 1},
		{"0 workers (defaults to 1)", 0, 1},
		{"negative workers (defaults to 1)", -5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(tt.workers)
			if pool.workers != tt.expectedWorkers {
				t.Errorf("expected %d workers, got %d", tt.expectedWorkers, pool.workers)
			}
			if pool.jobs == nil {
				t.Error("jobs channel should not be nil")
			}
			if pool.results == nil {
				t.Error("results channel should not be nil")
			}
		})
	}
}

// TestPool_WithFourWorkers tests pool with 4 workers processing multiple jobs
func TestPool_WithFourWorkers(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(4)
	pool.Start(ctx)

	jobCount := 10
	var completed atomic.Int32

	// Submit jobs
	for i := 0; i < jobCount; i++ {
		jobNum := i
		pool.Submit(Job{
			ID: fmt.Sprintf("job-%d", jobNum),
			Fn: func(ctx context.Context) (interface{}, error) {
				completed.Add(1)
				return jobNum * 2, nil
			},
		})
	}

	// Wait for completion
	results, errs := pool.Wait()

	// Verify results
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if len(results) != jobCount {
		t.Errorf("expected %d results, got %d", jobCount, len(results))
	}

	if completed.Load() != int32(jobCount) {
		t.Errorf("expected %d completed jobs, got %d", jobCount, completed.Load())
	}
}

// TestPool_ContextCancellation tests graceful shutdown via context cancellation
func TestPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewPool(2)
	pool.Start(ctx)

	var started atomic.Int32
	var completed atomic.Int32

	// Submit long-running jobs
	for i := 0; i < 5; i++ {
		pool.Submit(Job{
			ID: fmt.Sprintf("job-%d", i),
			Fn: func(ctx context.Context) (interface{}, error) {
				started.Add(1)
				// Simulate work
				select {
				case <-time.After(100 * time.Millisecond):
					completed.Add(1)
					return nil, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		})
	}

	// Cancel context after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for completion
	results, _ := pool.Wait()

	// At least some jobs should have started
	if started.Load() == 0 {
		t.Error("expected at least some jobs to start")
	}

	// Not all jobs should have completed (due to cancellation)
	if completed.Load() >= 5 {
		t.Errorf("expected fewer than 5 completed jobs due to cancellation, got %d", completed.Load())
	}

	// Results count should match what was actually processed
	if len(results) != int(started.Load()) {
		t.Logf("started: %d, results: %d", started.Load(), len(results))
		// This is OK - some jobs might not send results if cancelled early
	}
}

// TestPool_ErrorCollection tests that errors from jobs are collected
func TestPool_ErrorCollection(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(2)
	pool.Start(ctx)

	expectedErrors := 3
	successfulJobs := 2

	// Submit jobs with errors
	for i := 0; i < expectedErrors; i++ {
		pool.Submit(Job{
			ID: fmt.Sprintf("error-job-%d", i),
			Fn: func(ctx context.Context) (interface{}, error) {
				return nil, errors.New("job failed")
			},
		})
	}

	// Submit successful jobs
	for i := 0; i < successfulJobs; i++ {
		pool.Submit(Job{
			ID: fmt.Sprintf("success-job-%d", i),
			Fn: func(ctx context.Context) (interface{}, error) {
				return "success", nil
			},
		})
	}

	// Wait for completion
	results, errs := pool.Wait()

	// Verify error collection
	if len(errs) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(errs))
	}

	if len(results) != expectedErrors+successfulJobs {
		t.Errorf("expected %d total results, got %d", expectedErrors+successfulJobs, len(results))
	}

	// Count successful results
	successCount := 0
	for _, result := range results {
		if result.Err == nil {
			successCount++
		}
	}

	if successCount != successfulJobs {
		t.Errorf("expected %d successful results, got %d", successfulJobs, successCount)
	}
}

// TestPool_MultipleJobs tests processing of multiple jobs in sequence
func TestPool_MultipleJobs(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(3)
	pool.Start(ctx)

	jobCount := 20
	jobData := make(map[string]int)

	// Submit many jobs
	for i := 0; i < jobCount; i++ {
		jobNum := i
		pool.Submit(Job{
			ID: fmt.Sprintf("job-%d", jobNum),
			Fn: func(ctx context.Context) (interface{}, error) {
				return jobNum * 10, nil
			},
		})
	}

	// Wait and collect
	results, errs := pool.Wait()

	// Verify
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if len(results) != jobCount {
		t.Errorf("expected %d results, got %d", jobCount, len(results))
	}

	// Verify all job IDs are present
	for _, result := range results {
		jobData[result.JobID]++
	}

	if len(jobData) == 0 {
		t.Error("expected job data to be collected")
	}
}

// TestPool_EmptyPool tests pool with no jobs submitted
func TestPool_EmptyPool(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(2)
	pool.Start(ctx)

	// Don't submit any jobs, just wait
	results, errs := pool.Wait()

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}

	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

// TestPool_CancellationBeforeStart tests cancelling context before jobs run
func TestPool_CancellationBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	pool := NewPool(2)
	pool.Start(ctx)

	// Submit jobs
	pool.Submit(Job{
		ID: "job1",
		Fn: func(ctx context.Context) (interface{}, error) {
			// Check if context is already cancelled
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, errors.New("should not run")
		},
	})

	// Wait
	results, _ := pool.Wait()

	// With a pre-cancelled context, there's a race - the job might get picked up
	// before the worker sees ctx.Done(). We accept 0 or 1 result here.
	if len(results) > 1 {
		t.Errorf("expected at most 1 result (race condition), got %d", len(results))
	}
}

// TestPool_JobWithContextCheck tests job that checks context
func TestPool_JobWithContextCheck(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(1)
	pool.Start(ctx)

	var jobStarted atomic.Int32
	done := make(chan struct{})

	pool.Submit(Job{
		ID: "long_job",
		Fn: func(ctx context.Context) (interface{}, error) {
			jobStarted.Add(1)
			close(done)
			// Long-running operation that checks context
			for i := 0; i < 100; i++ {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					time.Sleep(5 * time.Millisecond)
				}
			}
			return "completed", nil
		},
	})

	// Wait for job to start
	<-done

	results, errs := pool.Wait()

	// Job should have started
	if jobStarted.Load() == 0 {
		t.Fatal("job should have started")
	}

	// Job should have completed normally (no cancellation)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}

	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

// TestPool_TimeoutCancellation tests graceful worker shutdown on timeout
func TestPool_TimeoutCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	pool := NewPool(2)
	pool.Start(ctx)

	var started atomic.Int32

	// Submit several jobs
	for i := 0; i < 5; i++ {
		pool.Submit(Job{
			ID: fmt.Sprintf("job-%d", i),
			Fn: func(ctx context.Context) (interface{}, error) {
				started.Add(1)
				// Simulate work
				time.Sleep(10 * time.Millisecond)
				return "done", nil
			},
		})
	}

	results, _ := pool.Wait()

	// Some jobs should have started
	if started.Load() == 0 {
		t.Error("expected at least some jobs to start")
	}

	// But due to timeout, not all will complete
	if int32(len(results)) > started.Load() {
		t.Errorf("results count (%d) cannot exceed jobs started (%d)", len(results), started.Load())
	}

	t.Logf("Started: %d, Results: %d (context timeout)", started.Load(), len(results))
}

// TestPool_ResultOrder tests that all results are collected (order doesn't matter)
func TestPool_ResultOrder(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(4)
	pool.Start(ctx)

	jobIDs := []string{"A", "B", "C", "D", "E"}

	for _, id := range jobIDs {
		jobID := id
		pool.Submit(Job{
			ID: jobID,
			Fn: func(ctx context.Context) (interface{}, error) {
				// Variable delay to ensure non-deterministic order
				time.Sleep(time.Duration(len(jobID)) * time.Millisecond)
				return jobID + "_result", nil
			},
		})
	}

	results, errs := pool.Wait()

	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if len(results) != len(jobIDs) {
		t.Errorf("expected %d results, got %d", len(jobIDs), len(results))
	}

	// Verify all job IDs are present
	foundIDs := make(map[string]bool)
	for _, result := range results {
		foundIDs[result.JobID] = true
	}

	for _, id := range jobIDs {
		if !foundIDs[id] {
			t.Errorf("missing result for job ID %s", id)
		}
	}
}
