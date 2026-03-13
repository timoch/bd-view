package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/timoch/bd-view/internal/tree"
)

func (m Model) View() tea.View {
	var content string

	if !m.ready {
		content = "Loading..."
	} else if m.isTooSmall() {
		content = fmt.Sprintf("Terminal too small (%dx%d). Minimum size: %dx%d.", m.width, m.height, minTermWidth, minTermHeight)
	} else {
		statusBar := m.renderStatusBar()
		contentHeight := m.height - lipgloss.Height(statusBar)
		if contentHeight < 1 {
			contentHeight = 1
		}

		if m.showHelp {
			helpPanel := m.renderHelpOverlay(m.width, contentHeight)
			content = lipgloss.JoinVertical(lipgloss.Left, helpPanel, statusBar)
		} else if m.filtering {
			filterPanel := m.renderFilterOverlay(m.width, contentHeight)
			content = lipgloss.JoinVertical(lipgloss.Left, filterPanel, statusBar)
		} else if m.showOverlay {
			detailPanel := m.renderDetailPanel(m.width, contentHeight)
			content = lipgloss.JoinVertical(lipgloss.Left, detailPanel, statusBar)
		} else if m.isNarrow() {
			treePanel := m.renderTreePanel(m.width, contentHeight)
			content = lipgloss.JoinVertical(lipgloss.Left, treePanel, statusBar)
		} else {
			treeWidth := m.treeWidth()
			detailWidth := m.width - treeWidth - treeBorderRight

			treePanel := m.renderTreePanel(treeWidth, contentHeight)
			detailPanel := m.renderDetailPanel(detailWidth, contentHeight)
			body := lipgloss.JoinHorizontal(lipgloss.Top, treePanel, detailPanel)
			content = lipgloss.JoinVertical(lipgloss.Left, body, statusBar)
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// viewString returns the view content as a string for testing.
func (m Model) viewString() string {
	if !m.ready {
		return "Loading..."
	}

	if m.isTooSmall() {
		return fmt.Sprintf("Terminal too small (%dx%d). Minimum size: %dx%d.", m.width, m.height, minTermWidth, minTermHeight)
	}

	statusBar := m.renderStatusBar()
	contentHeight := m.height - lipgloss.Height(statusBar)
	if contentHeight < 1 {
		contentHeight = 1
	}

	if m.showHelp {
		helpPanel := m.renderHelpOverlay(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, helpPanel, statusBar)
	}
	if m.filtering {
		filterPanel := m.renderFilterOverlay(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, filterPanel, statusBar)
	}
	if m.showOverlay {
		detailPanel := m.renderDetailPanel(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, detailPanel, statusBar)
	}
	if m.isNarrow() {
		treePanel := m.renderTreePanel(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, treePanel, statusBar)
	}

	treeWidth := m.treeWidth()
	detailWidth := m.width - treeWidth - treeBorderRight
	treePanel := m.renderTreePanel(treeWidth, contentHeight)
	detailPanel := m.renderDetailPanel(detailWidth, contentHeight)
	body := lipgloss.JoinHorizontal(lipgloss.Top, treePanel, detailPanel)
	return lipgloss.JoinVertical(lipgloss.Left, body, statusBar)
}

func (m Model) treeWidth() int {
	w := m.width * treeWidthRatio / treeWidthDiv
	if w < minTreeWidth {
		w = minTreeWidth
	}
	if w > m.width {
		w = m.width
	}
	return w
}

func (m Model) renderTreePanel(width, height int) string {
	borderColor := colorBorderNormal
	if m.focusedPane == treePane {
		borderColor = colorAccentPrimary
	}
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)

	headerStyle := lipgloss.NewStyle().Bold(true)
	if m.focusedPane == treePane {
		headerStyle = headerStyle.Foreground(colorAccentPrimary)
	}
	header := headerStyle.Render("Beads")

	if m.tree == nil {
		content := header + "\n\n  (no beads loaded)"
		return style.Render(content)
	}

	visible := m.visibleNodes()
	if len(visible) == 0 {
		emptyMsg := "(no beads loaded)"
		if m.searchQuery != "" || m.hasActiveFilters() {
			emptyMsg = "(no matching beads)"
		}
		content := header + "\n\n  " + emptyMsg
		return style.Render(content)
	}

	selectedStyle := lipgloss.NewStyle().Reverse(true)

	// Viewport: available lines for tree rows (subtract 1 for header)
	viewportHeight := height - 1
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	var rows []string
	for i, node := range visible {
		row := m.renderTreeRow(node, visible, width)
		if i == m.selectedIdx {
			row = selectedStyle.Render(row)
		}
		rows = append(rows, row)
	}

	// Apply tree scroll to keep selection visible
	if len(rows) > viewportHeight {
		if len(rows) <= viewportHeight {
			// All rows fit, no scrolling needed
		} else {
			start := m.treeScroll
			end := start + viewportHeight
			if end > len(rows) {
				end = len(rows)
				start = end - viewportHeight
			}
			if start < 0 {
				start = 0
			}
			rows = rows[start:end]
		}
	}

	content := header + "\n" + strings.Join(rows, "\n")
	return style.Render(content)
}

// renderTreeRow renders a single tree row with indentation, tree chars, and bead info.
func (m Model) renderTreeRow(node *tree.Node, visible []*tree.Node, panelWidth int) string {
	var prefix string

	if node.Depth > 0 {
		// Build prefix: for each ancestor level, determine if we need a vertical bar or space
		parts := make([]string, node.Depth)
		current := node
		for d := node.Depth - 1; d >= 0; d-- {
			parent := findParent(current, visible)
			if parent != nil && !isLastChild(current, parent) {
				parts[d] = "│   "
			} else {
				parts[d] = "    "
			}
			current = parent
		}
		// Last part: connector for this node
		parent := findParent(node, visible)
		if parent != nil && isLastChild(node, parent) {
			parts[node.Depth-1] = "└── "
		} else {
			parts[node.Depth-1] = "├── "
		}
		prefix = strings.Join(parts, "")
	}

	// Expand/collapse indicator for nodes with children
	expandIndicator := ""
	if len(node.Children) > 0 {
		if node.Expanded {
			expandIndicator = "▼ "
		} else {
			expandIndicator = "▶ "
		}
	}

	typeLabel := shortType(node.Bead.IssueType)
	statusIcon := m.statusIcon(node.Bead.Status)

	id := node.Bead.ID
	if m.searchQuery != "" {
		id = highlightSearchMatches(id, m.searchQuery)
	}
	base := fmt.Sprintf("%s%s%s  %s  %s", prefix, expandIndicator, id, typeLabel, statusIcon)

	// Append progress indicator for nodes with children
	if len(node.Children) > 0 {
		done := 0
		total := len(node.Children)
		for _, child := range node.Children {
			if child.Bead.Status == "closed" {
				done++
			}
		}
		progress := fmt.Sprintf("[%d/%d]", done, total)
		var progressStyle lipgloss.Style
		switch {
		case done == total:
			progressStyle = lipgloss.NewStyle().Foreground(colorProgressDone)
		case done == 0:
			progressStyle = lipgloss.NewStyle().Faint(true)
		default:
			progressStyle = lipgloss.NewStyle().Foreground(colorProgressPartial)
		}
		base = base + "  " + progressStyle.Render(progress)
	}

	// Append truncated title if there's enough space
	if node.Bead.Title != "" && panelWidth > 0 {
		baseWidth := lipgloss.Width(base)
		available := panelWidth - baseWidth - 2 // 2 for "  " separator
		if available >= 10 {
			title := node.Bead.Title
			titleRunes := []rune(title)
			if len(titleRunes) > available {
				title = string(titleRunes[:available-1]) + "…"
			}
			titleStyle := lipgloss.NewStyle().Faint(true)
			styledTitle := titleStyle.Render(title)
			if m.searchQuery != "" {
				styledTitle = highlightSearchMatches(styledTitle, m.searchQuery)
			}
			base = base + "  " + styledTitle
		}
	}

	return base
}

// shortType returns abbreviated type labels.
func shortType(issueType string) string {
	switch issueType {
	case "feature":
		return "feat"
	case "decision":
		return "adr"
	default:
		return issueType
	}
}

// statusIcon returns the status with color and icon.
func (m Model) statusIcon(status string) string {
	var icon string
	var s lipgloss.Style

	switch status {
	case "open":
		icon = "○"
		s = lipgloss.NewStyle()
	case "in_progress":
		icon = "◉"
		s = lipgloss.NewStyle().Foreground(colorStatusInProgress)
	case "blocked":
		icon = "✗"
		s = lipgloss.NewStyle().Foreground(colorStatusBlocked)
	case "deferred":
		icon = "◌"
		s = lipgloss.NewStyle().Faint(true).Foreground(colorStatusDeferred)
	case "closed":
		icon = "✓"
		s = lipgloss.NewStyle().Foreground(colorStatusClosed)
	default:
		icon = "○"
		s = lipgloss.NewStyle()
	}

	return s.Render(icon)
}

// findParent finds the parent node of the given node in the tree.
func findParent(node *tree.Node, visible []*tree.Node) *tree.Node {
	if node.Bead.Parent == "" || node.Depth == 0 {
		return nil
	}
	for _, n := range visible {
		for _, child := range n.Children {
			if child == node {
				return n
			}
		}
	}
	return nil
}

// isLastChild checks if node is the last child of its parent.
func isLastChild(node *tree.Node, parent *tree.Node) bool {
	if parent == nil || len(parent.Children) == 0 {
		return false
	}
	return parent.Children[len(parent.Children)-1] == node
}

// colorStatus returns the status string with appropriate color styling.
func (m Model) colorStatus(status string) string {
	var s lipgloss.Style
	switch status {
	case "open":
		s = lipgloss.NewStyle()
	case "in_progress":
		s = lipgloss.NewStyle().Foreground(colorStatusInProgress)
	case "blocked":
		s = lipgloss.NewStyle().Foreground(colorStatusBlocked)
	case "deferred":
		s = lipgloss.NewStyle().Faint(true).Foreground(colorStatusDeferred)
	case "closed":
		s = lipgloss.NewStyle().Foreground(colorStatusClosed)
	default:
		s = lipgloss.NewStyle()
	}
	return s.Render(status)
}

func (m Model) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Background(colorStatusBarBg).
		Foreground(colorStatusBarFg)

	// Search input mode
	if m.searching {
		bar := fmt.Sprintf("Search: %s_", m.searchQuery)
		return style.Render(bar)
	}

	// Temporary status message (e.g., clipboard confirmation)
	if m.statusMsg != "" {
		return style.Render(m.statusMsg)
	}

	hints := []string{"[q] Quit", "[/] Search", "[f] Filter", "[?] Help"}
	left := strings.Join(hints, "  ")

	// Show active search query
	if m.searchQuery != "" {
		left = fmt.Sprintf("Search: %q  [Esc] Clear", m.searchQuery)
	}

	// Show active filters
	if m.hasActiveFilters() {
		var parts []string
		if len(m.filterTypes) > 0 {
			var types []string
			for _, t := range allTypes {
				if m.filterTypes[t] {
					types = append(types, t)
				}
			}
			parts = append(parts, "type="+strings.Join(types, ","))
		}
		if len(m.filterStats) > 0 {
			var statuses []string
			for _, s := range allStatuses {
				if m.filterStats[s] {
					statuses = append(statuses, s)
				}
			}
			parts = append(parts, "status="+strings.Join(statuses, ","))
		}
		filterStr := "Filter: " + strings.Join(parts, " ")
		if m.searchQuery != "" {
			left += "  " + filterStr
		} else {
			left = filterStr + "  [Esc] Clear"
		}
	}

	var right string
	if !m.lastRefresh.IsZero() {
		elapsed := m.nowFunc().Sub(m.lastRefresh)
		secs := int(elapsed.Seconds())
		right = fmt.Sprintf("Refreshed %ds ago", secs)
	} else {
		right = fmt.Sprintf("Refresh: %ds", m.config.Refresh)
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + right

	return style.Render(bar)
}
