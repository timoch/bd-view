package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelpOverlay renders the help overlay from the keybinding registry.
func (m Model) renderHelpOverlay(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingLeft(2)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colorAccentPrimary)
	headingStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	keyStyle := lipgloss.NewStyle().Bold(true)

	// Calculate max key width for alignment
	maxKeyLen := 0
	for _, kb := range keybindingRegistry {
		if len(kb.Keys) > maxKeyLen {
			maxKeyLen = len(kb.Keys)
		}
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Help — Keybindings"))
	lines = append(lines, "")

	// Group keybindings by section in defined order
	for _, section := range sectionOrder {
		lines = append(lines, headingStyle.Render(section))
		for _, kb := range keybindingRegistry {
			if kb.Section != section {
				continue
			}
			padding := strings.Repeat(" ", maxKeyLen-len(kb.Keys)+2)
			lines = append(lines, fmt.Sprintf("  %s%s%s", keyStyle.Render(kb.Keys), padding, kb.Description))
		}
		lines = append(lines, "")
	}

	lines = append(lines, lipgloss.NewStyle().Faint(true).Render("[Esc/?] Close  [j/k] Scroll"))

	// Apply scroll offset
	scrollOffset := m.helpScroll
	if scrollOffset > len(lines)-1 {
		scrollOffset = len(lines) - 1
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	visible := lines[scrollOffset:]
	if len(visible) > height {
		visible = visible[:height]
	}

	return style.Render(strings.Join(visible, "\n"))
}

// renderFilterOverlay renders the filter menu overlay.
func (m Model) renderFilterOverlay(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingLeft(2)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colorAccentPrimary)
	headingStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	selectedStyle := lipgloss.NewStyle().Reverse(true)

	var lines []string
	lines = append(lines, titleStyle.Render("Filter Beads"))
	lines = append(lines, "")
	lines = append(lines, headingStyle.Render("TYPE"))

	items := filterMenuItems()
	for i, item := range items {
		if item.section == 1 && (i == 0 || items[i-1].section == 0) {
			lines = append(lines, "")
			lines = append(lines, headingStyle.Render("STATUS"))
		}
		checked := " "
		if item.section == 0 && m.filterTypes[item.label] {
			checked = "x"
		} else if item.section == 1 && m.filterStats[item.label] {
			checked = "x"
		}
		row := fmt.Sprintf("  [%s] %s", checked, item.label)
		if i == m.filterCursor {
			row = selectedStyle.Render(row)
		}
		lines = append(lines, row)
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Faint(true).Render("[Space] Toggle  [Enter/f] Apply  [Esc] Clear & Close"))

	return style.Render(strings.Join(lines, "\n"))
}
