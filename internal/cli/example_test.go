package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/cli"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
)

// Example_progressBarBasicUsage demonstrates basic progress bar usage
func Example_progressBarBasicUsage() {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a progress bar for 100 files
	pb := cli.NewProgressBar(cli.ProgressBarConfig{
		Total:       100,
		Description: "Importing files",
		Width:       40,
		ShowSpinner: false,
		Output:      &buf, // Redirect output to buffer
	})

	// Create a progress tracker
	tracker := worker.NewProgressTracker(100)
	tracker.SetTotalRows(10000)

	// Simulate import in background
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		// Start progress bar updates
		pb.Start(ctx, tracker)
		close(done)
	}()

	// Simulate processing
	for i := 0; i < 100; i++ {
		tracker.Update(1, 100)
		time.Sleep(5 * time.Millisecond)
	}

	// Stop progress bar
	cancel()
	<-done
	pb.Finish()

	fmt.Println("Import completed")
	// Output: Import completed
}

// Example_progressBarWithSpinner demonstrates progress bar with file spinner
func Example_progressBarWithSpinner() {
	var buf bytes.Buffer

	// Create a progress bar with spinner enabled
	pb := cli.NewProgressBar(cli.ProgressBarConfig{
		Total:       10,
		Description: "Processing files",
		ShowSpinner: true,
		Output:      &buf,
	})

	// Create tracker
	tracker := worker.NewProgressTracker(10)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pb.Start(ctx, tracker)
		close(done)
	}()

	// Process files with spinner
	files := []string{"file1.jsonl", "file2.jsonl", "file3.jsonl"}
	for _, file := range files {
		pb.StartSpinner(file)
		time.Sleep(30 * time.Millisecond)
		tracker.Update(1, 100)
		pb.StopSpinner()
	}

	cancel()
	<-done
	pb.Finish()

	fmt.Println("Processing completed")
	// Output: Processing completed
}

// Example_progressBarLiveMetrics demonstrates live metrics display
func Example_progressBarLiveMetrics() {
	var buf bytes.Buffer

	// Create progress bar with live metrics
	pb := cli.NewProgressBar(cli.ProgressBarConfig{
		Total:       50,
		Description: "Importing with metrics",
		UpdateRate:  50 * time.Millisecond,
		Output:      &buf,
	})

	tracker := worker.NewProgressTracker(50)
	tracker.SetTotalRows(50000)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		pb.Start(ctx, tracker)
		close(done)
	}()

	// Simulate varying processing speeds
	for i := 0; i < 50; i++ {
		tracker.Update(1, 1000)
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	<-done
	pb.Finish()

	// Get final metrics
	progress := tracker.GetProgressDetailed()
	fmt.Printf("Processed %d files\n", progress.ParsedFiles)
	fmt.Printf("Inserted %d rows\n", progress.InsertedRows)

	// Output:
	// Processed 50 files
	// Inserted 50000 rows
}
