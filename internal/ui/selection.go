package ui

import "strings"

// screenToDetailCoord converts screen (x, y) coordinates to detail content
// row and column, accounting for panel offset, padding, and scroll.
func (m Model) screenToDetailCoord(x, y int) (row, col int) {
	// Calculate detail panel x offset
	panelX := 0
	if !m.showOverlay && !m.isNarrow() {
		panelX = m.treeWidth() + 1 // +1 for border
	}
	// Detail panel has PaddingLeft(1)
	col = x - panelX - 1
	if col < 0 {
		col = 0
	}
	// Row is relative to top of panel, plus scroll offset
	row = y + m.detailScroll
	if row < 0 {
		row = 0
	}
	return row, col
}

// extractSelectedText returns the text within the current selection range
// from the stored detail content lines.
func (m Model) extractSelectedText() string {
	if len(m.detailLines) == 0 {
		return ""
	}

	// Normalize selection direction (start <= end)
	startRow, startCol, endRow, endCol := m.selStartRow, m.selStartCol, m.selEndRow, m.selEndCol
	if startRow > endRow || (startRow == endRow && startCol > endCol) {
		startRow, startCol, endRow, endCol = endRow, endCol, startRow, startCol
	}

	var result strings.Builder
	for r := startRow; r <= endRow && r < len(m.detailLines); r++ {
		if r < 0 {
			continue
		}
		line := stripAnsi(m.detailLines[r])
		lineRunes := []rune(line)

		fromCol := 0
		toCol := len(lineRunes)
		if r == startRow {
			fromCol = startCol
		}
		if r == endRow {
			toCol = endCol + 1
		}
		if fromCol < 0 {
			fromCol = 0
		}
		if fromCol > len(lineRunes) {
			fromCol = len(lineRunes)
		}
		if toCol > len(lineRunes) {
			toCol = len(lineRunes)
		}
		if toCol < fromCol {
			toCol = fromCol
		}

		if r > startRow {
			result.WriteByte('\n')
		}
		result.WriteString(string(lineRunes[fromCol:toCol]))
	}
	return result.String()
}

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// selectionNormalized returns the selection range with start <= end.
func (m Model) selectionNormalized() (startRow, startCol, endRow, endCol int) {
	startRow, startCol = m.selStartRow, m.selStartCol
	endRow, endCol = m.selEndRow, m.selEndCol
	if startRow > endRow || (startRow == endRow && startCol > endCol) {
		startRow, startCol, endRow, endCol = endRow, endCol, startRow, startCol
	}
	return
}
