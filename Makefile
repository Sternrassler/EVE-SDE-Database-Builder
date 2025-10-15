# Makefile – Zentrale Orchestrierung für Projekt-Automationen
# Referenz: copilot-instructions.md Abschnitt 3.1
# Prinzip: Delegiere zu scripts/ statt eigenen Code zu duplizieren

.PHONY: help test lint scan adr-check adr-ref commit-lint normative-check release-check security-blockers check-hooks check-local check-pr check-ci release clean ensure-trivy

# Standardwerte
TRIVY_JSON_REPORT ?= tmp/trivy-fs-report.json
VERSION ?=

help: ## Zeigt verfügbare Targets
	@echo "Projekt Automations – Make Targets"
	@echo ""
	@echo "Basis-Checks:"
	@echo "  lint            - Statische Analysen / Format / Stil"
	@echo "  test            - Test-Suite ausführen"
	@echo "  scan            - Security & Dependency Checks (Trivy)"
	@echo ""
	@echo "Governance-Checks:"
	@echo "  normative-check - Prüft normative Schlüsselwörter (MUST/SHOULD)"
	@echo "  adr-check       - Prüft ADR Konsistenz"
	@echo "  adr-ref         - Prüft ADR-Referenzen in PR Bodies"
	@echo "  commit-lint     - Validiert Commit Messages (RANGE oder COMMIT_FILE nötig)"
	@echo "  release-check   - VERSION/CHANGELOG Synchronität prüfen"
	@echo "  security-blockers - Prüft kritische Security Findings"
	@echo ""
	@echo "Composite Targets:"
	@echo "  check-hooks   - Governance-Checks für Git Hooks (normative + adr)"
	@echo "  check-local   - lint + test (schnell, lokal)"
	@echo "  check-pr      - check-local + scan + security-blockers (vor PR)"
	@echo "  check-ci      - check-pr + governance (vollständig)"
	@echo ""
	@echo "Weitere:"
	@echo "  release       - Version bump + CHANGELOG (VERSION=X.Y.Z nötig)"
	@echo "  clean         - Temporäre Dateien entfernen"

# === Basis-Checks (atomar) ===

test: ## Führt die definierte Test-Suite aus
	@echo "[make test] Keine Tests konfiguriert – bitte projektspezifische Testbefehle ergänzen"

lint: ## Statische Analysen / Format / Stil
	@echo "[make lint] Kein Lint-Tool definiert – bitte projektspezifische Checks ergänzen"

scan: ensure-trivy ## Security & Dependency Checks (Trivy mit JSON Report)
	@echo "[make scan] Führe Security Scan aus..."
	@mkdir -p tmp
	@if command -v trivy >/dev/null 2>&1; then \
		trivy fs --ignore-unfixed --scanners vuln --format json -o $(TRIVY_JSON_REPORT) .; \
		echo "[make scan] Trivy JSON Report: $(TRIVY_JSON_REPORT)"; \
	else \
		echo "[make scan] trivy Installation fehlgeschlagen – überspringe Scan"; \
	fi

# === Governance-Checks (delegiert zu scripts/) ===

normative-check: ## Prüft normative Schlüsselwörter
	@bash scripts/common/check-normative.sh

adr-check: ## Prüft ADR Konsistenz
	@bash scripts/common/check-adr.sh

adr-ref: ## Prüft ADR-Referenzen in PR Bodies
	@bash scripts/common/check-adr-ref.sh

commit-lint: ## Validiert Commit Messages (RANGE oder COMMIT_FILE nötig)
	@if [ -n "$${RANGE:-}" ]; then \
		bash scripts/common/check-commit-msg.sh --range "$$RANGE"; \
	elif [ -n "$${COMMIT_FILE:-}" ]; then \
		bash scripts/common/check-commit-msg.sh --file "$$COMMIT_FILE"; \
	else \
		echo "[make commit-lint] ERROR: Bitte RANGE oder COMMIT_FILE angeben" >&2; \
		exit 1; \
	fi

release-check: ## Prüft VERSION/CHANGELOG Synchronität
	@bash scripts/common/check-version-changelog.sh

security-blockers: ## Prüft kritische Security Findings (benötigt scan vorher)
	@bash scripts/common/check-security-blockers.sh

# === Composite Targets ===

check-hooks: normative-check adr-check ## Governance-Checks für Git Hooks
	@echo "[make check-hooks] ✅ Git Hook Checks abgeschlossen"

check-local: lint test ## Schnelle lokale Checks (lint + test)
	@echo "[make check-local] ✅ Lokale Checks abgeschlossen"

check-pr: check-local scan security-blockers ## PR-Vorbereitung (check-local + scan + security)
	@echo "[make check-pr] ✅ PR Checks abgeschlossen"

check-ci: check-pr check-hooks adr-ref ## Vollständige CI-Simulation (check-pr + governance)
	@echo "[make check-ci] ✅ Alle CI Checks abgeschlossen"


# === Weitere Targets ===

release: ## Version bump + CHANGELOG Transform (Beispiel: make release VERSION=0.2.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "[make release] ERROR: VERSION Parameter fehlt (Beispiel: make release VERSION=0.2.0)" >&2; \
		exit 1; \
	fi
	@echo "[make release] Bump Version auf $(VERSION)..."
	@echo "$(VERSION)" > VERSION
	@sed -i "s/^## \[Unreleased\]/## [Unreleased]\n\n## [$(VERSION)] - $$(date +%Y-%m-%d)/" CHANGELOG.md
	@echo "[make release] VERSION und CHANGELOG aktualisiert – bitte commit + tag erstellen"

clean: ## Entfernt Build-Artefakte und temporäre Dateien
	@echo "[make clean] Räume temporäre Dateien auf..."
	@rm -rf tmp/*.md tmp/test-fixtures/ tmp/*.json
	@echo "[make clean] ✅ Clean abgeschlossen"

# === Interne Hilfs-Targets ===

ensure-trivy: ## Stellt sicher, dass Trivy verfügbar ist
	@if command -v trivy >/dev/null 2>&1; then \
		echo "[make ensure-trivy] trivy bereits verfügbar"; \
	else \
		echo "[make ensure-trivy] trivy nicht installiert – versuche Installation"; \
		if command -v apt-get >/dev/null 2>&1; then \
			if command -v sudo >/dev/null 2>&1; then \
				sudo apt-get update -y >/dev/null 2>&1 || true; \
				sudo apt-get install -y wget jq >/dev/null 2>&1 || true; \
			else \
				apt-get update -y >/dev/null 2>&1 || true; \
				apt-get install -y wget jq >/dev/null 2>&1 || true; \
			fi; \
		fi; \
		if command -v sudo >/dev/null 2>&1; then \
			curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sudo sh -s -- -b /usr/local/bin || true; \
		else \
			curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin || true; \
		fi; \
	fi


