package worker

import (
	"sync"
	"testing"
	"time"
)

// TestProgressTracker_NewProgressTracker tests tracker creation
func TestProgressTracker_NewProgressTracker(t *testing.T) {
	pt := NewProgressTracker(10)

	if pt.totalFiles != 10 {
		t.Errorf("expected totalFiles=10, got %d", pt.totalFiles)
	}

	if pt.parsedFiles.Load() != 0 {
		t.Errorf("expected parsedFiles=0, got %d", pt.parsedFiles.Load())
	}

	if pt.insertedRows.Load() != 0 {
		t.Errorf("expected insertedRows=0, got %d", pt.insertedRows.Load())
	}

	if pt.totalRows.Load() != 0 {
		t.Errorf("expected totalRows=0, got %d", pt.totalRows.Load())
	}

	if pt.startTime.IsZero() {
		t.Error("startTime should be initialized")
	}
}

// TestProgressTracker_SetTotalRows tests setting total rows
func TestProgressTracker_SetTotalRows(t *testing.T) {
	pt := NewProgressTracker(5)
	pt.SetTotalRows(1000)

	if pt.totalRows.Load() != 1000 {
		t.Errorf("expected totalRows=1000, got %d", pt.totalRows.Load())
	}
}

// TestProgressTracker_Update tests Update method
func TestProgressTracker_Update(t *testing.T) {
	pt := NewProgressTracker(10)

	pt.Update(1, 100)

	if pt.parsedFiles.Load() != 1 {
		t.Errorf("expected parsedFiles=1, got %d", pt.parsedFiles.Load())
	}

	if pt.insertedRows.Load() != 100 {
		t.Errorf("expected insertedRows=100, got %d", pt.insertedRows.Load())
	}

	// Multiple updates
	pt.Update(2, 250)

	if pt.parsedFiles.Load() != 3 {
		t.Errorf("expected parsedFiles=3, got %d", pt.parsedFiles.Load())
	}

	if pt.insertedRows.Load() != 350 {
		t.Errorf("expected insertedRows=350, got %d", pt.insertedRows.Load())
	}
}

// TestProgressTracker_UpdateZero tests Update with zero values
func TestProgressTracker_UpdateZero(t *testing.T) {
	pt := NewProgressTracker(10)

	// Update only files
	pt.Update(1, 0)
	if pt.parsedFiles.Load() != 1 {
		t.Errorf("expected parsedFiles=1, got %d", pt.parsedFiles.Load())
	}
	if pt.insertedRows.Load() != 0 {
		t.Errorf("expected insertedRows=0, got %d", pt.insertedRows.Load())
	}

	// Update only rows
	pt.Update(0, 50)
	if pt.parsedFiles.Load() != 1 {
		t.Errorf("expected parsedFiles=1, got %d", pt.parsedFiles.Load())
	}
	if pt.insertedRows.Load() != 50 {
		t.Errorf("expected insertedRows=50, got %d", pt.insertedRows.Load())
	}
}

// TestProgressTracker_AddInsertedRows tests AddInsertedRows
func TestProgressTracker_AddInsertedRows(t *testing.T) {
	pt := NewProgressTracker(5)

	pt.AddInsertedRows(100)
	if pt.insertedRows.Load() != 100 {
		t.Errorf("expected insertedRows=100, got %d", pt.insertedRows.Load())
	}

	pt.AddInsertedRows(250)
	if pt.insertedRows.Load() != 350 {
		t.Errorf("expected insertedRows=350, got %d", pt.insertedRows.Load())
	}
}

// TestProgressTracker_GetProgressDetailed tests detailed progress calculations
func TestProgressTracker_GetProgressDetailed(t *testing.T) {
	pt := NewProgressTracker(10)
	pt.SetTotalRows(1000)

	// Initial progress (nothing processed yet)
	progress := pt.GetProgressDetailed()
	if progress.TotalFiles != 10 {
		t.Errorf("expected TotalFiles=10, got %d", progress.TotalFiles)
	}
	if progress.TotalRows != 1000 {
		t.Errorf("expected TotalRows=1000, got %d", progress.TotalRows)
	}
	if progress.ParsedFiles != 0 {
		t.Errorf("expected ParsedFiles=0, got %d", progress.ParsedFiles)
	}
	if progress.InsertedRows != 0 {
		t.Errorf("expected InsertedRows=0, got %d", progress.InsertedRows)
	}
	if progress.PercentFiles != 0 {
		t.Errorf("expected PercentFiles=0, got %f", progress.PercentFiles)
	}

	// After processing some data
	pt.Update(5, 500)
	progress = pt.GetProgressDetailed()

	if progress.ParsedFiles != 5 {
		t.Errorf("expected ParsedFiles=5, got %d", progress.ParsedFiles)
	}
	if progress.InsertedRows != 500 {
		t.Errorf("expected InsertedRows=500, got %d", progress.InsertedRows)
	}

	// Check percentage calculations
	if progress.PercentFiles != 50.0 {
		t.Errorf("expected PercentFiles=50.0, got %f", progress.PercentFiles)
	}
	if progress.PercentRows != 50.0 {
		t.Errorf("expected PercentRows=50.0, got %f", progress.PercentRows)
	}

	// Check that elapsed time is set
	if progress.ElapsedTime == 0 {
		t.Error("ElapsedTime should be > 0")
	}
}

// TestProgressTracker_PercentageCalculations tests percentage edge cases
func TestProgressTracker_PercentageCalculations(t *testing.T) {
	tests := []struct {
		name                string
		totalFiles          int64
		totalRows           int64
		parsedFiles         int64
		insertedRows        int64
		expectedPercentFile float64
		expectedPercentRows float64
	}{
		{
			name:                "0% completion",
			totalFiles:          10,
			totalRows:           1000,
			parsedFiles:         0,
			insertedRows:        0,
			expectedPercentFile: 0.0,
			expectedPercentRows: 0.0,
		},
		{
			name:                "50% completion",
			totalFiles:          10,
			totalRows:           1000,
			parsedFiles:         5,
			insertedRows:        500,
			expectedPercentFile: 50.0,
			expectedPercentRows: 50.0,
		},
		{
			name:                "100% completion",
			totalFiles:          10,
			totalRows:           1000,
			parsedFiles:         10,
			insertedRows:        1000,
			expectedPercentFile: 100.0,
			expectedPercentRows: 100.0,
		},
		{
			name:                "No total rows set",
			totalFiles:          10,
			totalRows:           0,
			parsedFiles:         5,
			insertedRows:        500,
			expectedPercentFile: 50.0,
			expectedPercentRows: 0.0, // Can't calculate without total
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := NewProgressTracker(int(tt.totalFiles))
			pt.SetTotalRows(tt.totalRows)
			pt.parsedFiles.Store(tt.parsedFiles)
			pt.insertedRows.Store(tt.insertedRows)

			progress := pt.GetProgressDetailed()

			if progress.PercentFiles != tt.expectedPercentFile {
				t.Errorf("expected PercentFiles=%f, got %f",
					tt.expectedPercentFile, progress.PercentFiles)
			}

			if progress.PercentRows != tt.expectedPercentRows {
				t.Errorf("expected PercentRows=%f, got %f",
					tt.expectedPercentRows, progress.PercentRows)
			}
		})
	}
}

// TestProgressTracker_ETA tests ETA calculation
func TestProgressTracker_ETA(t *testing.T) {
	pt := NewProgressTracker(10)
	pt.SetTotalRows(1000)

	// Simulate some time passing
	pt.startTime = time.Now().Add(-2 * time.Second)

	// Simulate 50% completion
	pt.Update(5, 500)

	progress := pt.GetProgressDetailed()

	// ETA should be approximately equal to elapsed time (50% done, 50% remaining)
	// With some tolerance for timing variance
	if progress.ETA == 0 {
		t.Error("ETA should be > 0 when work is in progress")
	}

	// Check rows per second is calculated
	if progress.RowsPerSecond == 0 {
		t.Error("RowsPerSecond should be > 0")
	}

	// RowsPerSecond should be approximately 500/2 = 250
	expectedRate := 250.0
	tolerance := 50.0
	if progress.RowsPerSecond < expectedRate-tolerance || progress.RowsPerSecond > expectedRate+tolerance {
		t.Logf("RowsPerSecond=%f (expected ~%f with tolerance %f)",
			progress.RowsPerSecond, expectedRate, tolerance)
	}
}

// TestProgressTracker_ETAFallback tests ETA calculation based on files when rows unknown
func TestProgressTracker_ETAFallback(t *testing.T) {
	pt := NewProgressTracker(10)
	// Don't set totalRows - should fall back to file-based ETA

	pt.startTime = time.Now().Add(-2 * time.Second)
	pt.Update(5, 0) // 5 files parsed, but no row tracking

	progress := pt.GetProgressDetailed()

	// ETA should still be calculated based on files
	if progress.ETA == 0 {
		t.Error("ETA should be > 0 even without row tracking")
	}
}

// TestProgressTracker_ConcurrentUpdates tests thread-safety
func TestProgressTracker_ConcurrentUpdates(t *testing.T) {
	pt := NewProgressTracker(100)
	pt.SetTotalRows(10000)

	var wg sync.WaitGroup
	workers := 10
	updatesPerWorker := 100

	// Spawn concurrent workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < updatesPerWorker; j++ {
				pt.Update(1, 10)
			}
		}()
	}

	wg.Wait()

	// Verify totals
	expectedFiles := int64(workers * updatesPerWorker)
	expectedRows := int64(workers * updatesPerWorker * 10)

	if pt.parsedFiles.Load() != expectedFiles {
		t.Errorf("expected parsedFiles=%d, got %d", expectedFiles, pt.parsedFiles.Load())
	}

	if pt.insertedRows.Load() != expectedRows {
		t.Errorf("expected insertedRows=%d, got %d", expectedRows, pt.insertedRows.Load())
	}
}

// TestProgressTracker_ConcurrentGetProgress tests concurrent GetProgressDetailed calls
func TestProgressTracker_ConcurrentGetProgress(t *testing.T) {
	pt := NewProgressTracker(100)
	pt.SetTotalRows(10000)

	var wg sync.WaitGroup

	// Concurrent updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			pt.Update(1, 100)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Concurrent reads (should not panic or corrupt data)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				progress := pt.GetProgressDetailed()
				// Basic sanity checks
				if progress.ParsedFiles > progress.TotalFiles {
					t.Errorf("ParsedFiles (%d) > TotalFiles (%d)",
						progress.ParsedFiles, progress.TotalFiles)
				}
				if progress.TotalRows > 0 && progress.InsertedRows > progress.TotalRows {
					t.Errorf("InsertedRows (%d) > TotalRows (%d)",
						progress.InsertedRows, progress.TotalRows)
				}
			}
		}()
	}

	wg.Wait()
}

// TestProgressTracker_ZeroTotalFiles tests behavior with zero total files
func TestProgressTracker_ZeroTotalFiles(t *testing.T) {
	pt := NewProgressTracker(0)
	pt.Update(1, 100)

	progress := pt.GetProgressDetailed()

	// Should not panic, percentages should be 0 (no division by zero)
	if progress.PercentFiles != 0 {
		t.Errorf("expected PercentFiles=0 with zero totalFiles, got %f", progress.PercentFiles)
	}
}

// TestProgressTracker_ElapsedTime tests elapsed time tracking
func TestProgressTracker_ElapsedTime(t *testing.T) {
	pt := NewProgressTracker(10)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	progress := pt.GetProgressDetailed()

	if progress.ElapsedTime < 100*time.Millisecond {
		t.Errorf("expected ElapsedTime >= 100ms, got %v", progress.ElapsedTime)
	}

	if progress.ElapsedTime > 200*time.Millisecond {
		t.Errorf("expected ElapsedTime < 200ms (reasonable bound), got %v", progress.ElapsedTime)
	}
}

// TestProgressTracker_RowsPerSecond tests throughput calculation
func TestProgressTracker_RowsPerSecond(t *testing.T) {
	pt := NewProgressTracker(10)
	pt.SetTotalRows(1000)

	// Set start time in the past
	pt.startTime = time.Now().Add(-1 * time.Second)

	// Add 100 rows
	pt.AddInsertedRows(100)

	progress := pt.GetProgressDetailed()

	// Should be approximately 100 rows/second
	// With some tolerance for timing variance
	expectedRate := 100.0
	tolerance := 20.0

	if progress.RowsPerSecond < expectedRate-tolerance || progress.RowsPerSecond > expectedRate+tolerance {
		t.Logf("RowsPerSecond=%f (expected ~%f, tolerance %f)",
			progress.RowsPerSecond, expectedRate, tolerance)
	}

	if progress.RowsPerSecond <= 0 {
		t.Error("RowsPerSecond should be > 0")
	}
}

// TestProgressTracker_LegacyCompatibility tests backward compatibility with GetProgress
func TestProgressTracker_LegacyCompatibility(t *testing.T) {
	pt := NewProgressTracker(10)

	pt.IncrementParsed()
	pt.IncrementParsed()
	pt.IncrementFailed()

	parsed, inserted, failed, total := pt.GetProgress()

	if parsed != 2 {
		t.Errorf("expected parsed=2, got %d", parsed)
	}
	if inserted != 1 { // inserted = parsed - failed = 2 - 1 = 1
		t.Errorf("expected inserted=1 (parsed - failed), got %d", inserted)
	}
	if failed != 1 {
		t.Errorf("expected failed=1, got %d", failed)
	}
	if total != 10 {
		t.Errorf("expected total=10, got %d", total)
	}
}

// TestProgressTracker_FailedTracking tests failed file tracking
func TestProgressTracker_FailedTracking(t *testing.T) {
	pt := NewProgressTracker(10)

	pt.IncrementParsed()
	pt.IncrementParsed()
	pt.IncrementFailed()
	pt.IncrementParsed()

	progress := pt.GetProgressDetailed()

	if progress.ParsedFiles != 3 {
		t.Errorf("expected ParsedFiles=3, got %d", progress.ParsedFiles)
	}
	if progress.FailedFiles != 1 {
		t.Errorf("expected FailedFiles=1, got %d", progress.FailedFiles)
	}
	if progress.InsertedFiles != 2 { // ParsedFiles - FailedFiles = 3 - 1 = 2
		t.Errorf("expected InsertedFiles=2, got %d", progress.InsertedFiles)
	}
}

// BenchmarkProgressTracker_Update benchmarks Update performance
func BenchmarkProgressTracker_Update(b *testing.B) {
	pt := NewProgressTracker(1000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pt.Update(1, 100)
	}
}

// BenchmarkProgressTracker_GetProgressDetailed benchmarks GetProgressDetailed performance
func BenchmarkProgressTracker_GetProgressDetailed(b *testing.B) {
	pt := NewProgressTracker(1000)
	pt.SetTotalRows(100000)
	pt.Update(500, 50000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pt.GetProgressDetailed()
	}
}

// BenchmarkProgressTracker_ConcurrentUpdates benchmarks concurrent updates
func BenchmarkProgressTracker_ConcurrentUpdates(b *testing.B) {
	pt := NewProgressTracker(1000000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pt.Update(1, 100)
		}
	})
}
