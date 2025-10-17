// Package parser provides error recovery strategies for JSONL parsing.
package parser

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/rs/zerolog/log"
)

// ErrorMode defines the error handling strategy for parsing
type ErrorMode int

const (
	// ErrorModeSkip skips erroneous lines, logs them, and continues parsing
	ErrorModeSkip ErrorMode = iota
	// ErrorModeFailFast aborts parsing on the first error encountered
	ErrorModeFailFast
)

// String returns a string representation of the ErrorMode
func (m ErrorMode) String() string {
	switch m {
	case ErrorModeSkip:
		return "Skip"
	case ErrorModeFailFast:
		return "FailFast"
	default:
		return "Unknown"
	}
}

// ParseResult contains the results of parsing with error handling
type ParseResult[T any] struct {
	// Records contains successfully parsed records
	Records []T
	// Errors contains all errors encountered during parsing
	Errors []error
	// SkippedLines contains line numbers that were skipped
	SkippedLines []int
	// TotalLines is the total number of lines processed
	TotalLines int
}

// ParseWithErrorHandling parses a JSONL file with the specified error handling mode.
// It returns successfully parsed records and any errors encountered.
//
// Parameters:
//   - path: Path to the JSONL file
//   - mode: Error handling mode (Skip or FailFast)
//   - maxErrors: Maximum number of errors to tolerate (0 = unlimited for Skip mode)
//
// Returns:
//   - Slice of successfully parsed records
//   - Slice of errors encountered (empty if none)
func ParseWithErrorHandling[T any](path string, mode ErrorMode, maxErrors int) ([]T, []error) {
	ctx := context.Background()
	result := ParseWithErrorHandlingContext[T](ctx, path, mode, maxErrors)
	return result.Records, result.Errors
}

// ParseWithErrorHandlingContext parses a JSONL file with context support and error handling.
// This is the main implementation that provides detailed error recovery and reporting.
func ParseWithErrorHandlingContext[T any](ctx context.Context, path string, mode ErrorMode, maxErrors int) ParseResult[T] {
	// Use background context if none provided
	if ctx == nil {
		ctx = context.Background()
	}

	file, err := os.Open(path)
	if err != nil {
		return ParseResult[T]{
			Records:      nil,
			Errors:       []error{apperrors.NewFatal("failed to open file", err).WithContext("path", path)},
			SkippedLines: nil,
			TotalLines:   0,
		}
	}
	defer func() { _ = file.Close() }()

	return parseReaderWithErrorHandling[T](ctx, file, mode, maxErrors)
}

// parseReaderWithErrorHandling handles the actual line-by-line parsing with error recovery
func parseReaderWithErrorHandling[T any](ctx context.Context, r io.Reader, mode ErrorMode, maxErrors int) ParseResult[T] {
	scanner := bufio.NewScanner(r)
	// Set buffer size for potentially large JSON lines (10MB max line size)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	var results []T
	var errors []error
	var skippedLines []int
	lineNum := 0
	errorCount := 0

	for scanner.Scan() {
		lineNum++

		// Check for context cancellation
		select {
		case <-ctx.Done():
			errors = append(errors, apperrors.NewFatal("parsing cancelled", ctx.Err()).WithContext("line", lineNum))
			return ParseResult[T]{
				Records:      results,
				Errors:       errors,
				SkippedLines: skippedLines,
				TotalLines:   lineNum,
			}
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue // Skip empty lines
		}

		var item T
		if err := json.Unmarshal(line, &item); err != nil {
			errorCount++

			// Create skippable error with context
			parseErr := apperrors.NewSkippable("failed to parse JSON line", err).
				WithContext("line", lineNum).
				WithContext("content_preview", truncateString(string(line), 100))

			errors = append(errors, parseErr)
			skippedLines = append(skippedLines, lineNum)

			// Log the error
			log.Warn().
				Int("line", lineNum).
				Str("error", err.Error()).
				Str("mode", mode.String()).
				Msg("JSON parse error")

			// Handle based on error mode
			switch mode {
			case ErrorModeFailFast:
				// Fail immediately on first error
				log.Error().
					Int("line", lineNum).
					Msg("aborting parse due to error (FailFast mode)")
				return ParseResult[T]{
					Records:      results,
					Errors:       errors,
					SkippedLines: skippedLines,
					TotalLines:   lineNum,
				}

			case ErrorModeSkip:
				// Check if we've exceeded the error threshold
				if maxErrors > 0 && errorCount >= maxErrors {
					thresholdErr := apperrors.NewFatal("error threshold exceeded", nil).
						WithContext("max_errors", maxErrors).
						WithContext("error_count", errorCount).
						WithContext("line", lineNum)
					errors = append(errors, thresholdErr)

					log.Error().
						Int("error_count", errorCount).
						Int("max_errors", maxErrors).
						Int("line", lineNum).
						Msg("error threshold exceeded, aborting parse")

					return ParseResult[T]{
						Records:      results,
						Errors:       errors,
						SkippedLines: skippedLines,
						TotalLines:   lineNum,
					}
				}
				// Continue to next line
				continue
			}
		}

		results = append(results, item)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		scannerErr := apperrors.NewFatal("scanner error", err).
			WithContext("line", lineNum)
		errors = append(errors, scannerErr)

		log.Error().
			Int("line", lineNum).
			Err(err).
			Msg("scanner error occurred")
	}

	// Log summary if there were skipped lines
	if len(skippedLines) > 0 {
		log.Info().
			Int("total_lines", lineNum).
			Int("skipped_lines", len(skippedLines)).
			Int("successful_records", len(results)).
			Str("mode", mode.String()).
			Msg("parsing completed with errors")
	}

	return ParseResult[T]{
		Records:      results,
		Errors:       errors,
		SkippedLines: skippedLines,
		TotalLines:   lineNum,
	}
}

// truncateString truncates a string to the specified length and adds ellipsis if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ErrorSummary returns a formatted summary of parsing errors
func (r *ParseResult[T]) ErrorSummary() string {
	if len(r.Errors) == 0 {
		return "No errors"
	}

	return fmt.Sprintf("Encountered %d error(s) while parsing %d lines. Successfully parsed %d records. Skipped %d lines.",
		len(r.Errors), r.TotalLines, len(r.Records), len(r.SkippedLines))
}

// HasErrors returns true if any errors were encountered during parsing
func (r *ParseResult[T]) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasFatalErrors returns true if any fatal errors were encountered
func (r *ParseResult[T]) HasFatalErrors() bool {
	for _, err := range r.Errors {
		if apperrors.IsFatal(err) {
			return true
		}
	}
	return false
}
