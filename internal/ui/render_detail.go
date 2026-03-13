package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/data"
)

// buildDetailContentLines builds the full list of rendered detail content lines
// for the selected bead. Used by both renderDetailPanel (for display) and
// refreshDetailLines (for text selection coordinate mapping in Update).
func (m Model) buildDetailContentLines(width int) []string {
	if m.selectedBead == nil {
		return nil
	}

	b := m.selectedBead
	var lines []string
	contentWidth := width - 2 // account for padding

	// Title line: bead ID as pane title
	titleStyle := lipgloss.NewStyle().Bold(true)
	if m.focusedPane == detailPane || m.showOverlay {
		titleStyle = titleStyle.Foreground(colorAccentPrimary)
	}
	lines = append(lines, titleStyle.Render(b.ID))
	lines = append(lines, strings.Repeat("─", contentWidth))

	// Title field displayed prominently
	if b.Title != "" {
		titleText := b.Title
		if m.searchQuery != "" {
			titleText = highlightSearchMatches(titleText, m.searchQuery)
		}
		lines = append(lines, fmt.Sprintf("Title:  %s", titleText))
	}

	// Metadata row: Type, Status (with color), Priority, Owner
	statusStr := m.colorStatus(b.Status)
	lines = append(lines, fmt.Sprintf("Type:   %-12s Status: %s", b.IssueType, statusStr))
	lines = append(lines, fmt.Sprintf("Priority: %-10d Owner: %s", b.Priority, b.Owner))

	// Parent bead ID (if present)
	if b.Parent != "" {
		lines = append(lines, fmt.Sprintf("Parent: %s", b.Parent))
	}

	// Date fields (only non-empty)
	var dateParts []string
	if b.CreatedAt != nil {
		dateParts = append(dateParts, fmt.Sprintf("Created: %s", b.CreatedAt.Format("2006-01-02")))
	}
	if b.UpdatedAt != nil {
		dateParts = append(dateParts, fmt.Sprintf("Updated: %s", b.UpdatedAt.Format("2006-01-02")))
	}
	if b.ClosedAt != nil {
		dateParts = append(dateParts, fmt.Sprintf("Closed: %s", b.ClosedAt.Format("2006-01-02")))
	}
	if len(dateParts) > 0 {
		lines = append(lines, strings.Join(dateParts, "  "))
	}

	// Horizontal separator between header and body sections
	lines = append(lines, strings.Repeat("─", contentWidth))

	// Body sections: only show non-empty sections
	sections := []struct {
		heading string
		content string
	}{
		{"DESCRIPTION", b.Description},
		{"DESIGN", b.Design},
		{"ACCEPTANCE CRITERIA", b.AcceptanceCriteria},
		{"NOTES", b.Notes},
	}
	for _, sec := range sections {
		if strings.TrimSpace(sec.content) == "" {
			continue
		}
		headingStyle := lipgloss.NewStyle().Bold(true).Underline(true)
		lines = append(lines, "")
		lines = append(lines, headingStyle.Render(sec.heading))
		rendered := m.renderMarkdown(sec.content, contentWidth)
		if m.searchQuery != "" {
			rendered = highlightSearchMatches(rendered, m.searchQuery)
		}
		lines = append(lines, rendered)
	}

	// Dependencies section
	depLines := m.renderDependencies(b)
	if len(depLines) > 0 {
		headingStyle := lipgloss.NewStyle().Bold(true).Underline(true)
		lines = append(lines, "")
		lines = append(lines, headingStyle.Render("DEPENDENCIES"))
		lines = append(lines, depLines...)
	}

	return strings.Split(strings.Join(lines, "\n"), "\n")
}

// refreshDetailLines rebuilds detailLines for text selection coordinate mapping.
// Must be called from Update() since View() uses a value receiver and mutations are lost.
func (m *Model) refreshDetailLines() {
	detailWidth := m.width
	if !m.showOverlay && !m.isNarrow() {
		detailWidth = m.width - m.treeWidth() - 1
	}
	m.detailLines = m.buildDetailContentLines(detailWidth)
}

func (m Model) renderDetailPanel(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingLeft(1)

	if m.selectedBead == nil {
		content := lipgloss.NewStyle().Faint(true).Render("Select a bead to view details")
		return style.Render(content)
	}

	allLines := m.buildDetailContentLines(width)

	// Apply scroll offset
	scrollOffset := m.detailScroll
	if scrollOffset > len(allLines)-1 {
		scrollOffset = len(allLines) - 1
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	visibleLines := allLines[scrollOffset:]
	if len(visibleLines) > height {
		visibleLines = visibleLines[:height]
	}

	// Apply selection highlighting before lipgloss render.
	// highlightSelectionRange only inserts \x1b[7m/\x1b[27m (reverse video on/off)
	// which are zero-width SGR codes and don't affect lipgloss width calculations.
	if m.selecting || m.hasSelection {
		startRow, startCol, endRow, endCol := m.selectionNormalized()
		for i, line := range visibleLines {
			contentRow := i + scrollOffset
			if contentRow < startRow || contentRow > endRow {
				continue
			}
			fromCol := 0
			toCol := -1 // -1 means end of line
			if contentRow == startRow {
				fromCol = startCol
			}
			if contentRow == endRow {
				toCol = endCol + 1
			}
			visibleLines[i] = highlightSelectionRange(line, fromCol, toCol)
		}
	}

	return style.Render(strings.Join(visibleLines, "\n"))
}

// renderDependencies returns lines showing dependency relationships.
func (m Model) renderDependencies(b *data.Bead) []string {
	var lines []string

	if len(b.Dependencies) > 0 {
		var ids []string
		for _, dep := range b.Dependencies {
			ids = append(ids, dep.DependsOnID)
		}
		lines = append(lines, fmt.Sprintf("  depends on: %s", strings.Join(ids, ", ")))
	}

	if len(m.dependents) > 0 {
		var ids []string
		for _, dep := range m.dependents {
			ids = append(ids, dep.ID)
		}
		lines = append(lines, fmt.Sprintf("  depended on by: %s", strings.Join(ids, ", ")))
	}

	return lines
}

// renderMarkdown renders markdown content using glamour for the terminal.
func (m Model) renderMarkdown(text string, width int) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	if width < 10 {
		width = 10
	}

	styleName := "dark"
	if m.config.NoColor {
		styleName = "notty"
	}

	profile := termenv.ColorProfile()
	if m.config.NoColor {
		profile = termenv.Ascii
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styleName),
		glamour.WithWordWrap(width),
		glamour.WithColorProfile(profile),
	)
	if err != nil {
		return text
	}

	rendered, err := r.Render(text)
	if err != nil {
		return text
	}

	// Trim trailing whitespace/newlines that glamour adds
	rendered = strings.TrimRight(rendered, "\n ")

	return rendered
}
