package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ANSI escape sequence pattern for SGR codes
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// highlightSearchMatches highlights all case-insensitive occurrences of query in text.
// Works with both plain and ANSI-styled text. Uses background color only (reset via
// \x1b[49m) so it composes safely with existing foreground/faint/bold styles.
// Returns text unchanged if query is empty or color profile is Ascii.
func highlightSearchMatches(text, query string) string {
	if query == "" || text == "" {
		return text
	}
	if lipgloss.ColorProfile() == termenv.Ascii {
		return text
	}

	lowerQuery := strings.ToLower(query)

	// Build mapping: for each visible byte, record its position in the original text.
	var visibleBytes []byte
	var origPos []int
	i := 0
	for i < len(text) {
		if loc := ansiRe.FindStringIndex(text[i:]); loc != nil && loc[0] == 0 {
			i += loc[1]
			continue
		}
		visibleBytes = append(visibleBytes, text[i])
		origPos = append(origPos, i)
		i++
	}

	lowerVisible := strings.ToLower(string(visibleBytes))

	// Find all match ranges and collect original positions to highlight.
	matchSet := make(map[int]bool)
	pos := 0
	for pos < len(lowerVisible) {
		idx := strings.Index(lowerVisible[pos:], lowerQuery)
		if idx == -1 {
			break
		}
		for vi := pos + idx; vi < pos+idx+len(lowerQuery) && vi < len(origPos); vi++ {
			matchSet[origPos[vi]] = true
		}
		pos = pos + idx + len(lowerQuery)
	}

	if len(matchSet) == 0 {
		return text
	}

	// Walk through original text, inserting background highlight codes.
	// Uses \x1b[49m (reset background only) to preserve faint/bold/foreground.
	const hlOpen = "\x1b[48;2;224;175;104m"
	const hlClose = "\x1b[49m"

	var result strings.Builder
	inHL := false
	i = 0
	for i < len(text) {
		if loc := ansiRe.FindStringIndex(text[i:]); loc != nil && loc[0] == 0 {
			result.WriteString(text[i : i+loc[1]])
			i += loc[1]
			continue
		}
		shouldHL := matchSet[i]
		if shouldHL && !inHL {
			result.WriteString(hlOpen)
			inHL = true
		} else if !shouldHL && inHL {
			result.WriteString(hlClose)
			inHL = false
		}
		result.WriteByte(text[i])
		i++
	}
	if inHL {
		result.WriteString(hlClose)
	}

	return result.String()
}
