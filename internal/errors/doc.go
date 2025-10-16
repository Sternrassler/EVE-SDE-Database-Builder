// Package errors provides custom error types with classification for precise error handling.
//
// Das errors-Package erweitert die Standard-Go-Error-Behandlung um typisierte Fehlerklassifizierung,
// die eine präzise Fehlerbehandlung in der EVE SDE Database Builder Anwendung ermöglicht.
//
// # Fehlertypen
//
// Das Package definiert vier primäre Fehlerklassifizierungen:
//
//   - ErrorTypeFatal: Kritische Fehler, die nicht behoben werden können und zum Programmabbruch führen
//   - ErrorTypeRetryable: Transiente Fehler, die bei erneutem Versuch möglicherweise erfolgreich sind
//   - ErrorTypeValidation: Eingabevalidierungsfehler, die auf ungültige Daten hinweisen
//   - ErrorTypeSkippable: Fehler, die sicher übersprungen werden können (z.B. optionale Felder)
//
// # Grundlegende Verwendung
//
// Erstellen Sie typisierte Fehler mit den entsprechenden Konstruktoren:
//
//	err := errors.NewRetryable("API request failed", originalErr)
//	err = err.WithContext("endpoint", "/api/v1/data")
//
// # Fehlerprüfung
//
// Verwenden Sie die Type-Checker-Funktionen für Fehlerbehandlung:
//
//	if errors.IsRetryable(err) {
//	    // Retry-Logik
//	} else if errors.IsFatal(err) {
//	    // Abbruch
//	}
//
// # Context-Informationen
//
// AppError unterstützt das Anhängen von strukturierten Context-Informationen:
//
//	err := errors.NewValidation("invalid email format", nil)
//	err = err.WithContext("field", "email").WithContext("value", "invalid@")
//
// # Error Wrapping
//
// AppError unterstützt Go 1.13+ Error Wrapping mit errors.Is und errors.As:
//
//	var appErr *errors.AppError
//	if errors.As(err, &appErr) {
//	    fmt.Printf("Error Type: %s\n", appErr.Type)
//	}
//
// # Integration mit Retry-Package
//
// AppError arbeitet nahtlos mit dem retry-Package zusammen:
//
//	policy := retry.DefaultPolicy()
//	err := policy.Do(ctx, func() error {
//	    if someError {
//	        return errors.NewRetryable("temporary failure", someError)
//	    }
//	    return nil
//	})
//
// Nur Fehler vom Typ ErrorTypeRetryable werden automatisch wiederholt.
//
// # Best Practices
//
//   - Verwenden Sie immer die entsprechende Fehlerklassifizierung
//   - Fügen Sie relevante Context-Informationen für Debugging hinzu
//   - Wrappen Sie ursprüngliche Fehler, um die Fehlerkette zu erhalten
//   - Nutzen Sie die Type-Checker anstelle von String-Vergleichen
//
// Siehe auch: internal/retry für Retry-Mechanismen mit ErrorTypeRetryable
package errors
