package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkPool_1Worker benchmarks pool performance with 1 worker
// Target: Establish baseline for single-threaded execution
func BenchmarkPool_1Worker(b *testing.B) {
	benchmarkPoolWorkers(b, 1, 100)
}

// BenchmarkPool_2Workers benchmarks pool performance with 2 workers
// Target: Measure scaling with minimal concurrency
func BenchmarkPool_2Workers(b *testing.B) {
	benchmarkPoolWorkers(b, 2, 100)
}

// BenchmarkPool_4Workers benchmarks pool performance with 4 workers
// Target: Test typical multi-core performance
func BenchmarkPool_4Workers(b *testing.B) {
	benchmarkPoolWorkers(b, 4, 100)
}

// BenchmarkPool_8Workers benchmarks pool performance with 8 workers
// Target: Test performance on higher-core systems
func BenchmarkPool_8Workers(b *testing.B) {
	benchmarkPoolWorkers(b, 8, 100)
}

// BenchmarkPool_16Workers benchmarks pool performance with 16 workers
// Target: Test performance at high concurrency levels
func BenchmarkPool_16Workers(b *testing.B) {
	benchmarkPoolWorkers(b, 16, 100)
}

// benchmarkPoolWorkers is a helper function that benchmarks pool performance
// with the specified number of workers and jobs
func benchmarkPoolWorkers(b *testing.B, workers int, jobCount int) {
	ctx := context.Background()

	// Report memory allocations for profiling
	b.ReportAllocs()

	// Reset timer to exclude setup time
	b.ResetTimer()

	// Run benchmark iterations
	for i := 0; i < b.N; i++ {
		pool := NewPool(workers)
		pool.Start(ctx)

		var completed atomic.Int32

		// Submit jobs
		for j := 0; j < jobCount; j++ {
			jobNum := j
			pool.Submit(Job{
				ID: fmt.Sprintf("job-%d", jobNum),
				Fn: func(ctx context.Context) (interface{}, error) {
					// Simulate realistic work (CPU-bound computation)
					sum := 0
					for k := 0; k < 1000; k++ {
						sum += k
					}
					completed.Add(1)
					return sum, nil
				},
			})
		}

		// Wait for completion
		results, errs := pool.Wait()

		// Verify results (prevents compiler optimization)
		if len(errs) != 0 {
			b.Fatalf("expected no errors, got %d", len(errs))
		}
		if len(results) != jobCount {
			b.Fatalf("expected %d results, got %d", jobCount, len(results))
		}
	}

	b.StopTimer()
}

// BenchmarkPool_1Worker_IOBound benchmarks pool with 1 worker on I/O-bound tasks
func BenchmarkPool_1Worker_IOBound(b *testing.B) {
	benchmarkPoolWorkersIOBound(b, 1, 100)
}

// BenchmarkPool_2Workers_IOBound benchmarks pool with 2 workers on I/O-bound tasks
func BenchmarkPool_2Workers_IOBound(b *testing.B) {
	benchmarkPoolWorkersIOBound(b, 2, 100)
}

// BenchmarkPool_4Workers_IOBound benchmarks pool with 4 workers on I/O-bound tasks
func BenchmarkPool_4Workers_IOBound(b *testing.B) {
	benchmarkPoolWorkersIOBound(b, 4, 100)
}

// BenchmarkPool_8Workers_IOBound benchmarks pool with 8 workers on I/O-bound tasks
func BenchmarkPool_8Workers_IOBound(b *testing.B) {
	benchmarkPoolWorkersIOBound(b, 8, 100)
}

// BenchmarkPool_16Workers_IOBound benchmarks pool with 16 workers on I/O-bound tasks
func BenchmarkPool_16Workers_IOBound(b *testing.B) {
	benchmarkPoolWorkersIOBound(b, 16, 100)
}

// benchmarkPoolWorkersIOBound benchmarks pool performance with I/O-bound tasks
// Simulates I/O operations like file reading or database queries
func benchmarkPoolWorkersIOBound(b *testing.B, workers int, jobCount int) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool := NewPool(workers)
		pool.Start(ctx)

		var completed atomic.Int32

		// Submit I/O-bound jobs
		for j := 0; j < jobCount; j++ {
			jobNum := j
			pool.Submit(Job{
				ID: fmt.Sprintf("io-job-%d", jobNum),
				Fn: func(ctx context.Context) (interface{}, error) {
					// Simulate I/O delay (e.g., file read, database query)
					time.Sleep(10 * time.Millisecond)
					completed.Add(1)
					return jobNum, nil
				},
			})
		}

		results, errs := pool.Wait()

		if len(errs) != 0 {
			b.Fatalf("expected no errors, got %d", len(errs))
		}
		if len(results) != jobCount {
			b.Fatalf("expected %d results, got %d", jobCount, len(results))
		}
	}

	b.StopTimer()
}

// BenchmarkPool_1Worker_MixedLoad benchmarks pool with 1 worker on mixed CPU/IO tasks
func BenchmarkPool_1Worker_MixedLoad(b *testing.B) {
	benchmarkPoolWorkersMixedLoad(b, 1, 100)
}

// BenchmarkPool_2Workers_MixedLoad benchmarks pool with 2 workers on mixed CPU/IO tasks
func BenchmarkPool_2Workers_MixedLoad(b *testing.B) {
	benchmarkPoolWorkersMixedLoad(b, 2, 100)
}

// BenchmarkPool_4Workers_MixedLoad benchmarks pool with 4 workers on mixed CPU/IO tasks
func BenchmarkPool_4Workers_MixedLoad(b *testing.B) {
	benchmarkPoolWorkersMixedLoad(b, 4, 100)
}

// BenchmarkPool_8Workers_MixedLoad benchmarks pool with 8 workers on mixed CPU/IO tasks
func BenchmarkPool_8Workers_MixedLoad(b *testing.B) {
	benchmarkPoolWorkersMixedLoad(b, 8, 100)
}

// BenchmarkPool_16Workers_MixedLoad benchmarks pool with 16 workers on mixed CPU/IO tasks
func BenchmarkPool_16Workers_MixedLoad(b *testing.B) {
	benchmarkPoolWorkersMixedLoad(b, 16, 100)
}

// benchmarkPoolWorkersMixedLoad benchmarks pool with mixed CPU and I/O workloads
// Simulates realistic scenarios with both computation and I/O operations
func benchmarkPoolWorkersMixedLoad(b *testing.B, workers int, jobCount int) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool := NewPool(workers)
		pool.Start(ctx)

		var completed atomic.Int32

		// Submit mixed workload jobs
		for j := 0; j < jobCount; j++ {
			jobNum := j
			pool.Submit(Job{
				ID: fmt.Sprintf("mixed-job-%d", jobNum),
				Fn: func(ctx context.Context) (interface{}, error) {
					// CPU-bound work
					sum := 0
					for k := 0; k < 500; k++ {
						sum += k
					}

					// I/O-bound work
					time.Sleep(5 * time.Millisecond)

					completed.Add(1)
					return sum, nil
				},
			})
		}

		results, errs := pool.Wait()

		if len(errs) != 0 {
			b.Fatalf("expected no errors, got %d", len(errs))
		}
		if len(results) != jobCount {
			b.Fatalf("expected %d results, got %d", jobCount, len(results))
		}
	}

	b.StopTimer()
}

// BenchmarkPool_ScalingTest_1Worker tests scaling characteristics with 1 worker
func BenchmarkPool_ScalingTest_1Worker(b *testing.B) {
	benchmarkPoolScaling(b, 1)
}

// BenchmarkPool_ScalingTest_2Workers tests scaling characteristics with 2 workers
func BenchmarkPool_ScalingTest_2Workers(b *testing.B) {
	benchmarkPoolScaling(b, 2)
}

// BenchmarkPool_ScalingTest_4Workers tests scaling characteristics with 4 workers
func BenchmarkPool_ScalingTest_4Workers(b *testing.B) {
	benchmarkPoolScaling(b, 4)
}

// BenchmarkPool_ScalingTest_8Workers tests scaling characteristics with 8 workers
func BenchmarkPool_ScalingTest_8Workers(b *testing.B) {
	benchmarkPoolScaling(b, 8)
}

// BenchmarkPool_ScalingTest_16Workers tests scaling characteristics with 16 workers
func BenchmarkPool_ScalingTest_16Workers(b *testing.B) {
	benchmarkPoolScaling(b, 16)
}

// benchmarkPoolScaling tests pool performance with varying job counts
// to measure scaling characteristics
func benchmarkPoolScaling(b *testing.B, workers int) {
	jobCounts := []int{10, 50, 100, 500, 1000}

	for _, jobCount := range jobCounts {
		b.Run(fmt.Sprintf("%dJobs", jobCount), func(b *testing.B) {
			ctx := context.Background()
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				pool := NewPool(workers)
				pool.Start(ctx)

				for j := 0; j < jobCount; j++ {
					jobNum := j
					pool.Submit(Job{
						ID: fmt.Sprintf("job-%d", jobNum),
						Fn: func(ctx context.Context) (interface{}, error) {
							sum := 0
							for k := 0; k < 100; k++ {
								sum += k
							}
							return sum, nil
						},
					})
				}

				results, errs := pool.Wait()

				if len(errs) != 0 {
					b.Fatalf("expected no errors, got %d", len(errs))
				}
				if len(results) != jobCount {
					b.Fatalf("expected %d results, got %d", jobCount, len(results))
				}
			}

			b.StopTimer()
		})
	}
}
