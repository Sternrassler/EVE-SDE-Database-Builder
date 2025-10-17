package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestSetColorMode(t *testing.T) {
	tests := []struct {
		name string
		mode ColorMode
	}{
		{"Auto mode", ColorAuto},
		{"Always mode", ColorAlways},
		{"Never mode", ColorNever},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetColorMode(tt.mode)
			if globalColorMode != tt.mode {
				t.Errorf("SetColorMode(%v) failed, got %v", tt.mode, globalColorMode)
			}
		})
	}

	// Reset to default after test
	SetColorMode(ColorAuto)
}

func TestColorsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		mode     ColorMode
		expected bool
	}{
		{"Always mode should enable colors", ColorAlways, true},
		{"Never mode should disable colors", ColorNever, false},
		// ColorAuto behavior depends on TTY, nicht deterministisch testbar
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetColorMode(tt.mode)
			result := colorsEnabled()
			if result != tt.expected {
				t.Errorf("colorsEnabled() with mode %v = %v, want %v", tt.mode, result, tt.expected)
			}
		})
	}

	// Reset to default
	SetColorMode(ColorAuto)
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name            string
		mode            ColorMode
		color           string
		text            string
		wantContains    string
		wantNotContains string
	}{
		{
			name:         "Always mode adds color codes",
			mode:         ColorAlways,
			color:        colorRed,
			text:         "error",
			wantContains: "\033[31merror\033[0m",
		},
		{
			name:            "Never mode strips color codes",
			mode:            ColorNever,
			color:           colorGreen,
			text:            "success",
			wantContains:    "success",
			wantNotContains: "\033[",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetColorMode(tt.mode)
			result := colorize(tt.color, tt.text)

			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("colorize() = %q, want to contain %q", result, tt.wantContains)
			}

			if tt.wantNotContains != "" && strings.Contains(result, tt.wantNotContains) {
				t.Errorf("colorize() = %q, should not contain %q", result, tt.wantNotContains)
			}
		})
	}

	// Reset to default
	SetColorMode(ColorAuto)
}

func TestSuccess(t *testing.T) {
	// Backup original output
	originalOutput := colorOutput
	defer func() { colorOutput = originalOutput }()

	// Test mit ColorAlways
	SetColorMode(ColorAlways)
	buf := &bytes.Buffer{}
	colorOutput = buf

	Success("Operation completed")

	output := buf.String()
	if !strings.Contains(output, "✅") {
		t.Errorf("Success() output missing checkmark, got: %q", output)
	}
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Success() output missing message, got: %q", output)
	}
	if !strings.Contains(output, colorGreen) {
		t.Errorf("Success() output missing green color code, got: %q", output)
	}

	// Test mit ColorNever
	SetColorMode(ColorNever)
	buf.Reset()

	Success("Plain message")

	output = buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("Success() with ColorNever should not contain color codes, got: %q", output)
	}
	if !strings.Contains(output, "Plain message") {
		t.Errorf("Success() output missing message, got: %q", output)
	}

	// Reset to default
	SetColorMode(ColorAuto)
}

func TestError(t *testing.T) {
	// Backup original error output
	originalError := colorError
	defer func() { colorError = originalError }()

	// Test mit ColorAlways
	SetColorMode(ColorAlways)
	buf := &bytes.Buffer{}
	colorError = buf

	Error("Something went wrong")

	output := buf.String()
	if !strings.Contains(output, "❌") {
		t.Errorf("Error() output missing error symbol, got: %q", output)
	}
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("Error() output missing message, got: %q", output)
	}
	if !strings.Contains(output, colorRed) {
		t.Errorf("Error() output missing red color code, got: %q", output)
	}

	// Test mit ColorNever
	SetColorMode(ColorNever)
	buf.Reset()

	Error("Plain error")

	output = buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("Error() with ColorNever should not contain color codes, got: %q", output)
	}

	// Reset to default
	SetColorMode(ColorAuto)
}

func TestWarning(t *testing.T) {
	// Backup original output
	originalOutput := colorOutput
	defer func() { colorOutput = originalOutput }()

	// Test mit ColorAlways
	SetColorMode(ColorAlways)
	buf := &bytes.Buffer{}
	colorOutput = buf

	Warning("This is a warning")

	output := buf.String()
	if !strings.Contains(output, "⚠️") {
		t.Errorf("Warning() output missing warning symbol, got: %q", output)
	}
	if !strings.Contains(output, "This is a warning") {
		t.Errorf("Warning() output missing message, got: %q", output)
	}
	if !strings.Contains(output, colorYellow) {
		t.Errorf("Warning() output missing yellow color code, got: %q", output)
	}

	// Test mit ColorNever
	SetColorMode(ColorNever)
	buf.Reset()

	Warning("Plain warning")

	output = buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("Warning() with ColorNever should not contain color codes, got: %q", output)
	}

	// Reset to default
	SetColorMode(ColorAuto)
}

func TestColoredTextFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		text     string
		wantCode string
	}{
		{"GreenText", GreenText, "success", colorGreen},
		{"RedText", RedText, "error", colorRed},
		{"YellowText", YellowText, "warning", colorYellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test mit ColorAlways
			SetColorMode(ColorAlways)
			result := tt.fn(tt.text)
			if !strings.Contains(result, tt.wantCode) {
				t.Errorf("%s() = %q, want to contain color code %q", tt.name, result, tt.wantCode)
			}
			if !strings.Contains(result, tt.text) {
				t.Errorf("%s() = %q, want to contain text %q", tt.name, result, tt.text)
			}

			// Test mit ColorNever
			SetColorMode(ColorNever)
			result = tt.fn(tt.text)
			if strings.Contains(result, "\033[") {
				t.Errorf("%s() with ColorNever = %q, should not contain color codes", tt.name, result)
			}
			if result != tt.text {
				t.Errorf("%s() with ColorNever = %q, want %q", tt.name, result, tt.text)
			}
		})
	}

	// Reset to default
	SetColorMode(ColorAuto)
}

func TestSuccessWithFormatting(t *testing.T) {
	// Backup original output
	originalOutput := colorOutput
	defer func() { colorOutput = originalOutput }()

	SetColorMode(ColorAlways)
	buf := &bytes.Buffer{}
	colorOutput = buf

	Success("Processed %d files in %s", 42, "10s")

	output := buf.String()
	if !strings.Contains(output, "Processed 42 files in 10s") {
		t.Errorf("Success() formatting failed, got: %q", output)
	}

	// Reset to default
	SetColorMode(ColorAuto)
}

// Benchmark tests
func BenchmarkSuccess(b *testing.B) {
	originalOutput := colorOutput
	colorOutput = io.Discard
	defer func() { colorOutput = originalOutput }()

	SetColorMode(ColorAlways)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Success("Test message")
	}
}

func BenchmarkColorsDisabled(b *testing.B) {
	originalOutput := colorOutput
	colorOutput = io.Discard
	defer func() { colorOutput = originalOutput }()

	SetColorMode(ColorNever)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Success("Test message")
	}
}
