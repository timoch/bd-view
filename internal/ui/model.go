package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/timoch/bd-view/internal/data"
)

// Config holds TUI configuration from CLI flags.
type Config struct {
	DBPath    string
	Refresh   int
	ExpandAll bool
	NoColor   bool
}

// paneID identifies which pane has focus.
type paneID int

const (
	treePane paneID = iota
	detailPane
)

// Model is the top-level Bubble Tea model.
type Model struct {
	config       Config
	width        int
	height       int
	ready        bool
	selectedBead *data.Bead
	dependents   []data.RelatedBead
	focusedPane  paneID
	detailScroll int
}

// New creates a new Model with the given config.
func New(cfg Config) Model {
	return Model{
		config: cfg,
	}
}

// SetSelectedBead sets the bead displayed in the detail pane.
// Pass nil to clear the selection.
func (m *Model) SetSelectedBead(b *data.Bead) {
	m.selectedBead = b
	m.dependents = nil
	m.detailScroll = 0
}

// SetSelectedBeadDetail sets the bead and its dependents for the detail pane.
func (m *Model) SetSelectedBeadDetail(detail *data.BeadDetail) {
	if detail == nil {
		m.selectedBead = nil
		m.dependents = nil
	} else {
		m.selectedBead = &detail.Bead
		m.dependents = detail.Dependents
	}
	m.detailScroll = 0
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.focusedPane == treePane {
				m.focusedPane = detailPane
			} else {
				m.focusedPane = treePane
			}
		case "j", "down":
			if m.focusedPane == detailPane {
				m.detailScroll++
			}
		case "k", "up":
			if m.focusedPane == detailPane && m.detailScroll > 0 {
				m.detailScroll--
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	}
	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	statusBar := m.renderStatusBar()
	contentHeight := m.height - lipgloss.Height(statusBar)
	if contentHeight < 1 {
		contentHeight = 1
	}

	treeWidth := m.treeWidth()
	detailWidth := m.width - treeWidth - 1 // 1 for border

	treePanel := m.renderTreePanel(treeWidth, contentHeight)
	detailPanel := m.renderDetailPanel(detailWidth, contentHeight)

	content := lipgloss.JoinHorizontal(lipgloss.Top, treePanel, detailPanel)

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (m Model) treeWidth() int {
	w := m.width * 2 / 5
	if w < 20 {
		w = 20
	}
	if w > m.width {
		w = m.width
	}
	return w
}

func (m Model) renderTreePanel(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder())

	header := lipgloss.NewStyle().Bold(true).Render("Beads")
	content := header + "\n\n  (no beads loaded)"

	return style.Render(content)
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

	b := m.selectedBead
	var lines []string
	contentWidth := width - 2 // account for padding

	// Title line: bead ID as pane title
	titleStyle := lipgloss.NewStyle().Bold(true)
	lines = append(lines, titleStyle.Render(b.ID))
	lines = append(lines, strings.Repeat("─", contentWidth))

	// Title field displayed prominently
	if b.Title != "" {
		lines = append(lines, fmt.Sprintf("Title:  %s", b.Title))
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
		rendered := renderMarkdown(sec.content, contentWidth)
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

	// Apply scroll offset
	allLines := strings.Split(strings.Join(lines, "\n"), "\n")
	scrollOffset := m.detailScroll
	if scrollOffset > len(allLines)-1 {
		scrollOffset = len(allLines) - 1
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	visible := allLines[scrollOffset:]
	if len(visible) > height {
		visible = visible[:height]
	}

	return style.Render(strings.Join(visible, "\n"))
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

// renderMarkdown applies basic terminal-friendly markdown rendering.
func renderMarkdown(text string, width int) string {
	boldStyle := lipgloss.NewStyle().Bold(true)
	codeBlockStyle := lipgloss.NewStyle().Faint(true)

	var result []string
	lines := strings.Split(text, "\n")
	inCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			result = append(result, codeBlockStyle.Render("  "+line))
			continue
		}

		// Wrap long lines
		wrapped := wrapLine(line, width)
		// Apply inline bold (**text**)
		wrapped = renderInlineBold(wrapped, boldStyle)
		result = append(result, wrapped)
	}

	return strings.Join(result, "\n")
}

// wrapLine wraps a single line to fit within the given width.
func wrapLine(line string, width int) string {
	if width <= 0 || len(line) <= width {
		return line
	}

	var result []string
	for len(line) > width {
		// Find last space before width
		breakAt := strings.LastIndex(line[:width], " ")
		if breakAt <= 0 {
			breakAt = width
		}
		result = append(result, line[:breakAt])
		line = line[breakAt:]
		if len(line) > 0 && line[0] == ' ' {
			line = line[1:]
		}
	}
	if line != "" {
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// renderInlineBold replaces **text** with bold-styled text.
func renderInlineBold(text string, boldStyle lipgloss.Style) string {
	for {
		start := strings.Index(text, "**")
		if start == -1 {
			break
		}
		end := strings.Index(text[start+2:], "**")
		if end == -1 {
			break
		}
		end += start + 2
		boldText := text[start+2 : end]
		text = text[:start] + boldStyle.Render(boldText) + text[end+2:]
	}
	return text
}

// colorStatus returns the status string with appropriate color styling.
func (m Model) colorStatus(status string) string {
	var s lipgloss.Style
	switch status {
	case "open":
		s = lipgloss.NewStyle()
	case "in_progress":
		s = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	case "blocked":
		s = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	case "deferred":
		s = lipgloss.NewStyle().Faint(true)
	case "closed":
		s = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	default:
		s = lipgloss.NewStyle()
	}
	return s.Render(status)
}

func (m Model) renderStatusBar() string {
	hints := []string{"[q] Quit", "[/] Search", "[f] Filter", "[?] Help"}
	left := strings.Join(hints, "  ")

	right := fmt.Sprintf("Refresh: %ds", m.config.Refresh)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + right

	style := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("252"))

	return style.Render(bar)
}
