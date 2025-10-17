package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompletionCmd_BashGeneration(t *testing.T) {
	// Setup
	cmd := &cobra.Command{
		Use:   "esdedb",
		Short: "Test command",
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "bash"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify bash completion script contains expected content
	output := buf.String()
	if !strings.Contains(output, "# bash completion V2 for esdedb") {
		t.Error("expected bash completion header")
	}
	if !strings.Contains(output, "__esdedb_debug") {
		t.Error("expected bash completion functions")
	}
	if !strings.Contains(output, "complete") {
		t.Error("expected bash complete command")
	}
}

func TestCompletionCmd_ZshGeneration(t *testing.T) {
	// Setup
	cmd := &cobra.Command{
		Use:   "esdedb",
		Short: "Test command",
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "zsh"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify zsh completion script contains expected content
	output := buf.String()
	if !strings.Contains(output, "#compdef esdedb") {
		t.Error("expected zsh compdef directive")
	}
	if !strings.Contains(output, "_esdedb") {
		t.Error("expected zsh completion function")
	}
	if !strings.Contains(output, "# zsh completion for esdedb") {
		t.Error("expected zsh completion header")
	}
}

func TestCompletionCmd_FishGeneration(t *testing.T) {
	// Setup
	cmd := &cobra.Command{
		Use:   "esdedb",
		Short: "Test command",
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "fish"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify fish completion script contains expected content
	output := buf.String()
	if !strings.Contains(output, "# fish completion for esdedb") {
		t.Error("expected fish completion header")
	}
	if !strings.Contains(output, "function __esdedb") {
		t.Error("expected fish completion functions")
	}
}

func TestCompletionCmd_NoLoggingInOutput(t *testing.T) {
	// Test all three shell completions to ensure no logging output
	shells := []string{"bash", "zsh", "fish"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "esdedb",
				Short: "Test command",
			}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs([]string{"completion", shell})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("expected no error for %s, got: %v", shell, err)
			}

			output := buf.String()

			// Verify no JSON logging output
			if strings.Contains(output, `"level":"info"`) {
				t.Errorf("%s completion output contains logging JSON", shell)
			}
			if strings.Contains(output, `"message":"Application started"`) {
				t.Errorf("%s completion output contains startup logging", shell)
			}
			if strings.Contains(output, `"message":"Application shutting down"`) {
				t.Errorf("%s completion output contains shutdown logging", shell)
			}
		})
	}
}

func TestCompletionCmd_HelpText(t *testing.T) {
	// Setup
	cmd := &cobra.Command{
		Use:   "esdedb",
		Short: "Test command",
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "--help"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify help text
	output := buf.String()
	if !strings.Contains(output, "completion") {
		t.Error("expected help text to mention completion")
	}
	if !strings.Contains(output, "bash") {
		t.Error("expected help text to mention bash")
	}
	if !strings.Contains(output, "zsh") {
		t.Error("expected help text to mention zsh")
	}
	if !strings.Contains(output, "fish") {
		t.Error("expected help text to mention fish")
	}
}
