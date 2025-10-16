// Package retry provides exponential backoff retry logic for transient errors.
//
// Das retry-Package implementiert robuste Retry-Mechanismen mit exponentiellem Backoff
// für die Behandlung transienter Fehler in der EVE SDE Database Builder Anwendung.
//
// # Grundkonzept
//
// Das Package arbeitet mit dem errors-Package zusammen und wiederholt nur Operationen,
// die einen ErrorTypeRetryable-Fehler zurückgeben. Andere Fehlertypen führen zu einem
// sofortigen Abbruch ohne Retry.
//
// # Retry-Policies
//
// Eine Policy definiert das Retry-Verhalten:
//
//	policy := retry.NewPolicy(
//	    3,                    // maxRetries
//	    100*time.Millisecond, // initialDelay
//	    5*time.Second,        // maxDelay
//	)
//
// # Vordefinierte Policies
//
// Das Package bietet spezialisierte Policies für häufige Szenarien:
//
//   - DefaultPolicy(): Allgemeine Anwendungen (3 retries, 100ms-5s)
//   - DatabasePolicy(): Datenbank-Operationen (5 retries, 50ms-2s)
//   - HTTPPolicy(): HTTP-Requests (3 retries, 100ms-5s)
//   - FileIOPolicy(): Datei-I/O (2 retries, 10ms-500ms)
//
// # Einfache Verwendung
//
// Verwenden Sie Do() für Funktionen ohne Rückgabewert:
//
//	policy := retry.DefaultPolicy()
//	err := policy.Do(ctx, func() error {
//	    return someOperation()
//	})
//
// # Generische Retry mit Rückgabewert
//
// Verwenden Sie DoWithResult() für Funktionen mit Rückgabewert:
//
//	result, err := retry.DoWithResult(ctx, policy, func() (Data, error) {
//	    return fetchData()
//	})
//
// # Policy Builder
//
// Für erweiterte Konfiguration verwenden Sie den PolicyBuilder:
//
//	policy := retry.NewPolicyBuilder().
//	    WithMaxRetries(5).
//	    WithInitialDelay(200*time.Millisecond).
//	    WithMaxDelay(10*time.Second).
//	    WithJitter(true).
//	    Build()
//
// # Exponentieller Backoff
//
// Die Verzögerung zwischen Versuchen wird exponentiell erhöht:
//   - Versuch 1: initialDelay
//   - Versuch 2: initialDelay * multiplier
//   - Versuch 3: initialDelay * multiplier^2
//   - usw., begrenzt durch maxDelay
//
// # Jitter
//
// Jitter (standardmäßig aktiviert) fügt zufällige Variation (±10%) zur Verzögerung hinzu,
// um "Thundering Herd"-Probleme bei parallelen Requests zu vermeiden.
//
// # Context-Unterstützung
//
// Alle Retry-Funktionen respektieren Context-Cancellation und Timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	err := policy.Do(ctx, func() error {
//	    // Diese Operation wird bei Context-Cancellation abgebrochen
//	    return someOperation()
//	})
//
// # Integration mit errors-Package
//
// Nur Fehler vom Typ errors.ErrorTypeRetryable werden wiederholt:
//
//	policy.Do(ctx, func() error {
//	    if transientError {
//	        return errors.NewRetryable("temporary failure", err)
//	    }
//	    if permanentError {
//	        return errors.NewFatal("permanent failure", err) // Kein Retry
//	    }
//	    return nil
//	})
//
// # TOML-Konfiguration
//
// Policies können aus TOML-Konfiguration geladen werden:
//
//	cfg := retry.PolicyConfig{
//	    MaxRetries:     5,
//	    InitialDelayMs: 100,
//	    MaxDelayMs:     5000,
//	    Multiplier:     2.0,
//	    Jitter:         true,
//	}
//	policy := retry.FromConfig(cfg)
//
// # Best Practices
//
//   - Verwenden Sie passende vordefinierte Policies für Ihren Use Case
//   - Aktivieren Sie Jitter für parallele Retry-Szenarien
//   - Setzen Sie angemessene Context-Timeouts
//   - Klassifizieren Sie Fehler korrekt (retryable vs. fatal)
//   - Beachten Sie externe Rate Limits bei der Policy-Konfiguration
//
// Siehe auch: internal/errors für Fehlerklassifizierung
package retry
