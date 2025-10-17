package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestPool_RaceConditions tests the worker pool for race conditions with 1000+ jobs
// This test should be run with the -race flag: go test -race ./internal/worker
// To handle the results channel buffer limit (100), we process jobs in batches
func TestPool_RaceConditions(t *testing.T) {
	ctx := context.Background()

	// Process 1000 jobs total across multiple pool instances
	totalJobs := 1000
	batchSize := 90 // Stay under the 100 results buffer limit
	batches := (totalJobs + batchSize - 1) / batchSize

	var totalCompleted atomic.Int32
	var allResults []Result
	var allErrors []error

	for batch := 0; batch < batches; batch++ {
		// Calculate jobs for this batch
		startJob := batch * batchSize
		endJob := startJob + batchSize
		if endJob > totalJobs {
			endJob = totalJobs
		}
		jobsInBatch := endJob - startJob

		pool := NewPool(10)
		pool.Start(ctx)

		// Submit jobs for this batch
		for i := startJob; i < endJob; i++ {
			jobNum := i
			pool.Submit(Job{
				ID: fmt.Sprintf("job-%d", jobNum),
				Fn: func(ctx context.Context) (interface{}, error) {
					totalCompleted.Add(1)
					return jobNum, nil
				},
			})
		}

		// Wait for batch to complete
		results, errs := pool.Wait()
		allResults = append(allResults, results...)
		allErrors = append(allErrors, errs...)

		// Verify batch results
		if len(results) != jobsInBatch {
			t.Errorf("batch %d: expected %d results, got %d", batch, jobsInBatch, len(results))
		}
	}

	// Verify all jobs completed successfully
	if len(allErrors) != 0 {
		t.Errorf("expected no errors, got %d", len(allErrors))
	}

	if len(allResults) != totalJobs {
		t.Errorf("expected %d total results, got %d", totalJobs, len(allResults))
	}

	if totalCompleted.Load() != int32(totalJobs) {
		t.Errorf("expected %d completed jobs, got %d", totalJobs, totalCompleted.Load())
	}

	t.Logf("Race condition test completed: %d jobs processed successfully across %d batches",
		totalCompleted.Load(), batches)
}

// TestPool_RaceConditions_StressTest is a more intensive stress test with 5000+ jobs
// Uses concurrent submission and processing to stress test race conditions
func TestPool_RaceConditions_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ctx := context.Background()

	// Process 5000 jobs total across multiple pool instances with concurrent workers
	totalJobs := 5000
	batchSize := 90 // Stay under results buffer limit
	batches := (totalJobs + batchSize - 1) / batchSize

	var totalCompleted atomic.Int32
	var mu sync.Mutex
	var allResults []Result
	var allErrors []error

	// Process batches concurrently using a worker pool pattern
	var wg sync.WaitGroup
	batchChan := make(chan int, batches)

	// Start batch processors
	concurrentProcessors := 3
	for p := 0; p < concurrentProcessors; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				// Calculate jobs for this batch
				startJob := batch * batchSize
				endJob := startJob + batchSize
				if endJob > totalJobs {
					endJob = totalJobs
				}

				pool := NewPool(20)
				pool.Start(ctx)

				// Submit jobs for this batch
				for i := startJob; i < endJob; i++ {
					jobNum := i
					pool.Submit(Job{
						ID: fmt.Sprintf("stress-job-%d", jobNum),
						Fn: func(ctx context.Context) (interface{}, error) {
							totalCompleted.Add(1)
							// Simulate varying work durations
							if jobNum%100 == 0 {
								time.Sleep(time.Millisecond)
							}
							return jobNum, nil
						},
					})
				}

				// Wait for batch to complete
				results, errs := pool.Wait()

				// Collect results thread-safely
				mu.Lock()
				allResults = append(allResults, results...)
				allErrors = append(allErrors, errs...)
				mu.Unlock()
			}
		}()
	}

	// Queue all batches
	for batch := 0; batch < batches; batch++ {
		batchChan <- batch
	}
	close(batchChan)

	// Wait for all batches to complete
	wg.Wait()

	// Verify results
	if len(allErrors) != 0 {
		t.Errorf("expected no errors, got %d", len(allErrors))
	}

	if len(allResults) != totalJobs {
		t.Errorf("expected %d total results, got %d", totalJobs, len(allResults))
	}

	if totalCompleted.Load() != int32(totalJobs) {
		t.Errorf("expected %d completed jobs, got %d", totalJobs, totalCompleted.Load())
	}

	t.Logf("Stress test completed: %d jobs processed successfully across %d batches with %d concurrent processors",
		totalCompleted.Load(), batches, concurrentProcessors)
}

// TestPool_RaceConditions_ConcurrentReads tests concurrent reads during job processing
func TestPool_RaceConditions_ConcurrentReads(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(10)
	pool.Start(ctx)

	jobCount := 90 // Stay within results buffer limit
	sharedCounter := &struct {
		sync.RWMutex
		count int
	}{}

	var wg sync.WaitGroup

	// Concurrently read from shared counter while jobs are being submitted
	readers := 5
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				sharedCounter.RLock()
				_ = sharedCounter.count
				sharedCounter.RUnlock()
				time.Sleep(time.Microsecond * 50)
			}
		}()
	}

	// Submit jobs that write to shared counter
	for i := 0; i < jobCount; i++ {
		pool.Submit(Job{
			ID: fmt.Sprintf("job-%d", i),
			Fn: func(ctx context.Context) (interface{}, error) {
				sharedCounter.Lock()
				sharedCounter.count++
				sharedCounter.Unlock()
				return nil, nil
			},
		})
	}

	// Wait for all jobs
	results, errs := pool.Wait()

	// Wait for readers to finish
	wg.Wait()

	// Verify
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if len(results) != jobCount {
		t.Errorf("expected %d results, got %d", jobCount, len(results))
	}

	sharedCounter.RLock()
	finalCount := sharedCounter.count
	sharedCounter.RUnlock()

	if finalCount != jobCount {
		t.Errorf("expected counter to be %d, got %d", jobCount, finalCount)
	}
}

// TestPool_RaceConditions_ChannelOperations tests concurrent channel operations
func TestPool_RaceConditions_ChannelOperations(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(15)
	pool.Start(ctx)

	jobCount := 90 // Stay within results buffer limit
	resultChan := make(chan int, jobCount)
	var wg sync.WaitGroup

	// Concurrently read from result channel
	var received atomic.Int32
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range resultChan {
			received.Add(1)
		}
	}()

	// Submit jobs
	for i := 0; i < jobCount; i++ {
		jobNum := i
		pool.Submit(Job{
			ID: fmt.Sprintf("chan-job-%d", i),
			Fn: func(ctx context.Context) (interface{}, error) {
				// Write to result channel
				resultChan <- jobNum
				return jobNum, nil
			},
		})
	}

	// Wait for pool
	results, errs := pool.Wait()
	close(resultChan)

	// Wait for reader
	wg.Wait()

	// Verify
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if len(results) != jobCount {
		t.Errorf("expected %d results, got %d", jobCount, len(results))
	}

	if int(received.Load()) != jobCount {
		t.Errorf("expected to receive %d results, got %d", jobCount, received.Load())
	}
}

// TestPool_RaceConditions_MapAccess tests concurrent map access patterns
func TestPool_RaceConditions_MapAccess(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(10)
	pool.Start(ctx)

	jobCount := 90 // Stay within results buffer limit
	safeMap := &struct {
		sync.RWMutex
		data map[string]int
	}{
		data: make(map[string]int),
	}

	var wg sync.WaitGroup

	// Concurrently read from map
	readers := 5
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				safeMap.RLock()
				_ = len(safeMap.data)
				safeMap.RUnlock()
				time.Sleep(time.Microsecond * 50)
			}
		}()
	}

	// Submit jobs that write to map
	for i := 0; i < jobCount; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		jobNum := i
		pool.Submit(Job{
			ID: jobID,
			Fn: func(ctx context.Context) (interface{}, error) {
				safeMap.Lock()
				safeMap.data[jobID] = jobNum
				safeMap.Unlock()
				return jobNum, nil
			},
		})
	}

	// Wait for pool
	results, errs := pool.Wait()

	// Wait for readers
	wg.Wait()

	// Verify
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if len(results) != jobCount {
		t.Errorf("expected %d results, got %d", jobCount, len(results))
	}

	safeMap.RLock()
	mapSize := len(safeMap.data)
	safeMap.RUnlock()

	if mapSize != jobCount {
		t.Errorf("expected map size to be %d, got %d", jobCount, mapSize)
	}
}

// TestPool_RaceConditions_ContextCancellation tests race conditions during context cancellation
func TestPool_RaceConditions_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewPool(10)
	pool.Start(ctx)

	jobCount := 50 // Use fewer jobs for cancellation test
	var started atomic.Int32
	var completed atomic.Int32

	// Submit jobs
	for i := 0; i < jobCount; i++ {
		pool.Submit(Job{
			ID: fmt.Sprintf("cancel-job-%d", i),
			Fn: func(ctx context.Context) (interface{}, error) {
				started.Add(1)
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

	// Cancel context after a short delay to create race conditions
	time.Sleep(20 * time.Millisecond)
	cancel()

	// Wait for pool
	results, _ := pool.Wait()

	// Verify at least some jobs started
	if started.Load() == 0 {
		t.Error("expected at least some jobs to start")
	}

	// Results should not exceed jobs that started
	if int32(len(results)) > started.Load() {
		t.Errorf("results count (%d) cannot exceed started jobs (%d)", len(results), started.Load())
	}

	t.Logf("Cancellation test: %d started, %d completed, %d results",
		started.Load(), completed.Load(), len(results))
}
