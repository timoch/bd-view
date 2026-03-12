package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/tree"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func sampleBeadsForState() []data.Bead {
	return []data.Bead{
		{ID: "epic-1", IssueType: "epic", Status: "open", Priority: 1},
		{ID: "task-1", IssueType: "task", Status: "open", Priority: 1, Parent: "epic-1"},
		{ID: "task-2", IssueType: "task", Status: "open", Priority: 2, Parent: "epic-1"},
		{ID: "epic-2", IssueType: "epic", Status: "open", Priority: 2},
		{ID: "task-3", IssueType: "task", Status: "open", Priority: 1, Parent: "epic-2"},
	}
}

func TestSaveAndLoadExpandState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	ids := []string{"epic-1", "epic-2"}
	SaveExpandState(path, ids)

	loaded := LoadExpandState(path)
	if loaded == nil {
		t.Fatal("expected non-nil loaded state")
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 expanded nodes, got %d", len(loaded))
	}
	if loaded[0] != "epic-1" || loaded[1] != "epic-2" {
		t.Errorf("expected [epic-1, epic-2], got %v", loaded)
	}
}

func TestLoadExpandState_MissingFile(t *testing.T) {
	loaded := LoadExpandState("/nonexistent/path/state.json")
	if loaded != nil {
		t.Error("expected nil for missing file")
	}
}

func TestLoadExpandState_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")
	os.WriteFile(path, []byte("not json"), 0644)

	loaded := LoadExpandState(path)
	if loaded != nil {
		t.Error("expected nil for corrupted file")
	}
}

func TestLoadExpandState_EmptyArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")
	raw, _ := json.Marshal(expandStateFile{ExpandedNodes: []string{}})
	os.WriteFile(path, raw, 0644)

	loaded := LoadExpandState(path)
	if loaded == nil {
		t.Fatal("expected non-nil for empty array")
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 expanded nodes, got %d", len(loaded))
	}
}

func TestSaveExpandState_SortsIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	SaveExpandState(path, []string{"z-node", "a-node", "m-node"})

	loaded := LoadExpandState(path)
	if len(loaded) != 3 {
		t.Fatalf("expected 3, got %d", len(loaded))
	}
	if loaded[0] != "a-node" || loaded[1] != "m-node" || loaded[2] != "z-node" {
		t.Errorf("expected sorted, got %v", loaded)
	}
}

func TestSaveExpandState_WriteError(t *testing.T) {
	// Writing to a non-existent directory should silently fail
	SaveExpandState("/nonexistent/dir/state.json", []string{"a"})
	// No panic = pass
}

func TestPersistExpandState_OnToggle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	beads := sampleBeadsForState()
	tr := tree.BuildTree(beads, false)

	m := New(Config{StatePath: path})
	m.SetTree(tr)
	m.width = 120
	m.height = 40
	m.ready = true
	m.beads = beads

	// Expand epic-1 via expandSelected
	m.selectedIdx = 0
	m.syncSelectedBead()
	m.expandSelected()

	// Wait briefly for async save
	time.Sleep(50 * time.Millisecond)

	loaded := LoadExpandState(path)
	if loaded == nil {
		t.Fatal("expected state file to be written")
	}
	found := false
	for _, id := range loaded {
		if id == "epic-1" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected epic-1 in expanded nodes, got %v", loaded)
	}
}

func TestPersistExpandState_OnExpandAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	beads := sampleBeadsForState()
	tr := tree.BuildTree(beads, false)

	m := New(Config{StatePath: path})
	m.SetTree(tr)
	m.width = 120
	m.height = 40
	m.ready = true
	m.beads = beads

	m.expandAllNodes()
	time.Sleep(50 * time.Millisecond)

	loaded := LoadExpandState(path)
	if loaded == nil {
		t.Fatal("expected state file after expand all")
	}
	// All parent nodes should be expanded
	if len(loaded) < 2 {
		t.Errorf("expected at least 2 expanded nodes, got %v", loaded)
	}
}

func TestPersistExpandState_OnCollapseAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	beads := sampleBeadsForState()
	tr := tree.BuildTree(beads, true) // start expanded

	m := New(Config{StatePath: path})
	m.SetTree(tr)
	m.width = 120
	m.height = 40
	m.ready = true
	m.beads = beads

	m.collapseAllNodes()
	time.Sleep(50 * time.Millisecond)

	loaded := LoadExpandState(path)
	if loaded == nil {
		t.Fatal("expected state file after collapse all")
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 expanded nodes after collapse all, got %v", loaded)
	}
}

func TestApplyRefresh_FirstLoad_RestoresPersistedState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	// Pre-save state: only epic-1 is expanded
	SaveExpandState(path, []string{"epic-1"})

	beads := sampleBeadsForState()
	m := New(Config{StatePath: path})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }

	m.applyRefresh(beads)

	// epic-1 should be expanded (from persisted state)
	if node, ok := m.tree.ByID["epic-1"]; !ok || !node.Expanded {
		t.Error("expected epic-1 to be expanded from persisted state")
	}
	// epic-2 should be collapsed (not in persisted state)
	if node, ok := m.tree.ByID["epic-2"]; !ok || node.Expanded {
		t.Error("expected epic-2 to be collapsed (not in persisted state)")
	}
}

func TestApplyRefresh_FirstLoad_ExpandAllExplicitOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	// Pre-save state: only epic-1 expanded
	SaveExpandState(path, []string{"epic-1"})

	beads := sampleBeadsForState()
	m := New(Config{StatePath: path, ExpandAll: true, ExpandAllExplicit: true})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }

	m.applyRefresh(beads)

	// Both epics should be expanded because --expand-all overrides persisted state
	if node, ok := m.tree.ByID["epic-1"]; !ok || !node.Expanded {
		t.Error("expected epic-1 expanded with --expand-all")
	}
	if node, ok := m.tree.ByID["epic-2"]; !ok || !node.Expanded {
		t.Error("expected epic-2 expanded with --expand-all")
	}
}

func TestApplyRefresh_FirstLoad_NoStateFile_FallsBackToExpandAll(t *testing.T) {
	beads := sampleBeadsForState()
	m := New(Config{StatePath: "/nonexistent/state.json", ExpandAll: true})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }

	m.applyRefresh(beads)

	// Should use ExpandAll=true since state file doesn't exist
	if node := m.tree.ByID["epic-1"]; !node.Expanded {
		t.Error("expected epic-1 expanded (fallback to --expand-all)")
	}
}

func TestApplyRefresh_FirstLoad_NoStateFile_DefaultCollapsed(t *testing.T) {
	beads := sampleBeadsForState()
	m := New(Config{StatePath: "/nonexistent/state.json"})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }

	m.applyRefresh(beads)

	// Should use default collapsed since state file doesn't exist
	if node := m.tree.ByID["epic-1"]; node.Expanded {
		t.Error("expected epic-1 collapsed by default")
	}
}

func TestApplyRefresh_SecondLoad_PreservesInMemoryState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	// Pre-save state: epic-1 expanded
	SaveExpandState(path, []string{"epic-1"})

	beads := sampleBeadsForState()
	m := New(Config{StatePath: path})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }

	// First load
	m.applyRefresh(beads)

	// Manually expand epic-2
	m.tree.ToggleExpand("epic-2")

	// Update the state file to have only epic-1 (simulating external change)
	SaveExpandState(path, []string{"epic-1"})

	// Second load (refresh) — should use in-memory state, not re-read file
	updated := time.Date(2026, 3, 12, 0, 1, 0, 0, time.UTC)
	modifiedBeads := make([]data.Bead, len(beads))
	copy(modifiedBeads, beads)
	modifiedBeads[0].UpdatedAt = &updated // trigger diff

	m.applyRefresh(modifiedBeads)

	// epic-2 should still be expanded (in-memory state preserved)
	if node := m.tree.ByID["epic-2"]; !node.Expanded {
		t.Error("expected epic-2 to remain expanded from in-memory state")
	}
}

func TestApplyRefresh_StaleIDsIgnored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bd-view-state.json")

	// Save state with an ID that doesn't exist
	SaveExpandState(path, []string{"epic-1", "nonexistent-id"})

	beads := sampleBeadsForState()
	m := New(Config{StatePath: path})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }

	m.applyRefresh(beads)

	// epic-1 should be expanded, nonexistent-id silently ignored
	if node := m.tree.ByID["epic-1"]; !node.Expanded {
		t.Error("expected epic-1 expanded")
	}
	if _, ok := m.tree.ByID["nonexistent-id"]; ok {
		t.Error("nonexistent-id should not be in tree")
	}
}

func TestCollectExpandedIDs(t *testing.T) {
	beads := sampleBeadsForState()
	tr := tree.BuildTree(beads, false)
	tr.ToggleExpand("epic-1")

	m := New(Config{})
	m.SetTree(tr)

	ids := m.collectExpandedIDs()
	if len(ids) != 1 {
		t.Fatalf("expected 1 expanded ID, got %d", len(ids))
	}
	if ids[0] != "epic-1" {
		t.Errorf("expected epic-1, got %s", ids[0])
	}
}

func TestPersistExpandState_NoStatePath(t *testing.T) {
	beads := sampleBeadsForState()
	tr := tree.BuildTree(beads, false)

	m := New(Config{}) // no StatePath
	m.SetTree(tr)

	// Should not panic or error
	m.persistExpandState()
}
