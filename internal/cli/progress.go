package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
	"github.com/schollz/progressbar/v3"
)

// ProgressBar stellt eine erweiterte Progress Bar für Import-Operationen bereit.
//
// Features:
//   - Live-Update der Zeilen/Sekunde
//   - ETA-Anzeige (Estimated Time to Arrival)
//   - Spinner für einzelne Dateien
//   - Thread-Safe Updates über Channels
type ProgressBar struct {
	bar        *progressbar.ProgressBar
	tracker    *worker.ProgressTracker
	updateRate time.Duration
	spinner    *Spinner
	output     io.Writer
}

// ProgressBarConfig enthält Konfigurationsoptionen für die ProgressBar.
type ProgressBarConfig struct {
	// Total ist die Gesamtzahl der zu verarbeitenden Elemente
	Total int
	// Description ist die Beschreibung der Progress Bar
	Description string
	// Width ist die Breite der Progress Bar (default: 40)
	Width int
	// UpdateRate ist die Rate, mit der die Progress Bar aktualisiert wird (default: 100ms)
	UpdateRate time.Duration
	// ShowSpinner aktiviert den Spinner für einzelne Dateien
	ShowSpinner bool
	// Output definiert, wohin die Progress Bar geschrieben wird (default: os.Stdout)
	Output io.Writer
}

// NewProgressBar erstellt eine neue erweiterte ProgressBar.
//
// Parameter:
//   - config: Konfiguration für die ProgressBar
//
// Die ProgressBar unterstützt Live-Updates von Zeilen/Sekunde, ETA und
// optional einen Spinner für einzelne Dateien.
func NewProgressBar(config ProgressBarConfig) *ProgressBar {
	// Defaults setzen
	if config.Width == 0 {
		config.Width = 40
	}
	if config.UpdateRate == 0 {
		config.UpdateRate = 100 * time.Millisecond
	}
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.Description == "" {
		config.Description = "Processing"
	}

	// Progress Bar erstellen
	bar := progressbar.NewOptions(config.Total,
		progressbar.OptionSetWriter(config.Output),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(config.Width),
		progressbar.OptionSetDescription(fmt.Sprintf("[cyan]%s[reset]", config.Description)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("files"),
		progressbar.OptionClearOnFinish(),
	)

	pb := &ProgressBar{
		bar:        bar,
		updateRate: config.UpdateRate,
		output:     config.Output,
	}

	// Spinner erstellen, wenn aktiviert
	if config.ShowSpinner {
		pb.spinner = NewSpinner(config.Output)
	}

	return pb
}

// Start startet die automatische Aktualisierung der ProgressBar.
//
// Diese Methode überwacht einen ProgressTracker und aktualisiert die
// Anzeige automatisch mit Live-Metriken (Zeilen/Sekunde, ETA).
//
// Parameter:
//   - ctx: Context für Cancellation
//   - tracker: ProgressTracker, der überwacht werden soll
//
// Die Methode blockiert bis der Context abgebrochen wird oder der
// Import abgeschlossen ist.
func (pb *ProgressBar) Start(ctx context.Context, tracker *worker.ProgressTracker) {
	pb.tracker = tracker
	ticker := time.NewTicker(pb.updateRate)
	defer ticker.Stop()

	lastParsed := int64(0)

	for {
		select {
		case <-ctx.Done():
			// Context abgebrochen, finale Aktualisierung
			pb.updateToCompletion(lastParsed)
			return
		case <-ticker.C:
			// Periodische Aktualisierung
			if tracker != nil {
				progressInfo := tracker.GetProgressDetailed()
				diff := progressInfo.ParsedFiles - lastParsed
				if diff > 0 {
					// Progress Bar aktualisieren
					for i := int64(0); i < diff; i++ {
						_ = pb.bar.Add(1)
					}
					lastParsed = progressInfo.ParsedFiles

					// Custom Beschreibung mit Live-Metriken aktualisieren
					pb.updateDescription(progressInfo)
				}
			}
		}
	}
}

// updateDescription aktualisiert die Beschreibung mit Live-Metriken.
func (pb *ProgressBar) updateDescription(progress worker.Progress) {
	desc := fmt.Sprintf("[cyan]Importing[reset] [yellow]%.0f rows/s[reset]", progress.RowsPerSecond)

	// ETA hinzufügen, wenn verfügbar
	if progress.ETA > 0 {
		eta := formatDuration(progress.ETA)
		desc += fmt.Sprintf(" [blue]ETA: %s[reset]", eta)
	}

	pb.bar.Describe(desc)
}

// updateToCompletion aktualisiert die ProgressBar zur Vervollständigung.
func (pb *ProgressBar) updateToCompletion(lastParsed int64) {
	if pb.tracker != nil {
		progressInfo := pb.tracker.GetProgressDetailed()
		remaining := progressInfo.ParsedFiles - lastParsed
		for i := int64(0); i < remaining; i++ {
			_ = pb.bar.Add(1)
		}
	}
}

// Finish schließt die ProgressBar ab und gibt eine Zusammenfassung aus.
func (pb *ProgressBar) Finish() {
	_ = pb.bar.Finish()
	_, _ = fmt.Fprintln(pb.output) // Neue Zeile nach Progress Bar
}

// StartSpinner startet einen Spinner für eine einzelne Datei.
//
// Parameter:
//   - filename: Name der aktuell verarbeiteten Datei
//
// Der Spinner wird angezeigt, während eine Datei verarbeitet wird.
func (pb *ProgressBar) StartSpinner(filename string) {
	if pb.spinner != nil {
		pb.spinner.Start(filename)
	}
}

// StopSpinner stoppt den aktuellen Spinner.
func (pb *ProgressBar) StopSpinner() {
	if pb.spinner != nil {
		pb.spinner.Stop()
	}
}

// formatDuration formatiert eine Duration in ein lesbares Format.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// Spinner repräsentiert einen rotierenden Spinner für Datei-Verarbeitung.
type Spinner struct {
	frames  []string
	current int
	active  bool
	output  io.Writer
	ticker  *time.Ticker
	done    chan struct{}
}

// NewSpinner erstellt einen neuen Spinner.
func NewSpinner(output io.Writer) *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		output: output,
		done:   make(chan struct{}),
	}
}

// Start startet den Spinner mit einer Datei-Beschreibung.
func (s *Spinner) Start(message string) {
	if s.active {
		s.Stop()
	}

	s.active = true
	s.current = 0
	s.ticker = time.NewTicker(80 * time.Millisecond)

	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				s.render(message)
				s.current = (s.current + 1) % len(s.frames)
			}
		}
	}()
}

// Stop stoppt den Spinner und löscht die Zeile.
func (s *Spinner) Stop() {
	if !s.active {
		return
	}

	s.active = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.done)
	s.done = make(chan struct{}) // Für nächsten Start

	// Zeile löschen
	_, _ = fmt.Fprint(s.output, "\r"+strings.Repeat(" ", 80)+"\r")
}

// render rendert den aktuellen Spinner-Frame.
func (s *Spinner) render(message string) {
	frame := s.frames[s.current]
	_, _ = fmt.Fprintf(s.output, "\r%s %s", frame, message)
}
