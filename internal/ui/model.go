package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	config Config
	width  int
	height int
	ready  bool
}

// New creates a new Model with the given config.
func New(cfg Config) Model {
	return Model{
		config: cfg,
	}
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

	content := lipgloss.NewStyle().Faint(true).Render("Select a bead to view details")

	return style.Render(content)
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
