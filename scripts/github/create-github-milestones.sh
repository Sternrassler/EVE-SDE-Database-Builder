#!/bin/bash
# GitHub Projects & Milestones Setup für Go Migration

set -euo pipefail

REPO="Sternrassler/EVE-SDE-Database-Builder"

echo "🎯 Erstelle GitHub Milestones..."

# Milestone 1: Foundation (v0.1)
gh api repos/${REPO}/milestones -X POST \
  -f title="v0.1 - Foundation" \
  -f description="Basis-Infrastruktur: Config, Logger, Error Handling, Retry Pattern" \
  -f due_on="2025-02-15T00:00:00Z" || echo "⚠️ Milestone v0.1 existiert bereits"

# Milestone 2: Database Layer (v0.2)
gh api repos/${REPO}/milestones -X POST \
  -f title="v0.2 - Database Layer" \
  -f description="SQLite Driver, Batch Insert, Migrations, Performance" \
  -f due_on="2025-02-28T00:00:00Z" || echo "⚠️ Milestone v0.2 existiert bereits"

# Milestone 3: Parser Core (v0.3)
gh api repos/${REPO}/milestones -X POST \
  -f title="v0.3 - Parser Core" \
  -f description="Code-Gen Tools, Generic JSONL Parser, Top-10 Parsers" \
  -f due_on="2025-03-15T00:00:00Z" || echo "⚠️ Milestone v0.3 existiert bereits"

# Milestone 4: Full Parser Migration (v0.4)
gh api repos/${REPO}/milestones -X POST \
  -f title="v0.4 - Full Parser Migration" \
  -f description="Alle 50+ SDE Tabellen generiert + getestet" \
  -f due_on="2025-03-31T00:00:00Z" || echo "⚠️ Milestone v0.4 existiert bereits"

# Milestone 5: CLI & Worker Pool (v0.5)
gh api repos/${REPO}/milestones -X POST \
  -f title="v0.5 - CLI & Worker Pool" \
  -f description="Vollständiges CLI, Worker Pool, 2-Phase Import" \
  -f due_on="2025-04-15T00:00:00Z" || echo "⚠️ Milestone v0.5 existiert bereits"

# Milestone 6: Testing & Performance (v0.6)
gh api repos/${REPO}/milestones -X POST \
  -f title="v0.6 - Testing & Performance" \
  -f description="Unit/Integration/E2E Tests, Benchmarks, Fuzz Tests" \
  -f due_on="2025-04-30T00:00:00Z" || echo "⚠️ Milestone v0.6 existiert bereits"

# Milestone 7: Release (v1.0)
gh api repos/${REPO}/milestones -X POST \
  -f title="v1.0 - Release" \
  -f description="Dokumentation, Binaries, Docker, Migration Guide" \
  -f due_on="2025-05-15T00:00:00Z" || echo "⚠️ Milestone v1.0 existiert bereits"

echo "✅ Milestones erstellt!"
echo ""
echo "📋 Verfügbare Milestones:"
gh api repos/${REPO}/milestones --jq '.[] | "  - \(.title) (Due: \(.due_on // "N/A"))"'
