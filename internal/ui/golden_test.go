package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/testutil"
	"github.com/timoch/bd-view/internal/tree"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// --- View snapshot tests using golden files ---

func TestGolden_TreeWithChildren(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	beads := testutil.SampleBeads()
	tr := tree.BuildTree(beads, true) // expand all
	m.SetTree(tr)
	m.syncSelectedBead()

	output := m.View()
	testutil.GoldenFile(t, "tree_with_children", output)
}

func TestGolden_TreeCollapsed(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	beads := testutil.SampleBeads()
	tr := tree.BuildTree(beads, false) // collapsed
	m.SetTree(tr)
	m.syncSelectedBead()

	output := m.View()
	testutil.GoldenFile(t, "tree_collapsed", output)
}

func TestGolden_DetailPaneFull(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	beads := testutil.SampleBeads()
	tr := tree.BuildTree(beads, true)
	m.SetTree(tr)

	detail := testutil.SampleBeadDetail()
	m.SetSelectedBeadDetail(detail)

	output := m.View()
	testutil.GoldenFile(t, "detail_pane_full", output)
}

func TestGolden_EmptyState(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	output := m.View()
	testutil.GoldenFile(t, "empty_state", output)
}

func TestGolden_NarrowMode(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 90 // narrow: < 100
	m.height = 30
	m.ready = true

	beads := testutil.SampleBeads()
	tr := tree.BuildTree(beads, true)
	m.SetTree(tr)
	m.syncSelectedBead()

	output := m.View()
	testutil.GoldenFile(t, "narrow_mode", output)
}

func TestGolden_MissingOptionalFields(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	// Bead with minimal fields (no description, design, notes, etc.)
	m.SetSelectedBead(&testutil.SampleBeads()[4]) // proj-2: feature with no optional text fields

	output := m.View()
	testutil.GoldenFile(t, "missing_optional_fields", output)
}

func TestGolden_HelpOverlay(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true

	output := m.View()
	testutil.GoldenFile(t, "help_overlay", output)
}

func TestGolden_FilterOverlay(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.filtering = true
	m.filterTypes["task"] = true
	m.filterStats["open"] = true

	output := m.View()
	testutil.GoldenFile(t, "filter_overlay", output)
}
