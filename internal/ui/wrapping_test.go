package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/timoch/bd-view/internal/data"
)

// ---------------------------------------------------------------------------
// shieldWrapBreakpoints / unshieldWrapBreakpoints
// ---------------------------------------------------------------------------

func TestShieldWrapBreakpoints(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Hyphens
		{"mid-word hyphen", "bd-view", "bd\uE000view"},
		{"multiple hyphenated", "copy-paste and bd-view", "copy\uE000paste and bd\uE000view"},
		{"leading hyphen", "-flag", "-flag"},
		{"trailing hyphen after word", "word-", "word\uE000"},
		{"double dash", "--verbose", "--verbose"},
		{"triple chain", "a-b-c", "a\uE000b\uE000c"},
		{"hyphen between digits", "v1-2", "v1\uE0002"},

		// Commas
		{"trailing comma", "something,", "something\uE001"},
		{"mid-word comma", "a,b", "a\uE001b"},
		{"leading comma", ",start", ",start"},

		// Dots
		{"trailing dot", "word.", "word\uE002"},
		{"mid-word dot", "file.txt", "file\uE002txt"},
		{"leading dot", ".hidden", ".hidden"},

		// Semicolons
		{"trailing semicolon", "stmt;", "stmt\uE003"},
		{"mid-word semicolon", "a;b", "a\uE003b"},

		// Plus
		{"mid-word plus", "c++", "c\uE004+"},
		{"trailing plus", "a+", "a\uE004"},

		// Pipe
		{"mid-word pipe", "a|b", "a\uE005b"},

		// ANSI-styled text
		{"ansi before comma", "\x1b[1mword\x1b[0m,", "\x1b[1mword\x1b[0m\uE001"},
		{"ansi before hyphen", "\x1b[1mbd\x1b[0m-view", "\x1b[1mbd\x1b[0m\uE000view"},
		{"ansi before dot", "\x1b[1mfile\x1b[0m.txt", "\x1b[1mfile\x1b[0m\uE002txt"},

		// Mixed
		{"no breakpoints", "hello world", "hello world"},
		{"empty", "", ""},
		{"single char", "-", "-"},
		{"space-hyphen-word", "foo -bar", "foo -bar"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shieldWrapBreakpoints(tc.input)
			if got != tc.want {
				t.Errorf("shieldWrapBreakpoints(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestUnshieldWrapBreakpoints(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"hyphen", "bd\uE000view", "bd-view"},
		{"comma", "something\uE001", "something,"},
		{"dot", "file\uE002txt", "file.txt"},
		{"semicolon", "a\uE003b", "a;b"},
		{"plus", "c\uE004\uE004", "c++"},
		{"pipe", "a\uE005b", "a|b"},
		{"no substitutes", "hello world", "hello world"},
		{"empty", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := unshieldWrapBreakpoints(tc.input)
			if got != tc.want {
				t.Errorf("unshieldWrapBreakpoints(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestShieldUnshieldRoundTrip(t *testing.T) {
	inputs := []string{
		"bd-view", "copy-paste", "a-b-c", "-flag", "word-", "--verbose",
		"hello world", "", "x-y-z-w", "multi-word hyph-enated text-here",
		"something, else", "file.txt", "stmt;next", "c++", "a|b",
		"complex-expr, with.dots; and-hyphens",
	}
	for _, s := range inputs {
		got := unshieldWrapBreakpoints(shieldWrapBreakpoints(s))
		if got != s {
			t.Errorf("round-trip failed for %q: got %q", s, got)
		}
	}
}

// ---------------------------------------------------------------------------
// No residual shielded characters in output
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_NoResidualShieldedChars(t *testing.T) {
	m := New(Config{NoColor: true})
	m.selectedBead = &data.Bead{
		ID:          "proj-1",
		Title:       "Set up bd-view CI/CD pipe-line",
		Description: "Install bd-view via copy-paste. Use apt-get or brew-cask for auto-install, then run config.setup; done.",
		Design:      "The multi-stage pipe-line runs end-to-end. Output goes to stdout|stderr.",
	}

	for _, width := range []int{40, 60, 80, 120} {
		lines := m.buildDetailContentLines(width)
		for i, line := range lines {
			for _, sub := range []string{"\uE000", "\uE001", "\uE002", "\uE003", "\uE004", "\uE005"} {
				if strings.Contains(line, sub) {
					t.Errorf("width=%d line %d contains shielded char %U: %q", width, i, []rune(sub)[0], stripAnsi(line))
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// All output lines fit within contentWidth
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_AllLinesFitContentWidth(t *testing.T) {
	m := New(Config{NoColor: true})
	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Title:       strings.Repeat("TitleWord ", 30),
		Description: "This is a paragraph with several words that should be wrapped properly across multiple lines without any single line exceeding the content width limit.",
		Design:      "Use `some-long-inline-code-identifier` in the configuration file for the system-under-test.",
		AcceptanceCriteria: "- [ ] First long criterion that has enough words to require wrapping\n- [ ] Second criterion\n- [ ] Third criterion with extra detail padding",
		Notes:       strings.Repeat("note ", 60),
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
	w := ansi.StringWidth(titleLine)
	if w != contentWidth {
		t.Errorf("expected title line width %d, got %d", contentWidth, w)
	}

	// Overflow with spaces so ansi.Wrap can word-break.
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
	panelWidth := 50
	contentWidth := detailContentWidth(panelWidth)
	padding := strings.Repeat("w ", contentWidth/2-3)
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
// Comma stays attached to preceding word
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_CommaStaysAttached(t *testing.T) {
	m := New(Config{NoColor: true})
	panelWidth := 50
	contentWidth := detailContentWidth(panelWidth)
	// Push "something," near the wrap boundary.
	padding := strings.Repeat("w ", contentWidth/2-4)
	desc := padding + "something, next word here."

	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Description: desc,
	}

	lines := m.buildDetailContentLines(panelWidth)
	// "something," (with comma) must appear intact on a single line.
	found := false
	for _, line := range lines {
		if strings.Contains(stripAnsi(line), "something,") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("\"something,\" was split from its comma:\n%s", strings.Join(lines, "\n"))
	}
}

// ---------------------------------------------------------------------------
// Styled text + comma — ANSI codes between word and punctuation
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_StyledTextCommaStaysAttached(t *testing.T) {
	// Use color mode so glamour emits ANSI codes around bold text.
	// This tests that shieldWrapBreakpoints handles \x1b[0m before a comma.
	m := New(Config{NoColor: false})
	panelWidth := 50
	contentWidth := detailContentWidth(panelWidth)
	padding := strings.Repeat("w ", contentWidth/2-4)
	// Markdown bold: glamour emits \x1b[1msomething\x1b[0m, so the comma
	// follows an ANSI reset, not a word character.
	desc := padding + "**something**, next word here."

	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Description: desc,
	}

	lines := m.buildDetailContentLines(panelWidth)
	found := false
	for _, line := range lines {
		s := stripAnsi(line)
		if strings.Contains(s, "something,") {
			found = true
			break
		}
	}
	if !found {
		joined := ""
		for _, line := range lines {
			joined += stripAnsi(line) + "\n"
		}
		t.Errorf("styled \"something,\" was split from its comma:\n%s", joined)
	}
}

// ---------------------------------------------------------------------------
// Dot stays attached (e.g. file.txt)
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_DotStaysAttached(t *testing.T) {
	m := New(Config{NoColor: true})
	panelWidth := 50
	contentWidth := detailContentWidth(panelWidth)
	padding := strings.Repeat("w ", contentWidth/2-4)
	desc := padding + "config.yaml is important."

	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Description: desc,
	}

	lines := m.buildDetailContentLines(panelWidth)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "config.yaml") {
		t.Errorf("\"config.yaml\" was split at dot:\n%s", joined)
	}
}

// ---------------------------------------------------------------------------
// Markdown rendering preserves compound words with all breakpoint chars
// ---------------------------------------------------------------------------

func TestRenderMarkdown_PreservesCompoundWords(t *testing.T) {
	m := New(Config{NoColor: true})
	text := "Install bd-view using apt-get. Then copy-paste the config.yaml file, done."
	result := m.renderMarkdown(text, 40)

	for _, word := range []string{"bd-view", "apt-get", "copy-paste", "config.yaml", "file,"} {
		if !strings.Contains(result, word) {
			t.Errorf("compound word %q was split in markdown output:\n%s", word, result)
		}
	}
	// No residual shielded characters
	for _, sub := range []string{"\uE000", "\uE001", "\uE002", "\uE003", "\uE004", "\uE005"} {
		if strings.Contains(result, sub) {
			t.Errorf("shielded char %U leaked into markdown output:\n%s", []rune(sub)[0], result)
		}
	}
}

// ---------------------------------------------------------------------------
// Selection/View width consistency — overlay mode
// ---------------------------------------------------------------------------

func TestBuildDetailContentLines_SelectionMatchesRendered_Overlay(t *testing.T) {
	m := New(Config{Refresh: 2, NoColor: true})
	m.width = 90
	m.height = 30
	m.ready = true
	m.showOverlay = true

	m.selectedBead = &data.Bead{
		ID:          "test-1",
		Title:       strings.Repeat("T", 200),
		Description: strings.Repeat("word ", 50),
	}

	m.refreshDetailLines()
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
