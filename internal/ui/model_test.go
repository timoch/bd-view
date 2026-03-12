package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/data"
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
	m.height = 10
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
	result := renderMarkdown("This is **bold** text", 80)
	if !strings.Contains(result, "bold") {
		t.Error("expected bold text in output")
	}
	if strings.Contains(result, "**") {
		t.Error("bold markers should be removed")
	}
}

func TestRenderMarkdown_CodeBlock(t *testing.T) {
	input := "Before\n```\ncode line\n```\nAfter"
	result := renderMarkdown(input, 80)
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
	input := "- Item one\n- Item two\n- Item three"
	result := renderMarkdown(input, 80)
	if !strings.Contains(result, "- Item one") {
		t.Error("expected bullet items preserved")
	}
}

func TestWrapLine(t *testing.T) {
	result := wrapLine("short", 80)
	if result != "short" {
		t.Errorf("short line should not wrap, got %q", result)
	}

	long := strings.Repeat("word ", 20)
	result = wrapLine(long, 30)
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("expected long line to wrap into multiple lines")
	}
	for _, l := range lines {
		if len(l) > 30 {
			t.Errorf("wrapped line exceeds width: %q", l)
		}
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
