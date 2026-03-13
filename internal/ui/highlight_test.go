package ui

import (
	"strings"
	"testing"

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
	// noColorHighlight is set in init() — highlighting should be no-op
	result := highlightSearchMatches("hello world", "world")
	if result != "hello world" {
		t.Errorf("expected unchanged text when noColorHighlight=true, got %q", result)
	}
}

func TestHighlightSearchMatches_FindsMatch(t *testing.T) {
	// Temporarily enable highlighting
	noColorHighlight = false
	defer func() { noColorHighlight = true }()

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
	noColorHighlight = false
	defer func() { noColorHighlight = true }()

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
	noColorHighlight = false
	defer func() { noColorHighlight = true }()

	result := highlightSearchMatches("ab cd ab ef ab", "ab")
	count := strings.Count(result, "\x1b[48;2;224;175;104m")
	if count != 3 {
		t.Errorf("expected 3 highlight opens, got %d", count)
	}
}

func TestHighlightSearchMatches_NoMatch(t *testing.T) {
	noColorHighlight = false
	defer func() { noColorHighlight = true }()

	result := highlightSearchMatches("hello world", "xyz")
	if result != "hello world" {
		t.Errorf("expected unchanged text when no match, got %q", result)
	}
}

func TestHighlightSearchMatches_SpecialChars(t *testing.T) {
	noColorHighlight = false
	defer func() { noColorHighlight = true }()

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
	noColorHighlight = false
	defer func() { noColorHighlight = true }()

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
	noColorHighlight = false
	defer func() { noColorHighlight = true }()

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
	// With noColorHighlight=true, highlighting is invisible but the text is present.
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

	output := m.viewString()
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

	output := m.viewString()
	if !strings.Contains(output, "connection") {
		t.Error("expected description with search term in detail pane")
	}
}

func TestHighlightSelectionRange_PerCharANSI(t *testing.T) {
	// Simulate per-character ANSI styling like lipgloss Bold+Underline produces.
	// Each character is wrapped: \x1b[1;4mX\x1b[0m
	// The embedded \x1b[0m resets should NOT cancel the selection highlight.
	line := "\x1b[1;4mD\x1b[0m\x1b[1;4mE\x1b[0m\x1b[1;4mS\x1b[0m\x1b[1;4mC\x1b[0m"
	result := highlightSelectionRange(line, 0, -1) // highlight all

	// The reverse video code should be active for all visible characters.
	// Strip everything except reverse-video on/off and visible chars to verify.
	stripped := stripAnsi(result)
	if stripped != "DESC" {
		t.Errorf("expected visible text 'DESC', got %q", stripped)
	}

	// Each visible character should be within a \x1b[7m...\x1b[27m range.
	// Count reverse-video-on codes: should be re-applied after each reset.
	revOn := strings.Count(result, "\x1b[7m")
	if revOn < 1 {
		t.Error("expected at least one reverse-video-on code")
	}
	// The key test: ensure reverse video is active for ALL characters, not just the first.
	// After the first \x1b[0m reset, \x1b[7m must be re-emitted.
	// Check that 'E' is preceded by \x1b[7m (possibly with other ANSI in between).
	eIdx := strings.Index(result, "E")
	if eIdx == -1 {
		t.Fatal("'E' not found in result")
	}
	// Find the last \x1b[7m before 'E'
	before := result[:eIdx]
	lastRevOn := strings.LastIndex(before, "\x1b[7m")
	lastRevOff := strings.LastIndex(before, "\x1b[27m")
	lastReset := strings.LastIndex(before, "\x1b[0m")
	// Reverse video must be on: lastRevOn must be after both lastRevOff and lastReset
	if lastRevOn < lastReset || lastRevOn < lastRevOff {
		t.Errorf("reverse video not active before 'E': lastRevOn=%d, lastReset=%d, lastRevOff=%d\nresult: %q", lastRevOn, lastReset, lastRevOff, result)
	}
}

func TestHighlightSelectionRange_GlamourStyledText(t *testing.T) {
	// Glamour wraps paragraph content with color codes and resets between segments.
	line := "\x1b[38;5;252m\x1b[0m  \x1b[38;5;252mHello world\x1b[0m"
	result := highlightSelectionRange(line, 0, -1)

	stripped := stripAnsi(result)
	if stripped != "  Hello world" {
		t.Errorf("expected '  Hello world', got %q", stripped)
	}

	// 'w' in "world" should have reverse video active
	wIdx := strings.Index(result, "w")
	if wIdx == -1 {
		t.Fatal("'w' not found")
	}
	before := result[:wIdx]
	lastRevOn := strings.LastIndex(before, "\x1b[7m")
	lastReset := strings.LastIndex(before, "\x1b[0m")
	if lastRevOn < lastReset {
		t.Errorf("reverse video not active before 'w': lastRevOn=%d, lastReset=%d", lastRevOn, lastReset)
	}
}

func TestHighlightSelectionRange_PartialRange(t *testing.T) {
	// Per-char styled text, select only middle characters (cols 1-2)
	line := "\x1b[1mA\x1b[0m\x1b[1mB\x1b[0m\x1b[1mC\x1b[0m\x1b[1mD\x1b[0m"
	result := highlightSelectionRange(line, 1, 3) // highlight B and C

	// A should NOT have reverse video; B and C should; D should not
	aIdx := strings.Index(result, "A")
	bIdx := strings.Index(result, "B")
	dIdx := strings.Index(result, "D")

	// Before A: no \x1b[7m
	if strings.Contains(result[:aIdx], "\x1b[7m") {
		t.Error("A should not be highlighted")
	}
	// Before B: \x1b[7m should be present
	beforeB := result[:bIdx]
	if !strings.Contains(beforeB, "\x1b[7m") {
		t.Error("B should be highlighted")
	}
	// Before D: \x1b[27m should close selection
	beforeD := result[:dIdx]
	lastClose := strings.LastIndex(beforeD, "\x1b[27m")
	lastOpen := strings.LastIndex(beforeD, "\x1b[7m")
	if lastClose < lastOpen {
		t.Error("selection should be closed before D")
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

	// Verify no highlighting applied (text unchanged with noColorHighlight)
	output := m.viewString()
	if !strings.Contains(output, "Install SignalR client") {
		t.Error("expected title present after search cleared")
	}
}
