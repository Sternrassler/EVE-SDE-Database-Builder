package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCmd_TextFormat(t *testing.T) {
	// Setup
	version = "1.0.0"
	commit = "abc123def456"
	buildTime = "2025-10-17T12:00:00Z"

	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--format", "text"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify output
	output := buf.String()
	if !strings.Contains(output, "Version:    1.0.0") {
		t.Errorf("expected output to contain version, got: %s", output)
	}
	if !strings.Contains(output, "Commit:     abc123def456") {
		t.Errorf("expected output to contain commit, got: %s", output)
	}
	if !strings.Contains(output, "Build Time: 2025-10-17T12:00:00Z") {
		t.Errorf("expected output to contain build time, got: %s", output)
	}
}

func TestVersionCmd_JSONFormat(t *testing.T) {
	// Setup
	version = "1.0.0"
	commit = "abc123def456"
	buildTime = "2025-10-17T12:00:00Z"

	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--format", "json"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify JSON output
	var info VersionInfo
	err = json.Unmarshal(buf.Bytes(), &info)
	if err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got: %s", info.Version)
	}
	if info.Commit != "abc123def456" {
		t.Errorf("expected commit 'abc123def456', got: %s", info.Commit)
	}
	if info.BuildTime != "2025-10-17T12:00:00Z" {
		t.Errorf("expected buildTime '2025-10-17T12:00:00Z', got: %s", info.BuildTime)
	}
}

func TestVersionCmd_InvalidFormat(t *testing.T) {
	// Setup
	version = "1.0.0"
	commit = "abc123"
	buildTime = "2025-10-17T12:00:00Z"

	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--format", "invalid"})

	// Execute
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}

	expectedErr := "unsupported format: invalid"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain '%s', got: %s", expectedErr, err.Error())
	}
}

func TestVersionCmd_DevVersion(t *testing.T) {
	// Setup - simulate development build
	version = "dev"
	commit = "unknown"
	buildTime = "unknown"

	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--format", "text"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify output contains dev values
	output := buf.String()
	if !strings.Contains(output, "Version:    dev") {
		t.Errorf("expected output to contain 'dev' version, got: %s", output)
	}
	if !strings.Contains(output, "Commit:     unknown") {
		t.Errorf("expected output to contain 'unknown' commit, got: %s", output)
	}
	if !strings.Contains(output, "Build Time: unknown") {
		t.Errorf("expected output to contain 'unknown' build time, got: %s", output)
	}
}

func TestVersionCmd_JSONStructure(t *testing.T) {
	// Setup
	version = "2.1.3"
	commit = "fedcba987654"
	buildTime = "2025-10-17T14:30:00Z"

	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--format", "json"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify JSON structure
	output := buf.String()
	if !strings.Contains(output, `"version"`) {
		t.Error("expected JSON output to contain 'version' field")
	}
	if !strings.Contains(output, `"commit"`) {
		t.Error("expected JSON output to contain 'commit' field")
	}
	if !strings.Contains(output, `"buildTime"`) {
		t.Error("expected JSON output to contain 'buildTime' field")
	}
}

func TestVersionCmd_HelpText(t *testing.T) {
	cmd := newVersionCmd()

	// Test Use field
	if cmd.Use != "version" {
		t.Errorf("expected Use to be 'version', got '%s'", cmd.Use)
	}

	// Test Short description
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Test Long description
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Test Example field
	if cmd.Example == "" {
		t.Error("expected Example field to be set")
	}

	// Verify Example contains multiple examples
	exampleLines := strings.Split(cmd.Example, "\n")
	exampleCount := 0
	for _, line := range exampleLines {
		if strings.Contains(line, "esdedb version") {
			exampleCount++
		}
	}
	if exampleCount < 2 {
		t.Errorf("expected at least 2 examples in Example field, found %d", exampleCount)
	}

	// Test Flags
	flags := cmd.Flags()
	formatFlag := flags.Lookup("format")
	if formatFlag == nil {
		t.Error("expected 'format' flag to be defined")
	} else if formatFlag.Usage == "" {
		t.Error("expected 'format' flag to have a usage description")
	}
}

func TestVersionCmd_FlagDefaults(t *testing.T) {
	cmd := newVersionCmd()

	// Check default value for format flag
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("expected 'format' flag to be defined")
	}

	if formatFlag.DefValue != "text" {
		t.Errorf("expected default format to be 'text', got: %s", formatFlag.DefValue)
	}
}

func TestVersionInfo_JSONMarshaling(t *testing.T) {
	info := VersionInfo{
		Version:   "1.2.3",
		Commit:    "abc123",
		BuildTime: "2025-10-17T12:00:00Z",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal VersionInfo: %v", err)
	}

	var decoded VersionInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal VersionInfo: %v", err)
	}

	if decoded.Version != info.Version {
		t.Errorf("expected version %s, got %s", info.Version, decoded.Version)
	}
	if decoded.Commit != info.Commit {
		t.Errorf("expected commit %s, got %s", info.Commit, decoded.Commit)
	}
	if decoded.BuildTime != info.BuildTime {
		t.Errorf("expected buildTime %s, got %s", info.BuildTime, decoded.BuildTime)
	}
}
