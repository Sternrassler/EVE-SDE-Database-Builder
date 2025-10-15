#!/usr/bin/env bash
# create-github-labels.sh – Erstellt GitHub Labels für EVE SDE Database Builder
# Usage: ./scripts/github/create-github-labels.sh

set -euo pipefail

REPO="Sternrassler/EVE-SDE-Database-Builder"

echo "[create-labels] Erstelle GitHub Labels für $REPO..."

# Funktion zum Erstellen eines Labels
create_label() {
    local name=$1
    local color=$2
    local description=$3
    
    # Prüfe ob Label bereits existiert
    if gh label list --repo "$REPO" --json name | grep -q "\"$name\""; then
        echo "  ⚠️  Label '$name' existiert bereits (überspringen)"
    else
        gh label create "$name" --color "$color" --description "$description" --repo "$REPO"
        echo "  ✅ Label '$name' erstellt"
    fi
}

# Priorität
create_label "priority: critical" "d73a4a" "Kritisch, sofort adressieren"
create_label "priority: high" "e99695" "Hohe Priorität"
create_label "priority: medium" "fbca04" "Mittlere Priorität"
create_label "priority: low" "d4c5f9" "Niedrige Priorität"

# Typ
create_label "type: feature" "0075ca" "Neues Feature"
create_label "type: bug" "d73a4a" "Bug Fix"
create_label "type: refactor" "0e8a16" "Code Refactoring"
create_label "type: docs" "0075ca" "Dokumentation"
create_label "type: test" "1d76db" "Testing"
create_label "type: chore" "fef2c0" "Maintenance/Chores"

# Komponente
create_label "component: database" "006b75" "Database Layer (SQLite)"
create_label "component: parser" "1d76db" "JSONL Parser"
create_label "component: cli" "5319e7" "CLI Interface"
create_label "component: config" "fbca04" "Configuration"
create_label "component: logging" "bfdadc" "Logging/Error Handling"
create_label "component: worker" "0e8a16" "Concurrency/Worker Pool"
create_label "component: esi" "d4c5f9" "ESI Client"

# Bereich
create_label "area: go-migration" "0075ca" "Go Migration spezifisch"
create_label "area: architecture" "5319e7" "Architektur-Entscheidungen"
create_label "area: performance" "0e8a16" "Performance-Optimierung"
create_label "area: security" "d73a4a" "Security-relevante Issues"

# Status
create_label "status: blocked" "d73a4a" "Blockiert durch Dependencies"
create_label "status: in-progress" "fbca04" "In Bearbeitung"
create_label "status: needs-review" "0075ca" "Wartet auf Review"
create_label "status: ready" "0e8a16" "Bereit für Implementierung"

# Epic
create_label "epic: foundation" "5319e7" "Foundation Phase"
create_label "epic: database-layer" "006b75" "Database Layer Epic"
create_label "epic: parser-migration" "1d76db" "Parser Migration Epic"
create_label "epic: cli-interface" "0075ca" "CLI Interface Epic"
create_label "epic: testing" "1d76db" "Testing Strategy Epic"

# Sonstige
create_label "good-first-issue" "7057ff" "Gut für Einsteiger"
create_label "help-wanted" "008672" "Hilfe erwünscht"
create_label "wontfix" "ffffff" "Wird nicht gefixt"
create_label "duplicate" "cfd3d7" "Duplikat"
create_label "question" "d876e3" "Frage"

echo "[create-labels] ✅ Fertig! Labels erstellt."
