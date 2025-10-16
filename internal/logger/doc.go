// Package logger provides structured logging with zerolog.
//
// Das logger-Package bietet eine einheitliche Logging-Schnittstelle für die EVE SDE Database Builder
// Anwendung. Es basiert auf zerolog und unterstützt verschiedene Log-Level und Ausgabeformate.
//
// # Grundlegende Verwendung
//
// Erstellen Sie einen neuen Logger mit dem gewünschten Log-Level und Format:
//
//	logger := logger.NewLogger("info", "json")
//	logger.Info("Application started", logger.Field{Key: "version", Value: "0.1.0"})
//
// # Ausgabeformate
//
// Das Package unterstützt zwei Ausgabeformate:
//   - "json": Strukturierte JSON-Ausgabe für Produktionsumgebungen
//   - "text": Menschenlesbare Konsolen-Ausgabe für Entwicklung
//
// # Log-Levels
//
// Verfügbare Log-Levels (von detailliert zu kritisch):
//   - "debug": Detaillierte Debug-Informationen
//   - "info": Allgemeine Informationsmeldungen
//   - "warn": Warnmeldungen
//   - "error": Fehlermeldungen
//   - "fatal": Kritische Fehler, die zum Programmabbruch führen
//
// # Strukturierte Felder
//
// Alle Logging-Methoden unterstützen optionale strukturierte Felder:
//
//	logger.Info("Database connected",
//	    logger.Field{Key: "host", Value: "localhost"},
//	    logger.Field{Key: "port", Value: 5432})
//
// # Context-basiertes Logging
//
// Logger können mit Context-Werten erweitert werden:
//
//	ctx := context.WithValue(context.Background(), logger.RequestIDKey, "req-123")
//	contextLogger := logger.WithContext(ctx)
//	contextLogger.Info("Processing request") // RequestID wird automatisch hinzugefügt
//
// # Globaler Logger
//
// Für einfache Anwendungsfälle kann ein globaler Logger verwendet werden:
//
//	logger.SetGlobalLogger(logger.NewLogger("info", "text"))
//	logger.LogAppStart("1.0.0", "abc123")
//
// # Helper-Funktionen
//
// Das Package bietet spezialisierte Helper-Funktionen für häufige Szenarien:
//   - LogHTTPRequest: HTTP-Request-Logging
//   - LogDBQuery: Datenbank-Query-Logging
//   - LogAppError: AppError-Logging mit automatischer Context-Extraktion
//   - LogAppStart/LogAppShutdown: Anwendungslebenszyklus-Logging
//
// Siehe auch: https://github.com/rs/zerolog für Details zur zugrunde liegenden Bibliothek
package logger
