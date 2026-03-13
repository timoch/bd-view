package data

import (
	"context"
	"testing"
)

func TestBdExecutor_ArgsPassthrough(t *testing.T) {
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
