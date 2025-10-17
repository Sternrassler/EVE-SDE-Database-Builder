package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigInitCmd_DefaultOutput testet config init mit Standard-Output-Pfad
func TestConfigInitCmd_DefaultOutput(t *testing.T) {
	// Setup: Temporäres Verzeichnis
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "config.toml")

	// Execute
	err := runConfigInit(outputPath)
	if err != nil {
		t.Fatalf("runConfigInit failed: %v", err)
	}

	// Verify: Datei existiert
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("config.toml was not created at %s", outputPath)
	}

	// Verify: Content enthält erwartete Abschnitte
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}

	expectedSections := []string{
		"[database]",
		"[import]",
		"[logging]",
		"[update]",
		"version = \"1.0.0\"",
		"path = \"./eve_sde.db\"",
		"sde_path = \"./sde-JSONL\"",
		"language = \"en\"",
		"workers = 4",
		"level = \"info\"",
	}

	for _, section := range expectedSections {
		if !strings.Contains(string(content), section) {
			t.Errorf("config.toml missing expected section or value: %s", section)
		}
	}
}

// TestConfigInitCmd_CustomOutput testet config init mit benutzerdefiniertem Pfad
func TestConfigInitCmd_CustomOutput(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "configs", "production.toml")

	err := runConfigInit(customPath)
	if err != nil {
		t.Fatalf("runConfigInit with custom path failed: %v", err)
	}

	// Verify: Verzeichnis wurde erstellt
	if _, err := os.Stat(filepath.Dir(customPath)); os.IsNotExist(err) {
		t.Fatalf("directory for custom path was not created")
	}

	// Verify: Datei existiert
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Fatalf("config.toml was not created at custom path %s", customPath)
	}
}

// TestConfigInitCmd_Overwrite testet, dass bestehende Dateien überschrieben werden
func TestConfigInitCmd_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "config.toml")

	// Erstelle bestehende Datei
	oldContent := []byte("old content")
	if err := os.WriteFile(outputPath, oldContent, 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	// Execute: Überschreiben
	err := runConfigInit(outputPath)
	if err != nil {
		t.Fatalf("runConfigInit failed on overwrite: %v", err)
	}

	// Verify: Content wurde überschrieben
	newContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read overwritten config.toml: %v", err)
	}

	if bytes.Equal(oldContent, newContent) {
		t.Error("config.toml was not overwritten (content unchanged)")
	}

	if !strings.Contains(string(newContent), "[database]") {
		t.Error("overwritten config.toml has invalid content")
	}
}

// TestConfigConvertCmd_ValidXML testet XML → TOML Konvertierung
func TestConfigConvertCmd_ValidXML(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "ApplicationSettings.xml")
	outputPath := filepath.Join(tmpDir, "config.toml")

	// Setup: XML-Datei erstellen
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<Settings>
  <SelectedDB>SQLite</SelectedDB>
  <SelectedLanguage>English</SelectedLanguage>
  <SQLiteDBPath>C:\EVE\Database.db</SQLiteDBPath>
  <SDEPath>C:\EVE\sde</SDEPath>
  <ThreadCount>8</ThreadCount>
</Settings>`

	if err := os.WriteFile(inputPath, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("failed to create XML file: %v", err)
	}

	// Execute
	err := runConfigConvert(inputPath, outputPath)
	if err != nil {
		t.Fatalf("runConfigConvert failed: %v", err)
	}

	// Verify: TOML existiert
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("config.toml was not created")
	}

	// Verify: Content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}

	expectedValues := []string{
		`path = "C:\EVE\Database.db"`,
		`sde_path = "C:\EVE\sde"`,
		`language = "en"`,
		`workers = 8`,
	}

	for _, value := range expectedValues {
		if !strings.Contains(string(content), value) {
			t.Errorf("config.toml missing expected converted value: %s", value)
		}
	}

	// Verify: Comment about conversion
	if !strings.Contains(string(content), "Converted from VB.NET XML") {
		t.Error("config.toml missing conversion comment")
	}
}

// TestConfigConvertCmd_LanguageConversion testet Sprach-Konvertierung
func TestConfigConvertCmd_LanguageConversion(t *testing.T) {
	tests := []struct {
		vbLang   string
		expected string
	}{
		{"English", "en"},
		{"German", "de"},
		{"French", "fr"},
		{"Japanese", "ja"},
		{"Russian", "ru"},
		{"Chinese", "zh"},
		{"Spanish", "es"},
		{"Korean", "ko"},
	}

	for _, tt := range tests {
		t.Run(tt.vbLang, func(t *testing.T) {
			result := convertLanguage(tt.vbLang)
			if result != tt.expected {
				t.Errorf("convertLanguage(%q) = %q, want %q", tt.vbLang, result, tt.expected)
			}
		})
	}
}

// TestConfigConvertCmd_LanguageFallback testet Fallback für unbekannte Sprachen
func TestConfigConvertCmd_LanguageFallback(t *testing.T) {
	// Unbekannte Sprache sollte auf "en" fallen
	result := convertLanguage("UnknownLanguage")
	if result != "Un" && result != "en" {
		// Accept both: first 2 chars or default "en"
		t.Errorf("convertLanguage(UnknownLanguage) = %q, want 'Un' or 'en'", result)
	}

	// Leerer String sollte auf "en" fallen
	result = convertLanguage("")
	if result != "en" {
		t.Errorf("convertLanguage('') = %q, want 'en'", result)
	}
}

// TestConfigConvertCmd_MissingInputFile testet Fehlerbehandlung bei fehlendem Input
func TestConfigConvertCmd_MissingInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "nonexistent.xml")
	outputPath := filepath.Join(tmpDir, "config.toml")

	err := runConfigConvert(inputPath, outputPath)
	if err == nil {
		t.Fatal("runConfigConvert should fail with missing input file")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error message should mention missing file, got: %v", err)
	}
}

// TestConfigConvertCmd_InvalidXML testet Fehlerbehandlung bei ungültigem XML
func TestConfigConvertCmd_InvalidXML(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid.xml")
	outputPath := filepath.Join(tmpDir, "config.toml")

	// Ungültiges XML schreiben
	invalidXML := `<?xml version="1.0"?><Settings><Invalid</Settings>`
	if err := os.WriteFile(inputPath, []byte(invalidXML), 0644); err != nil {
		t.Fatalf("failed to create invalid XML: %v", err)
	}

	err := runConfigConvert(inputPath, outputPath)
	if err == nil {
		t.Fatal("runConfigConvert should fail with invalid XML")
	}

	if !strings.Contains(err.Error(), "parse XML") {
		t.Errorf("error message should mention XML parsing, got: %v", err)
	}
}

// TestConfigConvertCmd_DefaultValues testet Default-Werte bei fehlenden XML-Feldern
func TestConfigConvertCmd_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "minimal.xml")
	outputPath := filepath.Join(tmpDir, "config.toml")

	// Minimales XML (fehlende Felder)
	minimalXML := `<?xml version="1.0" encoding="utf-8"?>
<Settings>
  <SelectedDB>SQLite</SelectedDB>
  <SelectedLanguage>English</SelectedLanguage>
</Settings>`

	if err := os.WriteFile(inputPath, []byte(minimalXML), 0644); err != nil {
		t.Fatalf("failed to create minimal XML: %v", err)
	}

	err := runConfigConvert(inputPath, outputPath)
	if err != nil {
		t.Fatalf("runConfigConvert failed: %v", err)
	}

	// Verify: Default-Werte wurden gesetzt
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}

	expectedDefaults := []string{
		`path = "./eve_sde.db"`,    // Default DB-Pfad
		`sde_path = "./sde-JSONL"`, // Default SDE-Pfad
		`workers = 4`,              // Default Worker-Count
	}

	for _, value := range expectedDefaults {
		if !strings.Contains(string(content), value) {
			t.Errorf("config.toml missing expected default value: %s", value)
		}
	}
}

// TestConfigCmd_HelpText testet die Help-Texte
func TestConfigCmd_HelpText(t *testing.T) {
	// Test config command
	cmd := newConfigCmd()
	if cmd == nil {
		t.Fatal("newConfigCmd returned nil")
	}

	if !strings.Contains(cmd.Long, "Configuration management") &&
		!strings.Contains(cmd.Long, "Verwaltungsfunktionen") {
		t.Error("config command missing expected help text")
	}

	// Test init subcommand
	initCmd := newConfigInitCmd()
	if initCmd == nil {
		t.Fatal("newConfigInitCmd returned nil")
	}

	if !strings.Contains(initCmd.Short, "config.toml") {
		t.Error("config init command missing expected help text")
	}

	// Test convert subcommand
	convertCmd := newConfigConvertCmd()
	if convertCmd == nil {
		t.Fatal("newConfigConvertCmd returned nil")
	}

	if !strings.Contains(convertCmd.Short, "XML") {
		t.Error("config convert command missing expected help text")
	}
}

// TestConfigInitCmd_IntegrationWithCLI testet config init über CLI
func TestConfigInitCmd_IntegrationWithCLI(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-config.toml")

	cmd := newConfigInitCmd()
	cmd.SetArgs([]string{"--output", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config init command failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("config.toml was not created by CLI command")
	}

	// Verify file content (output goes to stdout, not captured in test)
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created config: %v", err)
	}

	if !strings.Contains(string(content), "[database]") {
		t.Error("created config.toml has invalid content")
	}
}

// TestConfigConvertCmd_IntegrationWithCLI testet config convert über CLI
func TestConfigConvertCmd_IntegrationWithCLI(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.xml")
	outputPath := filepath.Join(tmpDir, "output.toml")

	// Setup: Create XML file
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<Settings>
  <SelectedDB>SQLite</SelectedDB>
  <SelectedLanguage>German</SelectedLanguage>
  <SQLiteDBPath>./test.db</SQLiteDBPath>
  <SDEPath>./test-sde</SDEPath>
  <ThreadCount>2</ThreadCount>
</Settings>`

	if err := os.WriteFile(inputPath, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("failed to create test XML: %v", err)
	}

	cmd := newConfigConvertCmd()
	cmd.SetArgs([]string{"--input", inputPath, "--output", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config convert command failed: %v", err)
	}

	// Verify output file
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("config.toml was not created by convert command")
	}

	// Verify converted content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read converted config: %v", err)
	}

	if !strings.Contains(string(content), `language = "de"`) {
		t.Error("converted config missing German language setting")
	}

	if !strings.Contains(string(content), `workers = 2`) {
		t.Error("converted config missing ThreadCount conversion")
	}
}
