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

// Model is the top-level Bubble Tea model.
type Model struct {
	config       Config
	width        int
	height       int
	ready        bool
	selectedBead *data.Bead
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

	// Title line: bead ID as pane title
	titleStyle := lipgloss.NewStyle().Bold(true)
	lines = append(lines, titleStyle.Render(b.ID))
	lines = append(lines, strings.Repeat("─", width-2))

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
	lines = append(lines, strings.Repeat("─", width-2))

	return style.Render(strings.Join(lines, "\n"))
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
