package data

import (
	"context"
	"testing"
)

func TestBdExecutor_DBFlagPassthrough(t *testing.T) {
	// Use a fake bd command (echo) to verify args are passed correctly.
	exec := &BdExecutor{
		BdPath: "echo",
		DBPath: "/path/to/beads.db",
	}

	out, err := exec.Execute(context.Background(), "list", "--all", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "--db /path/to/beads.db list --all --json\n"
	if string(out) != expected {
		t.Errorf("expected args %q, got %q", expected, string(out))
	}
}

func TestBdExecutor_NoDBFlag(t *testing.T) {
	exec := &BdExecutor{
		BdPath: "echo",
	}

	out, err := exec.Execute(context.Background(), "list", "--all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "list --all\n"
	if string(out) != expected {
		t.Errorf("expected args %q, got %q", expected, string(out))
	}
}

func TestBdExecutor_DefaultBdPath(t *testing.T) {
	exec := &BdExecutor{}

	// Calling with a nonexistent command should fail, confirming it defaults to "bd"
	_, err := exec.Execute(context.Background(), "version")
	// This may succeed or fail depending on environment, but we just verify no panic
	_ = err
}
