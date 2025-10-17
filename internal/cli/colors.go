package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// ColorMode definiert den aktuellen Farbmodus
type ColorMode int

const (
	// ColorAuto aktiviert Farben automatisch, wenn ein TTY erkannt wird
	ColorAuto ColorMode = iota
	// ColorAlways erzwingt Farben (auch wenn kein TTY)
	ColorAlways
	// ColorNever deaktiviert Farben komplett
	ColorNever
)

var (
	// globalColorMode speichert den aktuellen Farbmodus
	globalColorMode = ColorAuto
	// colorOutput ist der Writer für farbige Ausgaben (unterstützt Windows)
	colorOutput io.Writer = colorable.NewColorableStdout()
	// colorError ist der Writer für farbige Fehlerausgaben (unterstützt Windows)
	colorError io.Writer = colorable.NewColorableStderr()
)

// ANSI Color Codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// SetColorMode setzt den globalen Farbmodus
func SetColorMode(mode ColorMode) {
	globalColorMode = mode
}

// colorsEnabled prüft, ob Farben aktiviert sind
func colorsEnabled() bool {
	switch globalColorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		// Automatische Erkennung: Farben nur wenn Stdout ein TTY ist
		return isatty.IsTerminal(os.Stdout.Fd())
	default:
		return false
	}
}

// colorize wendet einen Farbcode auf einen Text an (wenn Farben aktiviert)
func colorize(color, text string) string {
	if !colorsEnabled() {
		return text
	}
	return color + text + colorReset
}

// Success gibt eine grüne Success-Nachricht aus
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(colorOutput, colorize(colorGreen, "✅ "+msg))
}

// Error gibt eine rote Error-Nachricht aus
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(colorError, colorize(colorRed, "❌ "+msg))
}

// Warning gibt eine gelbe Warning-Nachricht aus
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(colorOutput, colorize(colorYellow, "⚠️  "+msg))
}

// ColoredText gibt eingefärbten Text zurück (ohne Ausgabe)
func ColoredText(color, text string) string {
	return colorize(color, text)
}

// GreenText gibt grünen Text zurück
func GreenText(text string) string {
	return colorize(colorGreen, text)
}

// RedText gibt roten Text zurück
func RedText(text string) string {
	return colorize(colorRed, text)
}

// YellowText gibt gelben Text zurück
func YellowText(text string) string {
	return colorize(colorYellow, text)
}
