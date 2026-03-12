package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// shouldUpdate checks if the -update flag is set.
// We look up the flag by name to avoid conflicts with other packages
// (e.g., charmbracelet/x/exp/golden) that also define -update.
func shouldUpdate() bool {
	f := flag.Lookup("update")
	if f == nil {
		return false
	}
	return f.Value.String() == "true"
}

func init() {
	// Register -update only if not already defined by another package.
	if flag.Lookup("update") == nil {
		flag.Bool("update", false, "update golden files")
	}
}

// GoldenFile compares got against a golden file at testdata/<name>.golden.
// If -update is passed, it writes got to the golden file instead.
func GoldenFile(t *testing.T, name string, got string) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden")

	if shouldUpdate() {
		err := os.MkdirAll(filepath.Dir(goldenPath), 0o755)
		if err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		err = os.WriteFile(goldenPath, []byte(got), 0o644)
		if err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden file %s not found. Run with -update to create it: %v", goldenPath, err)
	}

	if got != string(expected) {
		t.Errorf("output does not match golden file %s.\n\nGot:\n%s\n\nWant:\n%s", goldenPath, got, string(expected))
	}
}
