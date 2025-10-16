// Package parser provides JSONL parsing interfaces and implementations for EVE SDE data files.
//
// Das parser-Package definiert ein einheitliches Interface für das Parsen von JSONL-Dateien
// (JSON Lines Format) der EVE Online Static Data Export (SDE).
//
// # Grundlegende Verwendung
//
// Erstellen Sie einen typisierten Parser für eine spezifische Datenstruktur:
//
//	type TypeRow struct {
//	    TypeID   int    `json:"typeID"`
//	    TypeName string `json:"typeName"`
//	}
//
//	parser := parser.NewJSONLParser[TypeRow](
//	    "invTypes",
//	    []string{"typeID", "typeName"},
//	)
//
//	ctx := context.Background()
//	records, err := parser.ParseFile(ctx, "types.jsonl")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Parser Interface
//
// Das Parser-Interface definiert drei Methoden:
//   - ParseFile(ctx, path): Liest und parsed eine JSONL-Datei
//   - TableName(): Gibt den Datenbanktabellennamen zurück
//   - Columns(): Gibt die Spaltennamen für den Datenbankimport zurück
//
// # Generischer JSONLParser
//
// JSONLParser[T] ist eine generische Implementierung, die für beliebige Typen verwendet
// werden kann. Der Parser liest die Datei Zeile für Zeile und unmarshalt jede Zeile
// als JSON-Objekt des Typs T.
//
// # Fehlerbehandlung
//
// Fehler werden mit Zeilennummern angereichert, um die Fehlersuche zu erleichtern:
//
//	// Fehlerausgabe bei ungültigem JSON in Zeile 42:
//	// "line 42: failed to parse JSON: unexpected end of JSON input"
//
// # Context-Unterstützung
//
// ParseFile unterstützt Context für Timeout und Cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	records, err := parser.ParseFile(ctx, "large-file.jsonl")
//	if errors.Is(err, context.DeadlineExceeded) {
//	    log.Println("Parsing timeout")
//	}
//
// # Referenzen
//
// Siehe ADR-003 für die Architektur-Entscheidung zur JSONL-Parser-Implementierung:
// docs/adr/ADR-003-jsonl-parser-architecture.md
package parser
