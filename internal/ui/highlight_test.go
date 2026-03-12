package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/tree"
)

func sampleBeadsForHighlight() []data.Bead {
	return []data.Bead{
		{
			ID:        "hep-ws-f6",
			Title:     "Install SignalR client",
			IssueType: "task",
			Status:    "open",
			Priority:  1,
		},
	}
}

func buildTreeForHighlight(beads []data.Bead) *tree.Model {
	return tree.BuildTree(beads, true)
}

func sampleBeadForHighlight() *data.Bead {
	return &data.Bead{
		ID:        "hep-ws-f6.3",
		Title:     "Install SignalR client",
		IssueType: "task",
		Status:    "open",
		Priority:  1,
	}
}

func TestHighlightSearchMatches_EmptyQuery(t *testing.T) {
	result := highlightSearchMatches("hello world", "")
	if result != "hello world" {
		t.Errorf("expected unchanged text for empty query, got %q", result)
	}
}

func TestHighlightSearchMatches_EmptyText(t *testing.T) {
	result := highlightSearchMatches("", "test")
	if result != "" {
		t.Errorf("expected empty string for empty text, got %q", result)
	}
}

func TestHighlightSearchMatches_AsciiProfile(t *testing.T) {
	// In init(), termenv.Ascii is set — highlighting should be no-op
	result := highlightSearchMatches("hello world", "world")
	if result != "hello world" {
		t.Errorf("expected unchanged text in Ascii profile, got %q", result)
	}
}

func TestHighlightSearchMatches_FindsMatch(t *testing.T) {
	// Temporarily set TrueColor profile
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	result := highlightSearchMatches("hello world", "world")
	if !strings.Contains(result, "world") {
		t.Error("expected result to contain 'world'")
	}
	if !strings.Contains(result, "\x1b[48;2;224;175;104m") {
		t.Error("expected result to contain highlight background ANSI code")
	}
	if !strings.Contains(result, "\x1b[49m") {
		t.Error("expected result to contain background reset code")
	}
}

func TestHighlightSearchMatches_CaseInsensitive(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	result := highlightSearchMatches("Hello World", "hello")
	// Should highlight "Hello" preserving original case
	if !strings.Contains(result, "\x1b[48;2;224;175;104m") {
		t.Error("expected case-insensitive match to be highlighted")
	}
	// The visible text should still contain "Hello" (original case)
	visible := ansiRe.ReplaceAllString(result, "")
	if !strings.Contains(visible, "Hello") {
		t.Errorf("expected original case preserved, visible text: %q", visible)
	}
}

func TestHighlightSearchMatches_MultipleMatches(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	result := highlightSearchMatches("ab cd ab ef ab", "ab")
	count := strings.Count(result, "\x1b[48;2;224;175;104m")
	if count != 3 {
		t.Errorf("expected 3 highlight opens, got %d", count)
	}
}

func TestHighlightSearchMatches_NoMatch(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	result := highlightSearchMatches("hello world", "xyz")
	if result != "hello world" {
		t.Errorf("expected unchanged text when no match, got %q", result)
	}
}

func TestHighlightSearchMatches_SpecialChars(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	// Special regex chars should be treated as literals
	result := highlightSearchMatches("test (foo) bar", "(foo)")
	if !strings.Contains(result, "\x1b[48;2;224;175;104m") {
		t.Error("expected literal match of special chars to be highlighted")
	}
	visible := ansiRe.ReplaceAllString(result, "")
	if visible != "test (foo) bar" {
		t.Errorf("expected text content unchanged, got %q", visible)
	}
}

func TestHighlightSearchMatches_WithANSICodes(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	// Simulate styled text: "he\x1b[1mllo" — "hello" with bold starting mid-word
	input := "he\x1b[1mllo world"
	result := highlightSearchMatches(input, "hello")

	// Should highlight across the ANSI boundary
	if !strings.Contains(result, "\x1b[48;2;224;175;104m") {
		t.Error("expected highlight across ANSI boundary")
	}
	// The bold code should still be present
	if !strings.Contains(result, "\x1b[1m") {
		t.Error("expected original ANSI codes preserved")
	}
	// Visible text should still be "hello world"
	visible := ansiRe.ReplaceAllString(result, "")
	if visible != "hello world" {
		t.Errorf("expected visible text 'hello world', got %q", visible)
	}
}

func TestHighlightSearchMatches_PreservesANSIStructure(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	// Faint text with a match
	input := "\x1b[2mhello world\x1b[0m"
	result := highlightSearchMatches(input, "world")

	// The faint codes should still be present
	if !strings.Contains(result, "\x1b[2m") {
		t.Error("expected faint code preserved")
	}
	if !strings.Contains(result, "\x1b[0m") {
		t.Error("expected reset code preserved")
	}
	// Background highlight should only reset background, not full reset
	if !strings.Contains(result, "\x1b[49m") {
		t.Error("expected background-only reset for highlight")
	}
}

func TestTreeRow_HighlightsIDOnSearch(t *testing.T) {
	// This test verifies that renderTreeRow includes the search query text in the ID.
	// In Ascii profile, highlighting is invisible but the text is present.
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.searchQuery = "f6"

	beads := sampleBeadsForHighlight()
	tr := buildTreeForHighlight(beads)
	m.SetTree(tr)

	visible := m.visibleNodes()
	if len(visible) == 0 {
		t.Fatal("expected visible nodes")
	}
	row := m.renderTreeRow(visible[0], visible, 80)
	if !strings.Contains(row, "hep-ws-f6") {
		t.Errorf("expected ID in tree row, got %q", row)
	}
}

func TestTreeRow_HighlightsTitleOnSearch(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.searchQuery = "install"

	beads := sampleBeadsForHighlight()
	tr := buildTreeForHighlight(beads)
	m.SetTree(tr)

	visible := m.visibleNodes()
	row := m.renderTreeRow(visible[0], visible, 120)
	if !strings.Contains(row, "Install") {
		t.Errorf("expected title in tree row, got %q", row)
	}
}

func TestDetailPanel_HighlightsTitleOnSearch(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.searchQuery = "SignalR"

	bead := sampleBeadForHighlight()
	m.SetSelectedBead(bead)

	output := m.View()
	if !strings.Contains(output, "Install SignalR client") {
		t.Error("expected title with search term in detail pane")
	}
}

func TestDetailPanel_HighlightsDescriptionOnSearch(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.searchQuery = "connection"

	bead := sampleBeadForHighlight()
	bead.Description = "Set up SignalR connection for real-time updates"
	m.SetSelectedBead(bead)

	output := m.View()
	if !strings.Contains(output, "connection") {
		t.Error("expected description with search term in detail pane")
	}
}

func TestHighlight_ClearedWhenSearchCleared(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true
	m.searchQuery = "test"

	bead := sampleBeadForHighlight()
	m.SetSelectedBead(bead)

	// Clear search
	m.searchQuery = ""

	// Verify no highlighting applied (text unchanged in Ascii mode)
	output := m.View()
	if !strings.Contains(output, "Install SignalR client") {
		t.Error("expected title present after search cleared")
	}
}
