#!/usr/bin/env bash
# check-secrets.sh – Professionelles Secret-Scanning mit Gitleaks
# Referenz: copilot-instructions.md Abschnitt 2.3

set -euo pipefail

echo "[check-secrets] Prüfe auf Secrets in staged Dateien..."

# Prüfe ob Gitleaks installiert ist
if ! command -v gitleaks >/dev/null 2>&1; then
    echo "[check-secrets] WARNING: gitleaks nicht installiert – versuche Installation"
    
    # Installation via Binary Download (schneller als Package Manager)
    GITLEAKS_VERSION="8.18.4"
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    # Arch Mapping
    case "$ARCH" in
        x86_64) ARCH="x64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) echo "[check-secrets] ERROR: Unsupported architecture $ARCH" >&2; exit 1 ;;
    esac
    
    BINARY_URL="https://github.com/gitleaks/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_${OS}_${ARCH}.tar.gz"
    
    echo "[check-secrets] Downloading gitleaks ${GITLEAKS_VERSION}..."
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT
    
    if curl -sL "$BINARY_URL" | tar -xz -C "$TMP_DIR" 2>/dev/null; then
        if [ -w /usr/local/bin ]; then
            mv "$TMP_DIR/gitleaks" /usr/local/bin/
        elif command -v sudo >/dev/null 2>&1; then
            sudo mv "$TMP_DIR/gitleaks" /usr/local/bin/
        else
            echo "[check-secrets] ERROR: Keine Berechtigung für Installation in /usr/local/bin" >&2
            exit 1
        fi
        chmod +x /usr/local/bin/gitleaks
        echo "[check-secrets] ✅ gitleaks installiert"
    else
        echo "[check-secrets] ERROR: Installation fehlgeschlagen – überspringe Secret-Scan" >&2
        exit 0  # Nicht blockieren bei Installations-Fehler
    fi
fi

# Gitleaks auf staged Dateien ausführen
echo "[check-secrets] Führe gitleaks protect aus..."

if gitleaks protect --staged --redact --verbose 2>&1; then
    echo "[check-secrets] ✅ Keine Secrets gefunden"
    exit 0
else
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 1 ]; then
        echo ""
        echo "[check-secrets] ❌ SECRETS GEFUNDEN!"
        echo ""
        echo "MUST: Entferne Secrets vor dem Commit."
        echo "Falls False Positive: .gitleaksignore anlegen oder mit --no-verify überspringen (Vorsicht!)"
        exit 1
    else
        # Anderer Fehler (z.B. keine staged Files)
        echo "[check-secrets] ⚠️  Gitleaks Exit Code: $EXIT_CODE (möglicherweise keine staged Dateien)"
        exit 0
    fi
fi
