package ui

import (
	"regexp"
	"strings"
)

// ANSI escape sequence pattern for SGR codes
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// noColorHighlight controls whether search highlighting is disabled.
// Set to true in tests with Ascii color profile.
var noColorHighlight bool

// highlightSearchMatches highlights all case-insensitive occurrences of query in text.
// Works with both plain and ANSI-styled text. Uses background color only (reset via
// \x1b[49m) so it composes safely with existing foreground/faint/bold styles.
// Returns text unchanged if query is empty or color profile is Ascii.
func highlightSearchMatches(text, query string) string {
	if query == "" || text == "" {
		return text
	}
	if noColorHighlight {
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

// highlightSelectionRange highlights visible characters from fromCol to toCol
// (exclusive) in a line that may contain ANSI escape sequences. Uses reverse
// video so it composes with existing styles. If toCol is -1, highlights to end
// of line. Columns are 0-based visible character indices (rune count, not bytes).
func highlightSelectionRange(line string, fromCol, toCol int) string {
	if fromCol < 0 {
		fromCol = 0
	}

	const selOpen = "\x1b[7m"   // reverse video
	const selClose = "\x1b[27m" // reverse video off

	runes := []rune(line)
	var result strings.Builder
	visIdx := 0 // visible character index (runes, skipping ANSI)
	inSel := false
	i := 0
	for i < len(runes) {
		// Check for ANSI escape sequence at current position
		// Convert back to string from current rune position to check
		remaining := string(runes[i:])
		if loc := ansiRe.FindStringIndex(remaining); loc != nil && loc[0] == 0 {
			// Write the ANSI sequence as-is
			result.WriteString(remaining[:loc[1]])
			// Advance i by the number of runes in the ANSI sequence
			ansiRunes := []rune(remaining[:loc[1]])
			i += len(ansiRunes)
			continue
		}
		shouldHL := visIdx >= fromCol && (toCol < 0 || visIdx < toCol)
		if shouldHL && !inSel {
			result.WriteString(selOpen)
			inSel = true
		} else if !shouldHL && inSel {
			result.WriteString(selClose)
			inSel = false
		}
		result.WriteRune(runes[i])
		visIdx++
		i++
	}
	if inSel {
		result.WriteString(selClose)
	}
	return result.String()
}
