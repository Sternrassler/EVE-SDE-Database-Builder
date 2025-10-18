package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
)

// syncedWriter wraps an io.Writer with a mutex for thread-safe concurrent writes
type syncedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (sw *syncedWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

func newSyncedWriter(w io.Writer) *syncedWriter {
	return &syncedWriter{w: w}
}

// TestNewProgressBar testet die Erstellung einer neuen ProgressBar
func TestNewProgressBar(t *testing.T) {
	config := ProgressBarConfig{
		Total:       10,
		Description: "Test Progress",
		Width:       40,
		UpdateRate:  100 * time.Millisecond,
		ShowSpinner: true,
	}

	pb := NewProgressBar(config)

	if pb == nil {
		t.Fatal("expected ProgressBar, got nil")
	}

	if pb.bar == nil {
		t.Error("expected bar to be initialized")
	}

	if pb.spinner == nil {
		t.Error("expected spinner to be initialized when ShowSpinner is true")
	}

	if pb.updateRate != 100*time.Millisecond {
		t.Errorf("expected updateRate 100ms, got %v", pb.updateRate)
	}
}

// TestNewProgressBar_Defaults testet Default-Werte
func TestNewProgressBar_Defaults(t *testing.T) {
	config := ProgressBarConfig{
		Total: 5,
	}

	pb := NewProgressBar(config)

	if pb.updateRate != 100*time.Millisecond {
		t.Errorf("expected default updateRate 100ms, got %v", pb.updateRate)
	}

	if pb.output == nil {
		t.Error("expected output to be set to default (os.Stdout)")
	}
}

// TestProgressBar_Start testet die automatische Aktualisierung
func TestProgressBar_Start(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressBarConfig{
		Total:       10,
		Description: "Testing",
		UpdateRate:  50 * time.Millisecond,
		ShowSpinner: false,
		Output:      &buf,
	}

	pb := NewProgressBar(config)
	tracker := worker.NewProgressTracker(10)
	tracker.SetTotalRows(1000)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Progress Bar in Goroutine
	done := make(chan struct{})
	go func() {
		pb.Start(ctx, tracker)
		close(done)
	}()

	// Simuliere Fortschritt
	time.Sleep(100 * time.Millisecond)
	tracker.Update(3, 300)

	time.Sleep(100 * time.Millisecond)
	tracker.Update(2, 200)

	// Stop
	cancel()
	<-done // Wait for goroutine to finish

	// Prüfe, dass Output generiert wurde
	output := buf.String()
	if output == "" {
		t.Error("expected some output from progress bar")
	}
}

// TestProgressBar_Finish testet das Abschließen der ProgressBar
func TestProgressBar_Finish(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressBarConfig{
		Total:  5,
		Output: &buf,
	}

	pb := NewProgressBar(config)
	pb.Finish()

	// Sollte mindestens eine neue Zeile ausgeben
	output := buf.String()
	if !strings.Contains(output, "\n") {
		t.Error("expected newline in output after Finish()")
	}
}

// TestFormatDuration testet die Formatierung von Zeitdauern
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Seconds only",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "Minutes and seconds",
			duration: 2*time.Minute + 30*time.Second,
			expected: "2m 30s",
		},
		{
			name:     "Hours and minutes",
			duration: 2*time.Hour + 15*time.Minute,
			expected: "2h 15m",
		},
		{
			name:     "Exactly 1 minute",
			duration: 1 * time.Minute,
			expected: "1m 0s",
		},
		{
			name:     "Less than a second",
			duration: 500 * time.Millisecond,
			expected: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// TestSpinner_StartStop testet Spinner Start/Stop
func TestSpinner_StartStop(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner(&buf)

	if spinner == nil {
		t.Fatal("expected Spinner, got nil")
	}

	// Start
	spinner.Start("Processing file.jsonl")
	time.Sleep(200 * time.Millisecond)

	// Stop
	spinner.Stop()

	// Prüfe, dass Output generiert wurde
	output := buf.String()
	if output == "" {
		t.Error("expected some output from spinner")
	}

	// Spinner sollte nicht mehr aktiv sein
	if spinner.active {
		t.Error("expected spinner to be inactive after Stop()")
	}
}

// TestSpinner_Frames testet, dass Spinner verschiedene Frames anzeigt
func TestSpinner_Frames(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner(&buf)

	spinner.Start("Testing")

	// Warte auf mehrere Frame-Updates
	time.Sleep(300 * time.Millisecond)

	spinner.Stop()

	output := buf.String()

	// Sollte mehrere Spinner-Symbole enthalten
	containsSpinnerChars := false
	for _, frame := range spinner.frames {
		if strings.Contains(output, frame) {
			containsSpinnerChars = true
			break
		}
	}

	if !containsSpinnerChars {
		t.Error("expected output to contain spinner characters")
	}
}

// TestProgressBar_WithSpinner testet ProgressBar mit Spinner
func TestProgressBar_WithSpinner(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressBarConfig{
		Total:       5,
		ShowSpinner: true,
		Output:      &buf,
	}

	pb := NewProgressBar(config)

	if pb.spinner == nil {
		t.Fatal("expected spinner to be initialized")
	}

	// Start/Stop Spinner
	pb.StartSpinner("test-file.jsonl")
	time.Sleep(100 * time.Millisecond)
	pb.StopSpinner()

	// Sollte Output von Spinner haben
	output := buf.String()
	if output == "" {
		t.Error("expected output from spinner")
	}
}

// TestProgressBar_WithoutSpinner testet ProgressBar ohne Spinner
func TestProgressBar_WithoutSpinner(t *testing.T) {
	config := ProgressBarConfig{
		Total:       5,
		ShowSpinner: false,
	}

	pb := NewProgressBar(config)

	if pb.spinner != nil {
		t.Error("expected spinner to be nil when ShowSpinner is false")
	}

	// Sollte nicht paniken, wenn Spinner-Methoden aufgerufen werden
	pb.StartSpinner("test.jsonl")
	pb.StopSpinner()
}

// TestProgressBar_UpdateDescription testet die Beschreibungsaktualisierung
func TestProgressBar_UpdateDescription(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressBarConfig{
		Total:  10,
		Output: &buf,
	}

	pb := NewProgressBar(config)

	// Simuliere Progress-Info
	progress := worker.Progress{
		ParsedFiles:   5,
		TotalFiles:    10,
		InsertedRows:  5000,
		RowsPerSecond: 250.5,
		ETA:           2 * time.Minute,
	}

	// Update Description
	pb.updateDescription(progress)

	// Keine direkte Prüfung möglich, aber sollte nicht paniken
}

// TestProgressBar_ConcurrentAccess testet Thread-Safety
func TestProgressBar_ConcurrentAccess(t *testing.T) {
	var buf bytes.Buffer
	syncedOut := newSyncedWriter(&buf)
	config := ProgressBarConfig{
		Total:       100,
		UpdateRate:  10 * time.Millisecond,
		Output:      syncedOut,
		ShowSpinner: true,
	}

	pb := NewProgressBar(config)
	tracker := worker.NewProgressTracker(100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Progress Bar
	pbDone := make(chan struct{})
	go func() {
		pb.Start(ctx, tracker)
		close(pbDone)
	}()

	// Concurrent Updates
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			tracker.Update(1, 100)
			time.Sleep(5 * time.Millisecond)
		}
		close(done)
	}()

	// Concurrent Spinner Updates
	spinnerDone := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			pb.StartSpinner("file.jsonl")
			time.Sleep(20 * time.Millisecond)
			pb.StopSpinner()
		}
		close(spinnerDone)
	}()

	<-done
	<-spinnerDone
	cancel()
	<-pbDone // Wait for progress bar goroutine to finish

	// Sollte nicht paniken
}

// BenchmarkProgressBar_Update benchmarkt Progress Bar Updates
func BenchmarkProgressBar_Update(b *testing.B) {
	var buf bytes.Buffer
	config := ProgressBarConfig{
		Total:  b.N,
		Output: &buf,
	}

	pb := NewProgressBar(config)
	tracker := worker.NewProgressTracker(b.N)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pbDone := make(chan struct{})
	go func() {
		pb.Start(ctx, tracker)
		close(pbDone)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Update(1, 10)
	}

	cancel()
	<-pbDone // Wait for progress bar goroutine to finish
}

// BenchmarkSpinner_Render benchmarkt Spinner Rendering
func BenchmarkSpinner_Render(b *testing.B) {
	var buf bytes.Buffer
	spinner := NewSpinner(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spinner.render("test message")
		spinner.current = (spinner.current + 1) % len(spinner.frames)
	}
}
