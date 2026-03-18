package commands_test

import (
	"bytes"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/commands"
)

func TestRootCmd_Version(t *testing.T) {
	cmd := commands.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("0.1.0")) {
		t.Errorf("version output = %q, want to contain '0.1.0'", buf.String())
	}
}

func TestInstallCmd_MissingFlags(t *testing.T) {
	cmd := commands.NewRootCmd()
	cmd.SetArgs([]string{"install"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing flags, got nil")
	}
}

func TestStatusCmd_Runs(t *testing.T) {
	cmd := commands.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"status", "--project-root", t.TempDir()})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("status should not error: %v", err)
	}
}

func TestResetCmd_RequiresYes(t *testing.T) {
	cmd := commands.NewRootCmd()
	cmd.SetArgs([]string{"reset", "--scope", "project", "--project-root", t.TempDir()})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error without --yes flag")
	}
}
