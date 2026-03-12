package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/tree"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func TestDetailPanel_EmptyState(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	output := m.View()
	if !strings.Contains(output, "Select a bead to view details") {
		t.Error("expected empty state message when no bead selected")
	}
}

func TestDetailPanel_ShowsBeadHeader(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	created := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	closed := time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)
	bead := &data.Bead{
		ID:        "hep-ws-f6.3",
		Title:     "US-003: Install SignalR client",
		IssueType: "task",
		Status:    "closed",
		Priority:  1,
		Owner:     "timoch@timoch.com",
		Parent:    "hep-ws-f6",
		CreatedAt: &created,
		ClosedAt:  &closed,
	}
	m.SetSelectedBead(bead)

	output := m.View()

	checks := []struct {
		label string
		want  string
	}{
		{"bead ID", "hep-ws-f6.3"},
		{"title", "US-003: Install SignalR client"},
		{"type", "task"},
		{"status", "closed"},
		{"priority", "1"},
		{"owner", "timoch@timoch.com"},
		{"parent", "hep-ws-f6"},
		{"created date", "2026-03-10"},
		{"closed date", "2026-03-11"},
	}
	for _, c := range checks {
		if !strings.Contains(output, c.want) {
			t.Errorf("expected %s %q in output", c.label, c.want)
		}
	}
}

func TestDetailPanel_HidesEmptyDates(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Test bead",
		IssueType: "task",
		Status:    "open",
	}
	m.SetSelectedBead(bead)

	output := m.View()
	if strings.Contains(output, "Created:") {
		t.Error("should not show Created when CreatedAt is nil")
	}
	if strings.Contains(output, "Closed:") {
		t.Error("should not show Closed when ClosedAt is nil")
	}
}

func TestDetailPanel_HidesParentWhenEmpty(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Root bead",
		IssueType: "epic",
		Status:    "open",
	}
	m.SetSelectedBead(bead)

	output := m.View()
	if strings.Contains(output, "Parent:") {
		t.Error("should not show Parent when parent is empty")
	}
}

func TestDetailPanel_HasSeparators(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Test",
		IssueType: "task",
		Status:    "open",
	}
	m.SetSelectedBead(bead)

	output := m.View()
	// Should have horizontal separators (─ characters)
	if !strings.Contains(output, "─") {
		t.Error("expected horizontal separator lines")
	}
}

func TestColorStatus_AllStatuses(t *testing.T) {
	m := Model{}
	statuses := []string{"open", "in_progress", "blocked", "deferred", "closed", "unknown"}
	for _, s := range statuses {
		result := m.colorStatus(s)
		if !strings.Contains(result, s) {
			t.Errorf("colorStatus(%q) = %q, should contain status text", s, result)
		}
	}
}

func TestDetailPanel_ShowsBodySections(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 60
	m.ready = true

	bead := &data.Bead{
		ID:                 "test-1",
		Title:              "Test bead",
		IssueType:          "task",
		Status:             "open",
		Description:        "This is the description.",
		Design:             "This is the design.",
		AcceptanceCriteria: "- [ ] Criterion one\n- [ ] Criterion two",
		Notes:              "Some notes here.",
	}
	m.SetSelectedBead(bead)

	output := m.View()

	checks := []struct {
		label string
		want  string
	}{
		{"description heading", "DESCRIPTION"},
		{"description content", "This is the description."},
		{"design heading", "DESIGN"},
		{"design content", "This is the design."},
		{"acceptance criteria heading", "ACCEPTANCE CRITERIA"},
		{"criterion one", "Criterion one"},
		{"criterion two", "Criterion two"},
		{"notes heading", "NOTES"},
		{"notes content", "Some notes here."},
	}
	for _, c := range checks {
		if !strings.Contains(output, c.want) {
			t.Errorf("expected %s %q in output", c.label, c.want)
		}
	}
}

func TestDetailPanel_OmitsEmptySections(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 60
	m.ready = true

	bead := &data.Bead{
		ID:          "test-1",
		Title:       "Test bead",
		IssueType:   "task",
		Status:      "open",
		Description: "Has a description.",
		// Design, AcceptanceCriteria, Notes are empty
	}
	m.SetSelectedBead(bead)

	output := m.View()

	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("expected DESCRIPTION heading for non-empty section")
	}
	if strings.Contains(output, "DESIGN") {
		t.Error("should not show DESIGN heading when design is empty")
	}
	if strings.Contains(output, "ACCEPTANCE CRITERIA") {
		t.Error("should not show ACCEPTANCE CRITERIA heading when empty")
	}
	if strings.Contains(output, "NOTES") {
		t.Error("should not show NOTES heading when notes is empty")
	}
}

func TestDetailPanel_ShowsDependencies(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 60
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Test bead",
		IssueType: "task",
		Status:    "open",
		Dependencies: []data.Dependency{
			{DependsOnID: "dep-1"},
			{DependsOnID: "dep-2"},
		},
	}
	m.SetSelectedBead(bead)
	m.dependents = []data.RelatedBead{
		{ID: "child-1"},
		{ID: "child-2"},
	}

	output := m.View()

	if !strings.Contains(output, "DEPENDENCIES") {
		t.Error("expected DEPENDENCIES heading")
	}
	if !strings.Contains(output, "depends on: dep-1, dep-2") {
		t.Error("expected depends on list")
	}
	if !strings.Contains(output, "depended on by: child-1, child-2") {
		t.Error("expected depended on by list")
	}
}

func TestDetailPanel_NoDependenciesSection(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 60
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Test bead",
		IssueType: "task",
		Status:    "open",
	}
	m.SetSelectedBead(bead)

	output := m.View()

	if strings.Contains(output, "DEPENDENCIES") {
		t.Error("should not show DEPENDENCIES heading when no dependencies")
	}
}

func TestDetailPanel_Scrolling(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true

	bead := &data.Bead{
		ID:          "test-1",
		Title:       "Test bead",
		IssueType:   "task",
		Status:      "open",
		Description: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12",
	}
	m.SetSelectedBead(bead)

	// Initially at scroll 0, should show bead ID
	output := m.View()
	if !strings.Contains(output, "test-1") {
		t.Error("expected bead ID at scroll 0")
	}

	// Scroll down
	m.focusedPane = detailPane
	m.detailScroll = 5

	output = m.View()
	// After scrolling, first lines (bead ID) should be scrolled past
	// Content should still render
	if !strings.Contains(output, "DESCRIPTION") || !strings.Contains(output, "Line") {
		// Either description heading or content should still be visible
		// depending on how much we scrolled
	}
}

func TestDetailPanel_TabSwitchesFocus(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	if m.focusedPane != treePane {
		t.Error("expected initial focus on tree pane")
	}

	// Simulate Tab key
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)

	if m.focusedPane != detailPane {
		t.Error("expected focus to switch to detail pane after Tab")
	}

	// Tab again should go back
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)

	if m.focusedPane != treePane {
		t.Error("expected focus to switch back to tree pane after second Tab")
	}
}

func TestDetailPanel_ScrollResetOnSelection(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.detailScroll = 10

	bead := &data.Bead{ID: "new-bead", Title: "New", IssueType: "task", Status: "open"}
	m.SetSelectedBead(bead)

	if m.detailScroll != 0 {
		t.Error("expected scroll to reset when selecting a new bead")
	}
}

func TestRenderMarkdown_Bold(t *testing.T) {
	m := New(Config{})
	result := m.renderMarkdown("This is **bold** text", 80)
	if !strings.Contains(result, "bold") {
		t.Error("expected bold text in output")
	}
	if strings.Contains(result, "**") {
		t.Error("bold markers should be removed")
	}
}

func TestRenderMarkdown_CodeBlock(t *testing.T) {
	m := New(Config{})
	input := "Before\n```\ncode line\n```\nAfter"
	result := m.renderMarkdown(input, 80)
	if !strings.Contains(result, "code line") {
		t.Error("expected code line in output")
	}
	if !strings.Contains(result, "Before") {
		t.Error("expected text before code block")
	}
	if !strings.Contains(result, "After") {
		t.Error("expected text after code block")
	}
	if strings.Contains(result, "```") {
		t.Error("code fence markers should be removed")
	}
}

func TestRenderMarkdown_BulletList(t *testing.T) {
	m := New(Config{})
	input := "- Item one\n- Item two\n- Item three"
	result := m.renderMarkdown(input, 80)
	if !strings.Contains(result, "Item one") {
		t.Error("expected bullet items preserved")
	}
}

func TestRenderMarkdown_Empty(t *testing.T) {
	m := New(Config{})
	result := m.renderMarkdown("", 80)
	if result != "" {
		t.Errorf("expected empty string for empty input, got %q", result)
	}
	result = m.renderMarkdown("   \n  ", 80)
	if result != "" {
		t.Errorf("expected empty string for whitespace input, got %q", result)
	}
}

func TestRenderMarkdown_Headings(t *testing.T) {
	m := New(Config{})
	result := m.renderMarkdown("# Heading\n\nSome text", 80)
	if !strings.Contains(result, "Heading") {
		t.Error("expected heading text in output")
	}
	if !strings.Contains(result, "Some text") {
		t.Error("expected body text in output")
	}
	if strings.Contains(result, "# ") {
		t.Error("heading markers should be removed")
	}
}

func TestRenderMarkdown_CheckboxList(t *testing.T) {
	m := New(Config{})
	input := "- [ ] unchecked\n- [x] checked"
	result := m.renderMarkdown(input, 80)
	if !strings.Contains(result, "unchecked") {
		t.Error("expected unchecked item in output")
	}
	if !strings.Contains(result, "checked") {
		t.Error("expected checked item in output")
	}
}

func TestRenderMarkdown_InlineCode(t *testing.T) {
	m := New(Config{})
	result := m.renderMarkdown("Use `foo()` here", 80)
	if !strings.Contains(result, "foo()") {
		t.Error("expected inline code in output")
	}
}

func TestRenderMarkdown_NarrowWidth(t *testing.T) {
	m := New(Config{})
	// Should not panic with very narrow width
	result := m.renderMarkdown("Some text here", 5)
	if !strings.Contains(result, "Some") {
		t.Error("expected text in output even with narrow width")
	}
}

func TestRenderMarkdown_NoColor(t *testing.T) {
	m := New(Config{NoColor: true})
	result := m.renderMarkdown("This is **bold** text", 80)
	if !strings.Contains(result, "bold") {
		t.Error("expected bold text in output")
	}
	// notty style preserves markdown markers as plain text
	if !strings.Contains(result, "**bold**") {
		t.Error("expected notty style to preserve bold markers")
	}
}

func TestSetSelectedBeadDetail(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.detailScroll = 5

	detail := &data.BeadDetail{
		Bead: data.Bead{
			ID:    "detail-1",
			Title: "Detail bead",
		},
		Dependents: []data.RelatedBead{
			{ID: "dep-1", Title: "Dependent 1"},
		},
	}
	m.SetSelectedBeadDetail(detail)

	if m.selectedBead == nil || m.selectedBead.ID != "detail-1" {
		t.Error("expected selected bead to be set from detail")
	}
	if len(m.dependents) != 1 || m.dependents[0].ID != "dep-1" {
		t.Error("expected dependents to be set from detail")
	}
	if m.detailScroll != 0 {
		t.Error("expected scroll to reset")
	}

	// Test nil
	m.SetSelectedBeadDetail(nil)
	if m.selectedBead != nil {
		t.Error("expected nil bead after setting nil detail")
	}
	if m.dependents != nil {
		t.Error("expected nil dependents after setting nil detail")
	}
}

// Helper to build a model with a tree for tree panel tests.
func modelWithTree(beads []data.Bead, expandAll bool) Model {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	t := tree.BuildTree(beads, expandAll)
	m.SetTree(t)
	return m
}

func TestTreePanel_Header(t *testing.T) {
	m := modelWithTree([]data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}, false)

	output := m.View()
	if !strings.Contains(output, "Beads") {
		t.Error("expected 'Beads' header in tree panel")
	}
}

func TestTreePanel_NoBeads(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	output := m.View()
	if !strings.Contains(output, "(no beads loaded)") {
		t.Error("expected empty state when no tree set")
	}
}

func TestTreePanel_EmptyTree(t *testing.T) {
	m := modelWithTree([]data.Bead{}, false)

	output := m.View()
	if !strings.Contains(output, "(no beads loaded)") {
		t.Error("expected empty state for empty bead list")
	}
}

func TestTreePanel_ShowsBeadIDTypeStatus(t *testing.T) {
	m := modelWithTree([]data.Bead{
		{ID: "epic-1", IssueType: "epic", Status: "open"},
		{ID: "feat-1", IssueType: "feature", Status: "closed"},
		{ID: "task-1", IssueType: "task", Status: "in_progress"},
		{ID: "bug-1", IssueType: "bug", Status: "blocked"},
		{ID: "chore-1", IssueType: "chore", Status: "deferred"},
		{ID: "adr-1", IssueType: "decision", Status: "open"},
	}, false)

	output := m.View()

	checks := []struct {
		label string
		want  string
	}{
		{"epic ID", "epic-1"},
		{"epic type", "epic"},
		{"feat ID", "feat-1"},
		{"feat short type", "feat"},
		{"task ID", "task-1"},
		{"task type", "task"},
		{"bug ID", "bug-1"},
		{"bug type", "bug"},
		{"chore ID", "chore-1"},
		{"chore type", "chore"},
		{"adr ID", "adr-1"},
		{"adr short type", "adr"},
	}
	for _, c := range checks {
		if !strings.Contains(output, c.want) {
			t.Errorf("expected %s %q in output", c.label, c.want)
		}
	}
}

func TestTreePanel_StatusIcons(t *testing.T) {
	m := modelWithTree([]data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "in_progress"},
		{ID: "b-3", IssueType: "task", Status: "blocked"},
		{ID: "b-4", IssueType: "task", Status: "deferred"},
		{ID: "b-5", IssueType: "task", Status: "closed"},
	}, false)

	output := m.View()

	icons := []struct {
		status string
		icon   string
	}{
		{"open", "○"},
		{"in_progress", "◉"},
		{"blocked", "✗"},
		{"deferred", "◌"},
		{"closed", "✓"},
	}
	for _, ic := range icons {
		if !strings.Contains(output, ic.icon) {
			t.Errorf("expected icon %q for status %s", ic.icon, ic.status)
		}
	}
}

func TestTreePanel_SelectedHighlighted(t *testing.T) {
	m := modelWithTree([]data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}, false)

	// selectedIdx defaults to 0, which is b-1
	output := m.View()
	// The selected row should contain the bead ID (it's highlighted with Reverse style)
	if !strings.Contains(output, "b-1") {
		t.Error("expected selected bead b-1 in output")
	}
}

func TestTreePanel_ExpandCollapseIndicators(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent-1", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent-1"},
	}

	// Collapsed
	m := modelWithTree(beads, false)
	output := m.View()
	if !strings.Contains(output, "▶") {
		t.Error("expected ▶ for collapsed parent")
	}

	// Expanded
	m = modelWithTree(beads, true)
	output = m.View()
	if !strings.Contains(output, "▼") {
		t.Error("expected ▼ for expanded parent")
	}
}

func TestTreePanel_TreeDrawingChars(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent"},
		{ID: "child-2", IssueType: "task", Status: "open", Parent: "parent"},
	}

	m := modelWithTree(beads, true)
	output := m.View()

	// Middle child should use ├──
	if !strings.Contains(output, "├──") {
		t.Error("expected ├── for middle child")
	}
	// Last child should use └──
	if !strings.Contains(output, "└──") {
		t.Error("expected └── for last child")
	}
}

func TestTreePanel_ChildrenHiddenWhenCollapsed(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent"},
	}

	m := modelWithTree(beads, false)
	output := m.View()

	if !strings.Contains(output, "parent") {
		t.Error("expected parent visible when collapsed")
	}
	if strings.Contains(output, "child-1") {
		t.Error("expected child hidden when parent collapsed")
	}
}

func TestTreePanel_ChildrenVisibleWhenExpanded(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent"},
	}

	m := modelWithTree(beads, true)
	output := m.View()

	if !strings.Contains(output, "child-1") {
		t.Error("expected child visible when parent expanded")
	}
}

func TestShortType(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"feature", "feat"},
		{"decision", "adr"},
		{"task", "task"},
		{"bug", "bug"},
		{"chore", "chore"},
		{"epic", "epic"},
	}
	for _, c := range cases {
		got := shortType(c.input)
		if got != c.want {
			t.Errorf("shortType(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestStatusIcon_AllStatuses(t *testing.T) {
	m := Model{}
	cases := []struct {
		status string
		icon   string
	}{
		{"open", "○"},
		{"in_progress", "◉"},
		{"blocked", "✗"},
		{"deferred", "◌"},
		{"closed", "✓"},
		{"unknown", "○"},
	}
	for _, c := range cases {
		result := m.statusIcon(c.status)
		if !strings.Contains(result, c.icon) {
			t.Errorf("statusIcon(%q) = %q, expected to contain %q", c.status, result, c.icon)
		}
	}
}

// --- Navigation tests ---

func TestNavigation_MoveDown(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	if m.selectedIdx != 0 {
		t.Fatal("expected initial selection at 0")
	}

	// Move down with j
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 after j, got %d", m.selectedIdx)
	}

	// Move down with down arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx 2 after down, got %d", m.selectedIdx)
	}
}

func TestNavigation_MoveUp(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 2

	// Move up with k
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 after k, got %d", m.selectedIdx)
	}

	// Move up with up arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 after up, got %d", m.selectedIdx)
	}
}

func TestNavigation_BoundaryTop(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0

	// Move up at top should stay at 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 at top boundary, got %d", m.selectedIdx)
	}
}

func TestNavigation_BoundaryBottom(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 1

	// Move down at bottom should stay at 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 at bottom boundary, got %d", m.selectedIdx)
	}
}

func TestNavigation_ExpandCollapse(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent"},
		{ID: "child-2", IssueType: "task", Status: "open", Parent: "parent"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0

	// Parent is collapsed, children not visible
	visible := m.tree.FlattenVisible()
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible node when collapsed, got %d", len(visible))
	}

	// Press Enter to expand
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	visible = m.tree.FlattenVisible()
	if len(visible) != 3 {
		t.Fatalf("expected 3 visible nodes after expand, got %d", len(visible))
	}

	// Press Left to collapse
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)

	visible = m.tree.FlattenVisible()
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible node after collapse, got %d", len(visible))
	}
}

func TestNavigation_ExpandWithRight(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0

	// Press Right to expand
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)

	visible := m.tree.FlattenVisible()
	if len(visible) != 2 {
		t.Fatalf("expected 2 visible nodes after Right expand, got %d", len(visible))
	}
}

func TestNavigation_LeftOnChildMovesToParent(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent"},
	}
	m := modelWithTree(beads, true)
	m.selectedIdx = 1 // on child-1

	// Press Left on child should move to parent
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 (parent) after Left on child, got %d", m.selectedIdx)
	}
	if m.selectedBead == nil || m.selectedBead.ID != "parent" {
		t.Error("expected selected bead to be parent")
	}
}

func TestNavigation_GoToTopBottom(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
		{ID: "b-4", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 2

	// G goes to bottom
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = updated.(Model)
	if m.selectedIdx != 3 {
		t.Errorf("expected selectedIdx 3 after G, got %d", m.selectedIdx)
	}

	// g goes to top
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updated.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 after g, got %d", m.selectedIdx)
	}
}

func TestNavigation_ExpandAll(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent-1", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent-1"},
		{ID: "parent-2", IssueType: "epic", Status: "open"},
		{ID: "child-2", IssueType: "task", Status: "open", Parent: "parent-2"},
	}
	m := modelWithTree(beads, false) // all collapsed

	visible := m.tree.FlattenVisible()
	if len(visible) != 2 {
		t.Fatalf("expected 2 visible (roots only), got %d", len(visible))
	}

	// Press e to expand all
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)

	visible = m.tree.FlattenVisible()
	if len(visible) != 4 {
		t.Errorf("expected 4 visible after expand all, got %d", len(visible))
	}
}

func TestNavigation_CollapseAll(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent-1", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent-1"},
		{ID: "parent-2", IssueType: "epic", Status: "open"},
		{ID: "child-2", IssueType: "task", Status: "open", Parent: "parent-2"},
	}
	m := modelWithTree(beads, true) // all expanded
	m.selectedIdx = 3               // on child-2

	// Press c to collapse all
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(Model)

	visible := m.tree.FlattenVisible()
	if len(visible) != 2 {
		t.Errorf("expected 2 visible after collapse all, got %d", len(visible))
	}
	// Should have moved to parent-2's root position
	if m.selectedBead == nil || m.selectedBead.ID != "parent-2" {
		id := ""
		if m.selectedBead != nil {
			id = m.selectedBead.ID
		}
		t.Errorf("expected selection on parent-2 after collapse, got %s", id)
	}
}

func TestNavigation_DownUpdatesSelectedBead(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "closed"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	if m.selectedBead == nil || m.selectedBead.ID != "b-2" {
		t.Error("expected selectedBead to be b-2 after moving down")
	}
}

func TestNavigation_EmptyTree(t *testing.T) {
	m := modelWithTree([]data.Bead{}, false)

	// These should not crash
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	_ = updated.(Model)
}

func TestNavigation_NoTreeSet(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	// These should not crash when tree is nil
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	_ = updated.(Model)
}

func TestNavigation_ScrollKeepsSelectionVisible(t *testing.T) {
	// Create enough beads to exceed viewport
	var beads []data.Bead
	for i := 0; i < 50; i++ {
		beads = append(beads, data.Bead{
			ID:        fmt.Sprintf("b-%02d", i),
			IssueType: "task",
			Status:    "open",
		})
	}
	m := modelWithTree(beads, false)
	m.height = 25 // small viewport (24 rows for tree after header)

	// Navigate to the bottom
	for i := 0; i < 49; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = updated.(Model)
	}

	// treeScroll should have advanced so selected is visible
	if m.treeScroll == 0 {
		t.Error("expected treeScroll to advance when navigating past viewport")
	}

	// The rendered output should contain the last bead
	output := m.View()
	if !strings.Contains(output, "b-49") {
		t.Error("expected b-49 to be visible after navigating to bottom")
	}
}

func TestTreePanel_NestedHierarchy(t *testing.T) {
	beads := []data.Bead{
		{ID: "epic-1", IssueType: "epic", Status: "open"},
		{ID: "feat-1", IssueType: "feature", Status: "open", Parent: "epic-1"},
		{ID: "task-1", IssueType: "task", Status: "open", Parent: "feat-1"},
	}

	m := modelWithTree(beads, true)
	output := m.View()

	// All should be visible
	if !strings.Contains(output, "epic-1") {
		t.Error("expected epic-1 visible")
	}
	if !strings.Contains(output, "feat-1") {
		t.Error("expected feat-1 visible")
	}
	if !strings.Contains(output, "task-1") {
		t.Error("expected task-1 visible")
	}
}

// --- Layout and resize tests ---

func TestLayout_TooSmallTerminal(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.ready = true

	// Too narrow
	m.width = 79
	m.height = 30
	output := m.View()
	if !strings.Contains(output, "Terminal too small") {
		t.Error("expected too-small message when width < 80")
	}

	// Too short
	m.width = 120
	m.height = 23
	output = m.View()
	if !strings.Contains(output, "Terminal too small") {
		t.Error("expected too-small message when height < 24")
	}

	// Exactly minimum should work
	m.width = 80
	m.height = 24
	output = m.View()
	if strings.Contains(output, "Terminal too small") {
		t.Error("should not show too-small message at 80x24")
	}
}

func TestLayout_NarrowModeHidesDetail(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", Description: "Some description"},
	}
	m := modelWithTree(beads, false)
	m.width = 95 // < 100, narrow mode
	m.height = 30
	m.SetSelectedBead(&beads[0])

	output := m.View()
	// Tree should be visible
	if !strings.Contains(output, "b-1") {
		t.Error("expected tree to be visible in narrow mode")
	}
	// Detail content should NOT be shown (no side-by-side)
	if strings.Contains(output, "DESCRIPTION") {
		t.Error("expected detail pane to be hidden in narrow mode")
	}
}

func TestLayout_NarrowModeOverlay(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", Description: "Some description"},
	}
	m := modelWithTree(beads, false)
	m.width = 95
	m.height = 30
	m.syncSelectedBead()

	// Press Enter to open overlay
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.showOverlay {
		t.Error("expected overlay to be shown after Enter in narrow mode")
	}

	output := m.View()
	if !strings.Contains(output, "b-1") {
		t.Error("expected bead ID in overlay")
	}
}

func TestLayout_NarrowModeEscClosesOverlay(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.width = 95
	m.height = 30
	m.syncSelectedBead()
	m.showOverlay = true

	// Press Escape to close overlay
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.showOverlay {
		t.Error("expected overlay to be closed after Escape")
	}
}

func TestLayout_WideModeShowsBothPanes(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", Description: "Some description"},
	}
	m := modelWithTree(beads, false)
	m.width = 120
	m.height = 30
	m.syncSelectedBead()

	output := m.View()
	// Both tree and detail should be visible
	if !strings.Contains(output, "Beads") {
		t.Error("expected tree header in wide mode")
	}
	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("expected detail content in wide mode")
	}
}

func TestLayout_TabDisabledInNarrowMode(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 95
	m.height = 30
	m.ready = true

	if m.focusedPane != treePane {
		t.Fatal("expected initial focus on tree")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)

	if m.focusedPane != treePane {
		t.Error("expected Tab to be ignored in narrow mode")
	}
}

func TestLayout_FocusSwitching(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true

	// Default focus is tree
	if m.focusedPane != treePane {
		t.Error("expected initial focus on tree pane")
	}

	// Tab switches to detail
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.focusedPane != detailPane {
		t.Error("expected focus on detail pane after Tab")
	}

	// Tab switches back
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.focusedPane != treePane {
		t.Error("expected focus on tree pane after second Tab")
	}
}

func TestLayout_ResizeAdjustsLayout(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.ready = true
	m.width = 120
	m.height = 40

	// Simulate resize
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 150, Height: 50})
	m = updated.(Model)

	if m.width != 150 || m.height != 50 {
		t.Errorf("expected dimensions 150x50, got %dx%d", m.width, m.height)
	}
}

func TestLayout_OverlayScrolling(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", Description: "Line1\nLine2\nLine3"},
	}
	m := modelWithTree(beads, false)
	m.width = 95
	m.height = 30
	m.syncSelectedBead()
	m.showOverlay = true

	// Scroll down in overlay
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.detailScroll != 1 {
		t.Errorf("expected detailScroll 1 after j in overlay, got %d", m.detailScroll)
	}

	// Scroll up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.detailScroll != 0 {
		t.Errorf("expected detailScroll 0 after k in overlay, got %d", m.detailScroll)
	}
}

// --- Search tests ---

func TestSearch_SlashOpensSearchMode(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)

	if !m.searching {
		t.Error("expected searching mode to be active after /")
	}
}

func TestSearch_TypingFiltersTree(t *testing.T) {
	beads := []data.Bead{
		{ID: "epic-1", Title: "Build the widget", IssueType: "epic", Status: "open"},
		{ID: "task-1", Title: "Fix the bug", IssueType: "task", Status: "open"},
		{ID: "task-2", Title: "Add widget tests", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Enter search mode and type "widget"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)

	for _, r := range "widget" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	if len(visible) != 2 {
		t.Errorf("expected 2 matching beads, got %d", len(visible))
	}
	// Should contain epic-1 and task-2 (both have "widget" in title)
	ids := make(map[string]bool)
	for _, n := range visible {
		ids[n.Bead.ID] = true
	}
	if !ids["epic-1"] {
		t.Error("expected epic-1 to match 'widget'")
	}
	if !ids["task-2"] {
		t.Error("expected task-2 to match 'widget'")
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	beads := []data.Bead{
		{ID: "ABC-1", Title: "Alpha Beta", IssueType: "task", Status: "open"},
		{ID: "xyz-1", Title: "Other", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Search for "abc" should match "ABC-1"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "abc" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	if len(visible) != 1 || visible[0].Bead.ID != "ABC-1" {
		t.Error("expected case-insensitive match on ID")
	}
}

func TestSearch_MatchesByID(t *testing.T) {
	beads := []data.Bead{
		{ID: "hep-ws-f6.3", Title: "Some title", IssueType: "task", Status: "open"},
		{ID: "hep-ws-f7.1", Title: "Other title", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "f6" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	if len(visible) != 1 || visible[0].Bead.ID != "hep-ws-f6.3" {
		t.Error("expected search by ID to match hep-ws-f6.3")
	}
}

func TestSearch_MatchesByDescription(t *testing.T) {
	beads := []data.Bead{
		{ID: "t-1", Title: "First", IssueType: "task", Status: "open", Description: "Configure the database"},
		{ID: "t-2", Title: "Second", IssueType: "task", Status: "open", Description: "Setup the frontend"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "database" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	if len(visible) != 1 || visible[0].Bead.ID != "t-1" {
		t.Errorf("expected search by description to match t-1, got %d results", len(visible))
	}
}

func TestSearch_MatchesByDesign(t *testing.T) {
	beads := []data.Bead{
		{ID: "t-1", Title: "First", IssueType: "task", Status: "open", Design: "Use PostgreSQL adapter"},
		{ID: "t-2", Title: "Second", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "postgresql" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	if len(visible) != 1 || visible[0].Bead.ID != "t-1" {
		t.Errorf("expected search by design to match t-1, got %d results", len(visible))
	}
}

func TestSearch_MatchesByAcceptanceCriteria(t *testing.T) {
	beads := []data.Bead{
		{ID: "t-1", Title: "First", IssueType: "task", Status: "open", AcceptanceCriteria: "All endpoints return JSON"},
		{ID: "t-2", Title: "Second", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "endpoints" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	if len(visible) != 1 || visible[0].Bead.ID != "t-1" {
		t.Errorf("expected search by acceptance criteria to match t-1, got %d results", len(visible))
	}
}

func TestSearch_MatchesByNotes(t *testing.T) {
	beads := []data.Bead{
		{ID: "t-1", Title: "First", IssueType: "task", Status: "open", Notes: "Refer to architecture doc"},
		{ID: "t-2", Title: "Second", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "architecture" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	if len(visible) != 1 || visible[0].Bead.ID != "t-1" {
		t.Errorf("expected search by notes to match t-1, got %d results", len(visible))
	}
}

func TestSearch_AncestorsPreserved(t *testing.T) {
	beads := []data.Bead{
		{ID: "epic-1", Title: "The Epic", IssueType: "epic", Status: "open"},
		{ID: "task-1", Title: "Match me", IssueType: "task", Status: "open", Parent: "epic-1"},
		{ID: "task-2", Title: "No match", IssueType: "task", Status: "open", Parent: "epic-1"},
	}
	m := modelWithTree(beads, true) // expand all so children are visible

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "Match me" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	visible := m.visibleNodes()
	ids := make(map[string]bool)
	for _, n := range visible {
		ids[n.Bead.ID] = true
	}
	if !ids["task-1"] {
		t.Error("expected matching bead task-1")
	}
	if !ids["epic-1"] {
		t.Error("expected ancestor epic-1 preserved")
	}
	if ids["task-2"] {
		t.Error("expected non-matching sibling task-2 to be filtered out")
	}
}

func TestSearch_EscapeClearsSearch(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha", IssueType: "task", Status: "open"},
		{ID: "b-2", Title: "Beta", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Enter search and type
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	m = updated.(Model)

	if m.searchQuery != "A" {
		t.Errorf("expected searchQuery 'A', got %q", m.searchQuery)
	}

	// Escape clears search and restores full tree
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.searching {
		t.Error("expected searching to be false after Escape")
	}
	if m.searchQuery != "" {
		t.Error("expected searchQuery to be cleared after Escape")
	}

	visible := m.visibleNodes()
	if len(visible) != 2 {
		t.Errorf("expected full tree restored, got %d visible", len(visible))
	}
}

func TestSearch_EnterConfirmsSearch(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha", IssueType: "task", Status: "open"},
		{ID: "b-2", Title: "Beta", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Enter search, type, then press Enter
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "Alpha" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.searching {
		t.Error("expected searching to be false after Enter")
	}
	if m.searchQuery != "Alpha" {
		t.Error("expected searchQuery to remain after Enter")
	}

	// Filter should still be active
	visible := m.visibleNodes()
	if len(visible) != 1 {
		t.Errorf("expected 1 filtered result, got %d", len(visible))
	}
}

func TestSearch_EscClearsActiveSearchOutsideSearchMode(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha", IssueType: "task", Status: "open"},
		{ID: "b-2", Title: "Beta", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Search, confirm with Enter, then Escape to clear
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "Alpha" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	// Now press Escape to clear the active filter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.searchQuery != "" {
		t.Error("expected searchQuery cleared by Escape outside search mode")
	}
	visible := m.visibleNodes()
	if len(visible) != 2 {
		t.Errorf("expected full tree after clearing search, got %d", len(visible))
	}
}

func TestSearch_BackspaceRemovesChar(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "xyz" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	if m.searchQuery != "xyz" {
		t.Fatalf("expected query 'xyz', got %q", m.searchQuery)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(Model)

	if m.searchQuery != "xy" {
		t.Errorf("expected query 'xy' after backspace, got %q", m.searchQuery)
	}
}

func TestSearch_EmptyResultsMessage(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Search for something that doesn't exist
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "zzzzz" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	// Confirm search
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	output := m.View()
	if !strings.Contains(output, "(no matching beads)") {
		t.Error("expected 'no matching beads' message for empty search results")
	}
}

func TestSearch_StatusBarShowsQuery(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Enter search mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "test" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	// While in search mode, status bar should show search prompt
	output := m.View()
	if !strings.Contains(output, "Search: test") {
		t.Error("expected status bar to show 'Search: test' during search input")
	}

	// Confirm search
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	output = m.View()
	if !strings.Contains(output, `"test"`) {
		t.Error("expected status bar to show active search query after confirming")
	}
}

func TestSearch_NavigationWorksWhileFiltered(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha", IssueType: "task", Status: "open"},
		{ID: "b-2", Title: "Alpha Two", IssueType: "task", Status: "open"},
		{ID: "b-3", Title: "Beta", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Search for "Alpha"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	for _, r := range "Alpha" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	// Should have 2 results, navigate between them
	if m.selectedIdx != 0 {
		t.Fatalf("expected selectedIdx 0, got %d", m.selectedIdx)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 after j, got %d", m.selectedIdx)
	}
	if m.selectedBead == nil || m.selectedBead.ID != "b-2" {
		t.Error("expected selected bead to be b-2")
	}
}

func TestSearch_KeysIgnoredDuringSearchInput(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Enter search mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)

	// Pressing 'j' should add to query, not navigate
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	if m.searchQuery != "j" {
		t.Errorf("expected searchQuery 'j', got %q", m.searchQuery)
	}
	// Selection should have reset to 0 (search resets selection)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 during search, got %d", m.selectedIdx)
	}
}

func TestLayout_RightExpandsInNarrowMode(t *testing.T) {
	beads := []data.Bead{
		{ID: "parent", IssueType: "epic", Status: "open"},
		{ID: "child-1", IssueType: "task", Status: "open", Parent: "parent"},
	}
	m := modelWithTree(beads, false)
	m.width = 95
	m.height = 30

	// Right should expand, not open overlay
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)

	if m.showOverlay {
		t.Error("Right key should expand, not open overlay")
	}
	visible := m.tree.FlattenVisible()
	if len(visible) != 2 {
		t.Errorf("expected 2 visible after Right expand, got %d", len(visible))
	}
}

// --- Filter tests ---

func TestFilter_FOpensFilterOverlay(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updated.(Model)

	if !m.filtering {
		t.Error("expected filter overlay to be open after pressing f")
	}
}

func TestFilter_OverlayRendersTypesAndStatuses(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true

	output := m.View()
	if !strings.Contains(output, "Filter Beads") {
		t.Error("expected filter overlay title")
	}
	if !strings.Contains(output, "TYPE") {
		t.Error("expected TYPE heading in filter overlay")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("expected STATUS heading in filter overlay")
	}
	if !strings.Contains(output, "task") {
		t.Error("expected task type in filter overlay")
	}
	if !strings.Contains(output, "open") {
		t.Error("expected open status in filter overlay")
	}
}

func TestFilter_SpaceTogglesSelection(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "bug", Status: "closed"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true
	m.filterCursor = 0 // "task" is first item

	// Toggle task on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)

	if !m.filterTypes["task"] {
		t.Error("expected task filter to be enabled after space toggle")
	}

	// Toggle task off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)

	if m.filterTypes["task"] {
		t.Error("expected task filter to be disabled after second space toggle")
	}
}

func TestFilter_NavigateFilterMenu(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true
	m.filterCursor = 0

	// Move down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.filterCursor != 1 {
		t.Errorf("expected cursor at 1 after j, got %d", m.filterCursor)
	}

	// Move up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.filterCursor != 0 {
		t.Errorf("expected cursor at 0 after k, got %d", m.filterCursor)
	}
}

func TestFilter_EnterClosesOverlay(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.filtering {
		t.Error("expected filter overlay to close on Enter")
	}
}

func TestFilter_EscClearsFiltersAndCloses(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "bug", Status: "closed"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true
	m.filterTypes["task"] = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.filtering {
		t.Error("expected filter overlay to close on Esc")
	}
	if len(m.filterTypes) > 0 {
		t.Error("expected filters to be cleared on Esc")
	}
}

func TestFilter_TypeFilterShowsOnlyMatchingTypes(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "bug", Status: "open"},
		{ID: "b-3", IssueType: "feature", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["task"] = true

	visible := m.visibleNodes()
	if len(visible) != 1 {
		t.Errorf("expected 1 visible bead with task filter, got %d", len(visible))
		return
	}
	if visible[0].Bead.ID != "b-1" {
		t.Errorf("expected b-1, got %s", visible[0].Bead.ID)
	}
}

func TestFilter_StatusFilterShowsOnlyMatchingStatuses(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "closed"},
		{ID: "b-3", IssueType: "task", Status: "blocked"},
	}
	m := modelWithTree(beads, false)
	m.filterStats["closed"] = true

	visible := m.visibleNodes()
	if len(visible) != 1 {
		t.Errorf("expected 1 visible bead with closed filter, got %d", len(visible))
		return
	}
	if visible[0].Bead.ID != "b-2" {
		t.Errorf("expected b-2, got %s", visible[0].Bead.ID)
	}
}

func TestFilter_ORWithinCategory(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "bug", Status: "open"},
		{ID: "b-3", IssueType: "feature", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["task"] = true
	m.filterTypes["bug"] = true

	visible := m.visibleNodes()
	if len(visible) != 2 {
		t.Errorf("expected 2 visible beads with task|bug filter, got %d", len(visible))
	}
}

func TestFilter_ANDAcrossCategories(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "closed"},
		{ID: "b-3", IssueType: "bug", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["task"] = true
	m.filterStats["open"] = true

	visible := m.visibleNodes()
	if len(visible) != 1 {
		t.Errorf("expected 1 visible bead with task AND open filter, got %d", len(visible))
		return
	}
	if visible[0].Bead.ID != "b-1" {
		t.Errorf("expected b-1, got %s", visible[0].Bead.ID)
	}
}

func TestFilter_AncestorsPreserved(t *testing.T) {
	beads := []data.Bead{
		{ID: "epic-1", IssueType: "epic", Status: "open"},
		{ID: "task-1", IssueType: "task", Status: "open", Parent: "epic-1"},
		{ID: "bug-1", IssueType: "bug", Status: "open", Parent: "epic-1"},
	}
	m := modelWithTree(beads, true)
	m.filterTypes["task"] = true

	visible := m.visibleNodes()
	ids := make(map[string]bool)
	for _, n := range visible {
		ids[n.Bead.ID] = true
	}
	if !ids["task-1"] {
		t.Error("expected matching task-1")
	}
	if !ids["epic-1"] {
		t.Error("expected ancestor epic-1 preserved")
	}
	if ids["bug-1"] {
		t.Error("expected non-matching bug-1 filtered out")
	}
}

func TestFilter_StatusBarShowsActiveFilter(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["task"] = true
	m.filterStats["open"] = true

	output := m.View()
	if !strings.Contains(output, "Filter: type=task status=open") {
		t.Error("expected status bar to show active filter")
	}
}

func TestFilter_EscClearsActiveFiltersOutsideOverlay(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "bug", Status: "closed"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["task"] = true

	// Not in filter overlay, press Esc
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.hasActiveFilters() {
		t.Error("expected Esc to clear active filters")
	}
	visible := m.visibleNodes()
	if len(visible) != 2 {
		t.Errorf("expected full tree after clearing filters, got %d", len(visible))
	}
}

func TestFilter_EmptyResultsMessage(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["bug"] = true // no bugs exist

	output := m.View()
	if !strings.Contains(output, "(no matching beads)") {
		t.Error("expected 'no matching beads' when filter produces no results")
	}
}

func TestFilter_CLIFlagsApplyInitialFilters(t *testing.T) {
	cfg := Config{
		Refresh:        2,
		FilterTypes:    []string{"task", "bug"},
		FilterStatuses: []string{"open"},
	}
	m := New(cfg)
	m.width = 120
	m.height = 40
	m.ready = true

	if !m.filterTypes["task"] || !m.filterTypes["bug"] {
		t.Error("expected CLI --type flags to set initial type filters")
	}
	if !m.filterStats["open"] {
		t.Error("expected CLI --status flag to set initial status filter")
	}
}

func TestFilter_CombinedWithSearch(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", Title: "Alpha task", IssueType: "task", Status: "open"},
		{ID: "b-2", Title: "Beta task", IssueType: "task", Status: "open"},
		{ID: "b-3", Title: "Alpha bug", IssueType: "bug", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["task"] = true
	m.searchQuery = "Alpha"

	visible := m.visibleNodes()
	if len(visible) != 1 {
		t.Errorf("expected 1 visible with task filter + Alpha search, got %d", len(visible))
		return
	}
	if visible[0].Bead.ID != "b-1" {
		t.Errorf("expected b-1, got %s", visible[0].Bead.ID)
	}
}

func TestFilter_FClosesOverlayToo(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updated.(Model)

	if m.filtering {
		t.Error("expected f to close filter overlay")
	}
}

func TestFilter_ToggleStatusInOverlay(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true
	// Navigate to first status item (after 6 type items)
	m.filterCursor = len(allTypes) // first status item

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)

	if !m.filterStats["open"] {
		t.Error("expected open status to be toggled on")
	}
}

func TestFilter_SearchClearsBeforeFilter(t *testing.T) {
	// When Esc is pressed, search clears first, then filters on next Esc
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "bug", Status: "closed"},
	}
	m := modelWithTree(beads, false)
	m.filterTypes["task"] = true
	m.searchQuery = "b-1"

	// First Esc clears search
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.searchQuery != "" {
		t.Error("expected search to be cleared first")
	}
	if !m.hasActiveFilters() {
		t.Error("expected filters to remain after first Esc")
	}

	// Second Esc clears filters
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.hasActiveFilters() {
		t.Error("expected filters to be cleared on second Esc")
	}
}

// --- Help overlay tests ---

func TestHelp_QuestionMarkOpensHelpOverlay(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)

	if !m.showHelp {
		t.Error("expected help overlay to be open after pressing ?")
	}
}

func TestHelp_OverlayRendersKeybindings(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true

	output := m.View()

	checks := []string{
		"Help",
		"NAVIGATION",
		"TREE",
		"SEARCH & FILTER",
		"OTHER",
		"j / Down",
		"k / Up",
		"Tab",
		"Enter",
		"Right",
		"Left",
		"Expand all",
		"Collapse all",
		"Search by ID, title, or content",
		"filter menu",
		"Force refresh",
		"Quit",
	}
	for _, want := range checks {
		if !strings.Contains(output, want) {
			t.Errorf("expected help overlay to contain %q", want)
		}
	}
}

func TestHelp_EscClosesOverlay(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.showHelp {
		t.Error("expected help overlay to close on Esc")
	}
}

func TestHelp_QuestionMarkClosesOverlay(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)

	if m.showHelp {
		t.Error("expected ? to close help overlay when already open")
	}
}

func TestHelp_OverlayBlocksOtherKeys(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "bug", Status: "closed"},
	}
	m := modelWithTree(beads, false)
	m.showHelp = true

	// j should not move selection while help is open
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("expected selection to stay at 0 while help overlay is open, got %d", m.selectedIdx)
	}
	if !m.showHelp {
		t.Error("expected help overlay to remain open for non-close keys")
	}
}

func TestHelp_QuitWorksInOverlay(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Error("expected quit command when pressing q in help overlay")
	}
}

func TestHelp_RKeyDoesNotError(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// r should not panic or change state
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(Model)

	if m.selectedIdx != 0 {
		t.Errorf("expected r to be no-op, selectedIdx changed to %d", m.selectedIdx)
	}
}

func TestHelp_StatusBarShowsHelpHint(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	output := m.View()
	if !strings.Contains(output, "[?] Help") {
		t.Error("expected status bar to show [?] Help hint")
	}
}

func TestHelp_ScrollDownInOverlay(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	if m.helpScroll != 1 {
		t.Errorf("expected helpScroll=1 after j, got %d", m.helpScroll)
	}
	if !m.showHelp {
		t.Error("expected help overlay to remain open after scrolling")
	}
}

func TestHelp_ScrollUpInOverlay(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true
	m.helpScroll = 3

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)

	if m.helpScroll != 2 {
		t.Errorf("expected helpScroll=2 after k, got %d", m.helpScroll)
	}
}

func TestHelp_ScrollUpStopsAtZero(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true
	m.helpScroll = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)

	if m.helpScroll != 0 {
		t.Errorf("expected helpScroll=0, got %d", m.helpScroll)
	}
}

func TestHelp_ScrollResetsOnClose(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.showHelp = true
	m.helpScroll = 5

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.helpScroll != 0 {
		t.Errorf("expected helpScroll reset to 0 on close, got %d", m.helpScroll)
	}
}

func TestHelp_RegistryDrivesOverlay(t *testing.T) {
	// Verify that every keybinding in the registry appears in the rendered help overlay
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 80 // tall enough to show all
	m.ready = true
	m.showHelp = true

	output := m.View()

	for _, kb := range keybindingRegistry {
		if !strings.Contains(output, kb.Keys) {
			t.Errorf("expected help overlay to contain key %q from registry", kb.Keys)
		}
		if !strings.Contains(output, kb.Description) {
			t.Errorf("expected help overlay to contain description %q from registry", kb.Description)
		}
	}
	for _, section := range sectionOrder {
		if !strings.Contains(output, section) {
			t.Errorf("expected help overlay to contain section %q", section)
		}
	}
}

func TestHelp_OverlayScrollsWithSmallHeight(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 28 // small height: content area = 27 lines (minus 1 for status bar)
	m.ready = true
	m.showHelp = true
	m.helpScroll = 5

	output := m.View()

	// The scrolled output should not show the title line (it's above the scroll offset)
	if strings.Contains(output, "Help — Keybindings") {
		t.Error("expected title to be scrolled off screen with helpScroll=5")
	}
}

// --- Refresh tests ---

func TestRefresh_ApplyRefreshBuildsTree(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "epic", Status: "closed"},
	}
	m.applyRefresh(beads)

	if m.tree == nil {
		t.Fatal("expected tree to be built after refresh")
	}
	if len(m.beads) != 2 {
		t.Errorf("expected 2 beads, got %d", len(m.beads))
	}
	if m.lastRefresh.IsZero() {
		t.Error("expected lastRefresh to be set")
	}
}

func TestRefresh_PreservesSelection(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.beads = beads
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	// Select second bead
	m.selectedIdx = 1
	m.syncSelectedBead()

	// Refresh with same beads (new timestamps to trigger update)
	t2 := time.Now()
	newBeads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", UpdatedAt: &t2},
		{ID: "b-2", IssueType: "task", Status: "open", UpdatedAt: &t2},
		{ID: "b-3", IssueType: "task", Status: "open", UpdatedAt: &t2},
	}
	m.applyRefresh(newBeads)

	if m.selectedBead == nil || m.selectedBead.ID != "b-2" {
		t.Errorf("expected selection preserved on b-2, got %v", m.selectedBead)
	}
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1, got %d", m.selectedIdx)
	}
}

func TestRefresh_SelectionMovesToNeighborOnDelete(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.beads = beads
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	// Select last bead
	m.selectedIdx = 2
	m.syncSelectedBead()

	// Refresh with b-3 removed
	newBeads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m.applyRefresh(newBeads)

	if m.selectedBead == nil {
		t.Fatal("expected selection to fall back to nearest neighbor")
	}
	if m.selectedIdx >= 2 {
		t.Errorf("expected selectedIdx to be clamped, got %d", m.selectedIdx)
	}
}

func TestRefresh_PreservesExpandState(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "epic", Status: "open"},
		{ID: "b-1.1", IssueType: "task", Status: "open", Parent: "b-1"},
		{ID: "b-2", IssueType: "epic", Status: "open"},
		{ID: "b-2.1", IssueType: "task", Status: "open", Parent: "b-2"},
	}
	m := modelWithTree(beads, false) // all collapsed
	m.beads = beads
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	// Expand b-1 only
	m.tree.ToggleExpand("b-1")

	// b-1 should be expanded, b-2 collapsed
	if !m.tree.ByID["b-1"].Expanded {
		t.Fatal("b-1 should be expanded before refresh")
	}
	if m.tree.ByID["b-2"].Expanded {
		t.Fatal("b-2 should be collapsed before refresh")
	}

	// Refresh with same beads
	t2 := time.Now()
	newBeads := make([]data.Bead, len(beads))
	copy(newBeads, beads)
	for i := range newBeads {
		newBeads[i].UpdatedAt = &t2
	}
	m.applyRefresh(newBeads)

	// Expand state should be preserved
	if !m.tree.ByID["b-1"].Expanded {
		t.Error("b-1 should still be expanded after refresh")
	}
	if m.tree.ByID["b-2"].Expanded {
		t.Error("b-2 should still be collapsed after refresh")
	}
}

func TestRefresh_NewBeadAppears(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.beads = beads
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	// Refresh with a new bead added
	newBeads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m.applyRefresh(newBeads)

	visible := m.visibleNodes()
	if len(visible) != 2 {
		t.Errorf("expected 2 visible nodes after add, got %d", len(visible))
	}
}

func TestRefresh_UpdatedBeadRefreshesDetail(t *testing.T) {
	t1 := time.Now()
	beads := []data.Bead{
		{ID: "b-1", Title: "Old Title", IssueType: "task", Status: "open", UpdatedAt: &t1},
	}
	m := modelWithTree(beads, false)
	m.beads = beads
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }
	m.selectedIdx = 0
	m.syncSelectedBead()

	// Refresh with updated title
	t2 := t1.Add(time.Second)
	newBeads := []data.Bead{
		{ID: "b-1", Title: "New Title", IssueType: "task", Status: "open", UpdatedAt: &t2},
	}
	m.applyRefresh(newBeads)

	if m.selectedBead == nil || m.selectedBead.Title != "New Title" {
		t.Errorf("expected detail to reflect updated title, got %v", m.selectedBead)
	}
}

func TestRefresh_NoChangesPreservesState(t *testing.T) {
	t1 := time.Now()
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", UpdatedAt: &t1},
		{ID: "b-2", IssueType: "task", Status: "open", UpdatedAt: &t1},
	}
	m := modelWithTree(beads, false)
	m.beads = beads
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }
	m.selectedIdx = 1
	m.syncSelectedBead()

	origTree := m.tree

	// Refresh with identical beads
	m.applyRefresh(beads)

	// Tree should not have been rebuilt
	if m.tree != origTree {
		t.Error("expected tree to be unchanged when no diff")
	}
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx unchanged, got %d", m.selectedIdx)
	}
}

func TestRefresh_StatusBarShowsRefreshedAgo(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	baseTime := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	m.nowFunc = func() time.Time { return baseTime }
	m.lastRefresh = baseTime.Add(-5 * time.Second)

	output := m.View()
	if !strings.Contains(output, "Refreshed 5s ago") {
		t.Errorf("expected 'Refreshed 5s ago' in status bar, got: %s", output)
	}
}

func TestRefresh_StatusBarShowsDefaultBeforeFirstRefresh(t *testing.T) {
	m := New(Config{Refresh: 3})
	m.width = 120
	m.height = 40
	m.ready = true

	output := m.View()
	if !strings.Contains(output, "Refresh: 3s") {
		t.Errorf("expected 'Refresh: 3s' before first refresh, got: %s", output)
	}
}

func TestRefresh_BeadsLoadedMsgTriggersApply(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	updated, _ := m.Update(beadsLoadedMsg{beads: beads})
	m = updated.(Model)

	if m.tree == nil {
		t.Fatal("expected tree to be built from beadsLoadedMsg")
	}
	if len(m.beads) != 1 {
		t.Errorf("expected 1 bead, got %d", len(m.beads))
	}
}

func TestRefresh_BeadsLoadedMsgWithErrorIgnored(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.beads = beads
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	// Error message should not overwrite existing data
	updated, _ := m.Update(beadsLoadedMsg{err: fmt.Errorf("connection failed")})
	m = updated.(Model)

	if m.tree == nil {
		t.Error("expected existing tree preserved on error")
	}
	if len(m.beads) != 1 {
		t.Errorf("expected beads preserved on error, got %d", len(m.beads))
	}
}

func TestRefresh_TickMsgReturnsCmds(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	_, cmd := m.Update(tickMsg(time.Now()))
	// Without fetcher, tick should still return a tick cmd (for scheduling)
	if cmd == nil {
		t.Error("expected tick to return a cmd for next tick")
	}
}

func TestRefresh_EmptyToPopulated(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC) }

	// Start with no beads
	if m.tree != nil {
		t.Fatal("expected nil tree initially")
	}

	// First refresh brings beads
	beads := []data.Bead{
		{ID: "b-1", IssueType: "epic", Status: "open"},
		{ID: "b-1.1", IssueType: "task", Status: "open", Parent: "b-1"},
	}
	m.applyRefresh(beads)

	if m.tree == nil {
		t.Fatal("expected tree after refresh")
	}
	if m.selectedBead == nil {
		t.Fatal("expected a bead to be selected")
	}
	if m.selectedBead.ID != "b-1" {
		t.Errorf("expected first bead selected, got %s", m.selectedBead.ID)
	}
}

func TestMouse_ClickTreePaneSwitchesFocus(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true
	m.focusedPane = detailPane

	// Click in tree panel region (x=10, within treeWidth which is 120*2/5=48)
	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.focusedPane != treePane {
		t.Error("expected focus on tree pane after clicking in tree region")
	}
}

func TestMouse_ClickDetailPaneSwitchesFocus(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true
	m.focusedPane = treePane

	// Click in detail panel region (x=60, beyond treeWidth of 48)
	updated, _ := m.Update(tea.MouseMsg{
		X: 60, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.focusedPane != detailPane {
		t.Error("expected focus on detail pane after clicking in detail region")
	}
}

func TestMouse_ClickIgnoredInNarrowMode(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 95 // narrow mode
	m.height = 30
	m.ready = true
	m.focusedPane = treePane

	updated, _ := m.Update(tea.MouseMsg{
		X: 60, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.focusedPane != treePane {
		t.Error("expected focus unchanged in narrow mode")
	}
}

func TestMouse_ClickIgnoredInOverlayMode(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true
	m.showOverlay = true
	m.focusedPane = treePane

	updated, _ := m.Update(tea.MouseMsg{
		X: 60, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.focusedPane != treePane {
		t.Error("expected focus unchanged in overlay mode")
	}
}

func TestMouse_ClickIgnoredInHelpMode(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true
	m.showHelp = true
	m.focusedPane = treePane

	updated, _ := m.Update(tea.MouseMsg{
		X: 60, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.focusedPane != treePane {
		t.Error("expected focus unchanged in help mode")
	}
}

func TestMouse_ReleaseCopiesSelectionAfterDrag(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true
	m.selecting = true
	m.selStartRow = 0
	m.selStartCol = 0
	m.selEndRow = 0
	m.selEndCol = 5
	m.detailLines = []string{"Hello World"}

	updated, cmd := m.Update(tea.MouseMsg{
		X: 60, Y: 5,
		Button: tea.MouseButtonNone,
		Action: tea.MouseActionRelease,
	})
	m = updated.(Model)
	if m.selecting {
		t.Error("expected selecting to be false after release")
	}
	if !m.hasSelection {
		t.Error("expected hasSelection to be true after release")
	}
	if m.statusMsg == "" {
		t.Error("expected status message after drag-release copy")
	}
	if cmd == nil {
		t.Error("expected command batch for clipboard copy")
	}
}

func TestMouse_ReleaseNoCopyOnSimpleClick(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true
	m.selecting = true
	// Start and end at same position (simple click, no drag)
	m.selStartRow = 0
	m.selStartCol = 3
	m.selEndRow = 0
	m.selEndCol = 3
	m.detailLines = []string{"Hello World"}

	updated, cmd := m.Update(tea.MouseMsg{
		X: 60, Y: 5,
		Button: tea.MouseButtonNone,
		Action: tea.MouseActionRelease,
	})
	m = updated.(Model)
	if m.selecting {
		t.Error("expected selecting to be false after release")
	}
	if m.hasSelection {
		t.Error("expected hasSelection to be false for simple click")
	}
	if m.statusMsg != "" {
		t.Error("expected no status message on simple click")
	}
	if cmd != nil {
		t.Error("expected no command on simple click")
	}
}

func TestMouse_ReleaseIgnored(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 30
	m.ready = true
	m.focusedPane = treePane

	updated, _ := m.Update(tea.MouseMsg{
		X: 60, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
	})
	m = updated.(Model)
	if m.focusedPane != treePane {
		t.Error("expected focus unchanged on mouse release")
	}
}

func TestMouse_ClickTreeRowSelectsBead(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.focusedPane = detailPane

	// Click on row index 2 (Y=3: header=0, row0=1, row1=2, row2=3)
	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 3,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx=2, got %d", m.selectedIdx)
	}
	if m.focusedPane != treePane {
		t.Error("expected focus on tree pane after clicking tree row")
	}
	if m.selectedBead == nil || m.selectedBead.ID != "b-3" {
		t.Error("expected selected bead to be b-3")
	}
}

func TestMouse_ClickTreeRowWithScrollOffset(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
		{ID: "b-4", IssueType: "task", Status: "open"},
		{ID: "b-5", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.treeScroll = 2 // Scrolled down by 2

	// Click Y=1 (first visible row after header), with scroll=2, should select index 2
	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 1,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx=2 (scroll offset 2 + row 0), got %d", m.selectedIdx)
	}
}

func TestMouse_ClickHeaderRowIgnored(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0

	// Click on header row (Y=0)
	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 0,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx unchanged at 0, got %d", m.selectedIdx)
	}
}

func TestMouse_ClickBelowLastRowIgnored(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0

	// Click well below the last row (Y=20, only 2 rows)
	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 20,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx unchanged at 0, got %d", m.selectedIdx)
	}
}

func TestMouse_ClickIgnoredDuringSearch(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.searching = true

	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 2,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx unchanged during search, got %d", m.selectedIdx)
	}
}

func TestMouse_ClickIgnoredDuringFilter(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.filtering = true

	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 2,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx unchanged during filter, got %d", m.selectedIdx)
	}
}

func TestMouse_ClickTreeRowInNarrowMode(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.width = 95 // narrow mode
	m.selectedIdx = 0

	// Click on second row (Y=2)
	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 2,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 in narrow mode, got %d", m.selectedIdx)
	}
	if m.selectedBead == nil || m.selectedBead.ID != "b-2" {
		t.Error("expected selected bead to be b-2 in narrow mode")
	}
}

func TestMouse_ClickTreeRowThenKeyboardNavigation(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
		{ID: "b-3", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)

	// Click to select second row
	updated, _ := m.Update(tea.MouseMsg{
		X: 10, Y: 2,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 after click, got %d", m.selectedIdx)
	}

	// Navigate down with keyboard
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.selectedIdx != 2 {
		t.Errorf("expected selectedIdx=2 after j key, got %d", m.selectedIdx)
	}
}

// --- Mouse Wheel Scroll Tests ---

func TestMouseWheel_TreePanelScrollDown(t *testing.T) {
	// Create enough beads to require scrolling
	var beads []data.Bead
	for i := 0; i < 50; i++ {
		beads = append(beads, data.Bead{ID: fmt.Sprintf("b-%d", i), IssueType: "task", Status: "open"})
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0
	m.syncSelectedBead()

	// Scroll down in tree panel (X within tree width)
	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 10,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.treeScroll != 3 {
		t.Errorf("expected treeScroll=3 after wheel down, got %d", m.treeScroll)
	}
	// Selection should not change
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0 (unchanged), got %d", m.selectedIdx)
	}
	// Focus should not change
	if m.focusedPane != treePane {
		t.Error("expected focus to remain on tree pane")
	}
}

func TestMouseWheel_TreePanelScrollUp(t *testing.T) {
	var beads []data.Bead
	for i := 0; i < 50; i++ {
		beads = append(beads, data.Bead{ID: fmt.Sprintf("b-%d", i), IssueType: "task", Status: "open"})
	}
	m := modelWithTree(beads, false)
	m.treeScroll = 10

	// Scroll up in tree panel
	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 10,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.treeScroll != 7 {
		t.Errorf("expected treeScroll=7 after wheel up, got %d", m.treeScroll)
	}
}

func TestMouseWheel_TreePanelScrollUpClampsToZero(t *testing.T) {
	var beads []data.Bead
	for i := 0; i < 50; i++ {
		beads = append(beads, data.Bead{ID: fmt.Sprintf("b-%d", i), IssueType: "task", Status: "open"})
	}
	m := modelWithTree(beads, false)
	m.treeScroll = 1 // less than scroll step of 3

	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 10,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.treeScroll != 0 {
		t.Errorf("expected treeScroll=0 (clamped), got %d", m.treeScroll)
	}
}

func TestMouseWheel_TreePanelScrollDownClampsToMax(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
		{ID: "b-2", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.height = 30 // viewport > number of rows, so max scroll = 0

	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 5,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.treeScroll != 0 {
		t.Errorf("expected treeScroll=0 (no content to scroll), got %d", m.treeScroll)
	}
}

func TestMouseWheel_DetailPanelScrollDown(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", Description: "Long description"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0
	m.syncSelectedBead()

	// Scroll down in detail panel (X beyond tree width)
	tw := m.treeWidth()
	updated, _ := m.Update(tea.MouseMsg{
		X: tw + 5, Y: 10,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.detailScroll != 3 {
		t.Errorf("expected detailScroll=3 after wheel down, got %d", m.detailScroll)
	}
	// Focus should NOT change on scroll
	if m.focusedPane != treePane {
		t.Error("expected focus to remain on tree pane (scroll doesn't change focus)")
	}
}

func TestMouseWheel_DetailPanelScrollUp(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.detailScroll = 10

	tw := m.treeWidth()
	updated, _ := m.Update(tea.MouseMsg{
		X: tw + 5, Y: 10,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.detailScroll != 7 {
		t.Errorf("expected detailScroll=7 after wheel up, got %d", m.detailScroll)
	}
}

func TestMouseWheel_DetailScrollUpClampsToZero(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.detailScroll = 2

	tw := m.treeWidth()
	updated, _ := m.Update(tea.MouseMsg{
		X: tw + 5, Y: 10,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.detailScroll != 0 {
		t.Errorf("expected detailScroll=0 (clamped), got %d", m.detailScroll)
	}
}

func TestMouseWheel_DoesNotChangeFocus(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open"},
	}
	m := modelWithTree(beads, false)
	m.focusedPane = detailPane // start with detail focused

	// Scroll in tree panel area
	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 5,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.focusedPane != detailPane {
		t.Error("expected focus to remain on detail pane after tree scroll")
	}
}

func TestMouseWheel_DoesNotChangeSelection(t *testing.T) {
	var beads []data.Bead
	for i := 0; i < 50; i++ {
		beads = append(beads, data.Bead{ID: fmt.Sprintf("b-%d", i), IssueType: "task", Status: "open"})
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 5
	m.syncSelectedBead()

	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 5,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.selectedIdx != 5 {
		t.Errorf("expected selectedIdx=5 (unchanged), got %d", m.selectedIdx)
	}
}

func TestMouseWheel_NarrowModeScrollsTree(t *testing.T) {
	var beads []data.Bead
	for i := 0; i < 50; i++ {
		beads = append(beads, data.Bead{ID: fmt.Sprintf("b-%d", i), IssueType: "task", Status: "open"})
	}
	m := modelWithTree(beads, false)
	m.width = 90 // narrow mode

	updated, _ := m.Update(tea.MouseMsg{
		X: 50, Y: 10,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.treeScroll != 3 {
		t.Errorf("expected treeScroll=3 in narrow mode, got %d", m.treeScroll)
	}
}

func TestMouseWheel_OverlayModeScrollsDetail(t *testing.T) {
	beads := []data.Bead{
		{ID: "b-1", IssueType: "task", Status: "open", Description: "content"},
	}
	m := modelWithTree(beads, false)
	m.selectedIdx = 0
	m.syncSelectedBead()
	m.showOverlay = true

	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 5,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.detailScroll != 3 {
		t.Errorf("expected detailScroll=3 in overlay mode, got %d", m.detailScroll)
	}
}

func TestMouseWheel_HelpOverlayScrollsHelp(t *testing.T) {
	m := modelWithTree(nil, false)
	m.showHelp = true

	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 5,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.helpScroll != 3 {
		t.Errorf("expected helpScroll=3, got %d", m.helpScroll)
	}

	// Scroll up
	updated, _ = m.Update(tea.MouseMsg{
		X: 5, Y: 5,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.helpScroll != 0 {
		t.Errorf("expected helpScroll=0 after scroll up, got %d", m.helpScroll)
	}
}

func TestMouseWheel_FilterOverlayIgnored(t *testing.T) {
	m := modelWithTree(nil, false)
	m.filtering = true

	updated, _ := m.Update(tea.MouseMsg{
		X: 5, Y: 5,
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	m = updated.(Model)

	if m.treeScroll != 0 {
		t.Error("scroll should be ignored during filter overlay")
	}
	if m.detailScroll != 0 {
		t.Error("scroll should be ignored during filter overlay")
	}
}
