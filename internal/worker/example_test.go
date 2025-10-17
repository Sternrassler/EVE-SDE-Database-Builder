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

// Example_customDataProcessingJob demonstrates a custom data transformation job
func Example_customDataProcessingJob() {
	// Custom job that transforms data
	type DataProcessor struct {
		InputData  []string
		Multiplier int
	}

	processData := func(ctx context.Context, processor *DataProcessor) (interface{}, error) {
		results := make([]string, 0, len(processor.InputData))
		for _, item := range processor.InputData {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				// Transform data
				for i := 0; i < processor.Multiplier; i++ {
					results = append(results, item)
				}
			}
		}
		return results, nil
	}

	ctx := context.Background()
	pool := worker.NewPool(2)
	pool.Start(ctx)

	// Submit custom processing jobs
	processor := &DataProcessor{
		InputData:  []string{"a", "b"},
		Multiplier: 2,
	}

	pool.Submit(worker.Job{
		ID: "data-transform",
		Fn: func(ctx context.Context) (interface{}, error) {
			return processData(ctx, processor)
		},
	})

	results, _ := pool.Wait()
	if len(results) > 0 {
		data := results[0].Data.([]string)
		fmt.Printf("Transformed %d items\n", len(data))
	}
	// Output: Transformed 4 items
}

// Example_customFileValidationJob demonstrates a custom validation job
func Example_customFileValidationJob() {
	// Custom validator job
	type FileValidator struct {
		FilePath string
		MinSize  int
	}

	validateFile := func(ctx context.Context, validator *FileValidator) (interface{}, error) {
		// Simulate file validation
		fileSize := len(validator.FilePath) * 10 // Mock size calculation

		if fileSize < validator.MinSize {
			return nil, fmt.Errorf("file %s too small: %d < %d",
				validator.FilePath, fileSize, validator.MinSize)
		}

		return map[string]interface{}{
			"file":  validator.FilePath,
			"size":  fileSize,
			"valid": true,
		}, nil
	}

	ctx := context.Background()
	pool := worker.NewPool(2)
	pool.Start(ctx)

	// Submit validation jobs for multiple files
	// data.txt (8*10=80), config.yaml (11*10=110), x.log (5*10=50)
	files := []string{"data.txt", "config.yaml", "x.log"}
	for _, file := range files {
		filePath := file
		pool.Submit(worker.Job{
			ID: fmt.Sprintf("validate-%s", filePath),
			Fn: func(ctx context.Context) (interface{}, error) {
				validator := &FileValidator{
					FilePath: filePath,
					MinSize:  60, // Only x.log (50) will fail
				}
				return validateFile(ctx, validator)
			},
		})
	}

	results, errors := pool.Wait()
	fmt.Printf("Validated: %d files, %d errors\n", len(results), len(errors))
	// Output: Validated: 3 files, 1 errors
}

// Example_customBatchProcessingJob demonstrates batch processing pattern
func Example_customBatchProcessingJob() {
	// Custom batch processor
	type BatchProcessor struct {
		BatchID int
		Items   []int
	}

	processBatch := func(ctx context.Context, processor *BatchProcessor) (interface{}, error) {
		sum := 0
		for _, item := range processor.Items {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				sum += item
			}
		}

		return map[string]interface{}{
			"batch_id": processor.BatchID,
			"count":    len(processor.Items),
			"sum":      sum,
		}, nil
	}

	ctx := context.Background()
	pool := worker.NewPool(3)
	pool.Start(ctx)

	// Create batches
	batches := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	// Submit batch processing jobs
	for i, batch := range batches {
		batchID := i
		batchData := batch
		pool.Submit(worker.Job{
			ID: fmt.Sprintf("batch-%d", batchID),
			Fn: func(ctx context.Context) (interface{}, error) {
				processor := &BatchProcessor{
					BatchID: batchID,
					Items:   batchData,
				}
				return processBatch(ctx, processor)
			},
		})
	}

	results, _ := pool.Wait()
	fmt.Printf("Processed %d batches\n", len(results))
	// Output: Processed 3 batches
}

// Example_customAPICallJob demonstrates parallel API call pattern
func Example_customAPICallJob() {
	// Custom API caller
	type APIRequest struct {
		URL    string
		Method string
	}

	callAPI := func(ctx context.Context, req *APIRequest) (interface{}, error) {
		// Simulate API call
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Mock response
			return map[string]interface{}{
				"url":    req.URL,
				"method": req.Method,
				"status": 200,
			}, nil
		}
	}

	ctx := context.Background()
	pool := worker.NewPool(4)
	pool.Start(ctx)

	// Submit parallel API calls
	endpoints := []string{"/users", "/products", "/orders", "/analytics"}
	for _, endpoint := range endpoints {
		ep := endpoint
		pool.Submit(worker.Job{
			ID: fmt.Sprintf("api%s", ep),
			Fn: func(ctx context.Context) (interface{}, error) {
				req := &APIRequest{
					URL:    "https://api.example.com" + ep,
					Method: "GET",
				}
				return callAPI(ctx, req)
			},
		})
	}

	results, _ := pool.Wait()
	fmt.Printf("Completed %d API calls\n", len(results))
	// Output: Completed 4 API calls
}
