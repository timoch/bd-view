package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/timoch/bd-view/internal/data"
)

// ---------------------------------------------------------------------------
// shieldHyphens / unshieldHyphens
// ---------------------------------------------------------------------------

func TestShieldHyphens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"mid-word", "bd-view", "bd\u2011view"},
		{"multiple words", "copy-paste and bd-view", "copy\u2011paste and bd\u2011view"},
		{"leading hyphen", "-flag", "-flag"},
		{"trailing hyphen", "word-", "word-"},
		{"double dash", "--verbose", "--verbose"},
		{"triple chain", "a-b-c", "a\u2011b\u2011c"},
		{"no hyphens", "hello world", "hello world"},
		{"empty", "", ""},
		{"single char", "-", "-"},
		{"hyphen between digits", "v1-2", "v1\u20112"},
		{"space-hyphen-word", "foo -bar", "foo -bar"},
		{"word-hyphen-space", "foo- bar", "foo- bar"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shieldHyphens(tc.input)
			if got != tc.want {
				t.Errorf("shieldHyphens(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestUnshieldHyphens(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"bd\u2011view", "bd-view"},
		{"no special", "no special"},
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := unshieldHyphens(tc.input)
			if got != tc.want {
				t.Errorf("unshieldHyphens(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestShieldUnshieldRoundTrip(t *testing.T) {
	inputs := []string{
		"bd-view", "copy-paste", "a-b-c", "-flag", "word-", "--verbose",
		"hello world", "", "x-y-z-w", "multi-word hyph-enated text-here",
	}
	for _, s := range inputs {
		got := unshieldHyphens(shieldHyphens(s))
		if got != s {
			t.Errorf("round-trip failed for %q: got %q", s, got)
		}
	}
}

// ---------------------------------------------------------------------------
// No residual non-breaking hyphens in output
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_NoResidualNonBreakingHyphen(t *testing.T) {
	m := New(Config{NoColor: true})
	m.selectedBead = &data.Bead{
		ID:          "proj-1",
		Title:       "Set up bd-view CI/CD pipe-line",
		Description: "Install bd-view via copy-paste. Use apt-get or brew-cask for auto-install.",
		Design:      "The multi-stage pipe-line runs end-to-end.",
	}

	for _, width := range []int{40, 60, 80, 120} {
		lines := m.buildDetailContentLines(width)
		for i, line := range lines {
			if strings.Contains(line, nonBreakingHyphen) {
				t.Errorf("width=%d line %d contains U+2011: %q", width, i, line)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// All output lines fit within contentWidth
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_AllLinesFitContentWidth(t *testing.T) {
	m := New(Config{NoColor: true})
	// Use a bead with content in every section to exercise all code paths.
	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Title:       strings.Repeat("TitleWord ", 30), // ~300 chars
		Description: "This is a paragraph with several words that should be wrapped properly across multiple lines without any single line exceeding the content width limit.",
		Design:      "Use `some-long-inline-code-identifier` in the configuration file for the system-under-test.",
		AcceptanceCriteria: "- [ ] First long criterion that has enough words to require wrapping\n- [ ] Second criterion\n- [ ] Third criterion with extra detail padding",
		Notes:       strings.Repeat("note ", 60), // ~300 chars
	}

	for _, panelWidth := range []int{30, 50, 80, 120} {
		contentWidth := detailContentWidth(panelWidth)
		lines := m.buildDetailContentLines(panelWidth)
		for i, line := range lines {
			w := ansi.StringWidth(line)
			if w > contentWidth {
				stripped := stripAnsi(line)
				t.Errorf("panelWidth=%d line %d: visible width %d > contentWidth %d: %q",
					panelWidth, i, w, contentWidth, stripped)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Exact-fit boundary: contentWidth chars should not wrap
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_ExactFitDoesNotWrap(t *testing.T) {
	panelWidth := 80
	contentWidth := detailContentWidth(panelWidth)

	m := New(Config{NoColor: true})
	// "Title:  " is 8 chars, so a title of contentWidth-8 fills exactly.
	prefix := "Title:  "
	titleLen := contentWidth - len(prefix)
	if titleLen < 1 {
		t.Skip("contentWidth too small for this test")
	}

	m.selectedBead = &data.Bead{
		ID:    "test-1",
		Title: strings.Repeat("X", titleLen),
	}

	lines := m.buildDetailContentLines(panelWidth)
	// Find the Title line
	titleLine := ""
	for _, line := range lines {
		if strings.HasPrefix(stripAnsi(line), "Title:") {
			titleLine = line
			break
		}
	}
	if titleLine == "" {
		t.Fatal("Title line not found in output")
	}
	// Exact fit: the title should be on one line, not wrapped
	w := ansi.StringWidth(titleLine)
	if w != contentWidth {
		t.Errorf("expected title line width %d, got %d", contentWidth, w)
	}

	// Now overflow: use spaces so ansi.Wrap can word-break.
	// "Title:  XXXX ... overflow word" should wrap the overflow onto a new line.
	m.selectedBead.Title = strings.Repeat("X", titleLen-6) + " overflow"
	lines = m.buildDetailContentLines(panelWidth)
	titleLines := 0
	for _, line := range lines {
		s := stripAnsi(line)
		if strings.Contains(s, "Title:") || strings.Contains(s, "overflow") || strings.Contains(s, "XXXX") {
			titleLines++
		}
	}
	if titleLines < 2 {
		t.Errorf("expected overflow to wrap across 2+ lines, got %d", titleLines)
	}
}

// ---------------------------------------------------------------------------
// Very narrow panel — no panics, lines fit
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_VeryNarrowPanel(t *testing.T) {
	m := New(Config{NoColor: true})
	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Title:       "A reasonably long title",
		Description: "Some description text that will need wrapping.",
	}

	panelWidth := 15
	contentWidth := detailContentWidth(panelWidth)
	lines := m.buildDetailContentLines(panelWidth)

	if len(lines) == 0 {
		t.Fatal("expected some output lines")
	}
	for i, line := range lines {
		w := ansi.StringWidth(line)
		if w > contentWidth {
			t.Errorf("line %d: visible width %d > contentWidth %d: %q",
				i, w, contentWidth, stripAnsi(line))
		}
	}
}

// ---------------------------------------------------------------------------
// Hyphenated words at wrap boundary stay together
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_HyphenAtWrapBoundary(t *testing.T) {
	m := New(Config{NoColor: true})
	// Construct a description where "copy-paste" is forced near the line end.
	// Use padding words to push it to the boundary.
	panelWidth := 50
	contentWidth := detailContentWidth(panelWidth)
	// "copy-paste" is 10 chars. Build a line that fills up to ~contentWidth-5
	// so "copy-paste" can't fit on the first line and must wrap to the next.
	padding := strings.Repeat("w ", contentWidth/2-3) // fills most of the line
	desc := padding + "copy-paste done."

	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Description: desc,
	}

	lines := m.buildDetailContentLines(panelWidth)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "copy-paste") {
		t.Errorf("copy-paste was split across lines:\n%s", joined)
	}
}

// ---------------------------------------------------------------------------
// Markdown rendering through glamour preserves hyphenated words
// ---------------------------------------------------------------------------

func TestRenderMarkdown_PreservesHyphenatedWords(t *testing.T) {
	m := New(Config{NoColor: true})
	text := "Install bd-view using apt-get. Then copy-paste the config."
	result := m.renderMarkdown(text, 40)

	for _, word := range []string{"bd-view", "apt-get", "copy-paste"} {
		if !strings.Contains(result, word) {
			t.Errorf("hyphenated word %q was split in markdown output:\n%s", word, result)
		}
	}
	// No residual non-breaking hyphens
	if strings.Contains(result, nonBreakingHyphen) {
		t.Errorf("non-breaking hyphen leaked into markdown output:\n%s", result)
	}
}

// ---------------------------------------------------------------------------
// Selection/View width consistency — overlay mode
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_SelectionMatchesRendered_Overlay(t *testing.T) {
	m := New(Config{Refresh: 2, NoColor: true})
	m.width = 90 // narrow mode
	m.height = 30
	m.ready = true
	m.showOverlay = true

	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Title:       strings.Repeat("T", 200),
		Description: strings.Repeat("word ", 50),
	}

	// refreshDetailLines (Update path)
	m.refreshDetailLines()

	// View path: overlay uses m.width
	viewLines := m.buildDetailContentLines(m.width)

	if len(m.detailLines) != len(viewLines) {
		t.Errorf("overlay mode: detailLines (%d) != viewLines (%d)",
			len(m.detailLines), len(viewLines))
	}
	for i := 0; i < len(m.detailLines) && i < len(viewLines); i++ {
		if m.detailLines[i] != viewLines[i] {
			t.Errorf("overlay mode: line %d differs:\n  detail: %q\n  view:   %q",
				i, stripAnsi(m.detailLines[i]), stripAnsi(viewLines[i]))
			break
		}
	}
}

// ---------------------------------------------------------------------------
// Search highlighting + wrapping — ANSI codes don't inflate visible width
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_SearchHighlightWrapping(t *testing.T) {
	m := New(Config{NoColor: true})
	m.searchQuery = "test"

	m.selectedBead = &data.Bead{
		ID:          "test-long-id",
		Title:       "A test title with test words repeated to test wrapping near the test boundary " + strings.Repeat("x", 100),
		Description: "This tests that search highlighting ANSI codes do not cause width miscalculation.",
	}

	panelWidth := 80
	contentWidth := detailContentWidth(panelWidth)
	lines := m.buildDetailContentLines(panelWidth)

	for i, line := range lines {
		w := ansi.StringWidth(line)
		if w > contentWidth {
			t.Errorf("line %d with search highlight: visible width %d > contentWidth %d: %q",
				i, w, contentWidth, stripAnsi(line))
		}
	}
}
