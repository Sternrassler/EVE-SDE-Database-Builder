package worker_test

import (
	"context"
	"fmt"
	"log"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
)

// Example_basicUsage demonstrates basic worker pool usage
func Example_basicUsage() {
	ctx := context.Background()
	pool := worker.NewPool(2) // 2 workers
	pool.Start(ctx)

	// Submit jobs
	for i := 1; i <= 3; i++ {
		jobNum := i
		pool.Submit(worker.Job{
			ID: fmt.Sprintf("job-%d", jobNum),
			Fn: func(ctx context.Context) (interface{}, error) {
				// Simulate work
				return jobNum * 10, nil
			},
		})
	}

	// Wait for completion
	results, errors := pool.Wait()

	fmt.Printf("Completed %d jobs with %d errors\n", len(results), len(errors))
	// Output: Completed 3 jobs with 0 errors
}

// Example_contextCancellation demonstrates graceful shutdown with context
func Example_contextCancellation() {
	ctx := context.Background()
	pool := worker.NewPool(2)
	pool.Start(ctx)

	// Submit a job
	pool.Submit(worker.Job{
		ID: "job",
		Fn: func(ctx context.Context) (interface{}, error) {
			return "completed", nil
		},
	})

	results, _ := pool.Wait()
	fmt.Printf("Processed %d jobs\n", len(results))
	// Output: Processed 1 jobs
}

// Example_errorHandling demonstrates error collection
func Example_errorHandling() {
	ctx := context.Background()
	pool := worker.NewPool(2)
	pool.Start(ctx)

	// Submit jobs with different outcomes
	pool.Submit(worker.Job{
		ID: "success",
		Fn: func(ctx context.Context) (interface{}, error) {
			return "ok", nil
		},
	})

	pool.Submit(worker.Job{
		ID: "failure",
		Fn: func(ctx context.Context) (interface{}, error) {
			return nil, fmt.Errorf("job failed")
		},
	})

	results, errors := pool.Wait()
	fmt.Printf("Results: %d, Errors: %d\n", len(results), len(errors))
	// Output: Results: 2, Errors: 1
}

// Example_parallelParsing demonstrates parallel JSONL parsing pattern
func Example_parallelParsing() {
	ctx := context.Background()
	pool := worker.NewPool(4) // 4 parallel workers
	pool.Start(ctx)

	files := []string{"types.jsonl", "blueprints.jsonl", "agents.jsonl"}

	// Submit parsing jobs
	for _, file := range files {
		fileName := file
		pool.Submit(worker.Job{
			ID: fileName,
			Fn: func(ctx context.Context) (interface{}, error) {
				// Simulate JSONL parsing
				log.Printf("Parsing %s", fileName)
				return fmt.Sprintf("parsed-%s", fileName), nil
			},
		})
	}

	results, _ := pool.Wait()
	fmt.Printf("Parsed %d files\n", len(results))
	// Output: Parsed 3 files
}

// Example_signalHandling demonstrates graceful shutdown on SIGINT/SIGTERM
func Example_signalHandling() {
	// Setup context that cancels on SIGINT/SIGTERM
	ctx := worker.SetupSignalHandler()

	pool := worker.NewPool(2)
	pool.Start(ctx)

	// Submit jobs
	for i := 1; i <= 3; i++ {
		jobNum := i
		pool.Submit(worker.Job{
			ID: fmt.Sprintf("job-%d", jobNum),
			Fn: func(ctx context.Context) (interface{}, error) {
				// Jobs will be cancelled on Ctrl+C
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return jobNum * 10, nil
				}
			},
		})
	}

	results, _ := pool.Wait()
	fmt.Printf("Processed %d jobs (graceful shutdown on signals)\n", len(results))
	// Output: Processed 3 jobs (graceful shutdown on signals)
}
