package ui

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/tree"
)

// beadsLoadedMsg carries the result of a bead list fetch.
type beadsLoadedMsg struct {
	beads []data.Bead
	err   error
}

// tickMsg signals that the refresh interval has elapsed.
type tickMsg time.Time

// clearStatusMsg signals that the status bar message should be cleared.
type clearStatusMsg struct{}

// Config holds TUI configuration from CLI flags.
type Config struct {
	DBPath            string
	Refresh           int
	ExpandAll         bool
	ExpandAllExplicit bool // true when --expand-all was explicitly passed
	NoColor           bool
	FilterTypes       []string
	FilterStatuses    []string
	StatePath         string // path to bd-view-state.json for persist
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
	tree         *tree.Model
	selectedIdx  int
	treeScroll   int
	selectedBead *data.Bead
	dependents   []data.RelatedBead
	focusedPane  paneID
	detailScroll int
	showOverlay  bool // full-screen detail overlay in narrow mode
	searching    bool // true when search input is active
	searchQuery  string
	filtering    bool           // true when filter overlay is shown
	filterTypes  map[string]bool // selected type filters (OR within)
	filterStats  map[string]bool // selected status filters (OR within)
	filterCursor int            // cursor position in filter menu
	showHelp     bool           // true when help overlay is shown
	helpScroll   int            // scroll offset within help overlay
	fetcher      *data.Fetcher  // fetcher for bd CLI data
	beads        []data.Bead    // current in-memory bead list
	lastRefresh  time.Time      // time of last successful refresh
	nowFunc      func() time.Time // for testing; defaults to time.Now
	statusMsg    string           // temporary status bar message (e.g., "Copied: bd-view-0ny.3")
}

// allTypes lists the bead types in display order.
var allTypes = []string{"task", "bug", "feature", "chore", "epic", "decision"}

// allStatuses lists the bead statuses in display order.
var allStatuses = []string{"open", "in_progress", "blocked", "deferred", "closed"}

// filterMenuItems returns the combined list of filter menu items (types then statuses).
// Each item is a (label, section) pair where section is 0 for type, 1 for status.
func filterMenuItems() []filterMenuItem {
	var items []filterMenuItem
	for _, t := range allTypes {
		items = append(items, filterMenuItem{label: t, section: 0})
	}
	for _, s := range allStatuses {
		items = append(items, filterMenuItem{label: s, section: 1})
	}
	return items
}

type filterMenuItem struct {
	label   string
	section int // 0 = type, 1 = status
}

// New creates a new Model with the given config.
func New(cfg Config) Model {
	m := Model{
		config:      cfg,
		filterTypes: make(map[string]bool),
		filterStats: make(map[string]bool),
		nowFunc:     time.Now,
	}
	for _, t := range cfg.FilterTypes {
		m.filterTypes[t] = true
	}
	for _, s := range cfg.FilterStatuses {
		m.filterStats[s] = true
	}
	return m
}

// SetFetcher sets the data fetcher for refresh operations.
func (m *Model) SetFetcher(f *data.Fetcher) {
	m.fetcher = f
}

// SetTree sets the tree model for rendering.
func (m *Model) SetTree(t *tree.Model) {
	m.tree = t
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
	var cmds []tea.Cmd
	if m.fetcher != nil {
		cmds = append(cmds, m.fetchBeadsCmd())
		cmds = append(cmds, m.tickCmd())
	}
	return tea.Batch(cmds...)
}

// copyToClipboardCmd returns a command that copies text to the system clipboard
// using the OSC 52 escape sequence. This works in most modern terminals and
// over SSH with forwarding enabled.
func copyToClipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		b64 := base64.StdEncoding.EncodeToString([]byte(text))
		seq := fmt.Sprintf("\033]52;c;%s\a", b64)
		fmt.Print(seq)
		return nil
	}
}

// fetchBeadsCmd returns a command that fetches the bead list.
func (m Model) fetchBeadsCmd() tea.Cmd {
	fetcher := m.fetcher
	return func() tea.Msg {
		beads, err := fetcher.ListAll(context.Background())
		return beadsLoadedMsg{beads: beads, err: err}
	}
}

// tickCmd returns a command that sends a tickMsg after the refresh interval.
func (m Model) tickCmd() tea.Cmd {
	d := time.Duration(m.config.Refresh) * time.Second
	if d <= 0 {
		d = 2 * time.Second
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil
	case tea.KeyMsg:
		// Clear any temporary status message on keypress
		m.statusMsg = ""

		// Help overlay takes precedence over all other modes
		if m.showHelp {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc", "?":
				m.showHelp = false
				m.helpScroll = 0
			case "j", "down":
				m.helpScroll++
			case "k", "up":
				if m.helpScroll > 0 {
					m.helpScroll--
				}
			}
			return m, nil
		}

		// In overlay mode, handle overlay-specific keys first
		if m.showOverlay {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.showOverlay = false
			case "j", "down":
				m.detailScroll++
			case "k", "up":
				if m.detailScroll > 0 {
					m.detailScroll--
				}
			}
			return m, nil
		}

		// In filter overlay mode, handle filter-specific keys
		if m.filtering {
			items := filterMenuItems()
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.filtering = false
				// Clear all filters
				m.filterTypes = make(map[string]bool)
				m.filterStats = make(map[string]bool)
				m.selectedIdx = 0
				m.treeScroll = 0
				m.syncSelectedBead()
			case "enter", "f":
				m.filtering = false
			case "j", "down":
				if m.filterCursor < len(items)-1 {
					m.filterCursor++
				}
			case "k", "up":
				if m.filterCursor > 0 {
					m.filterCursor--
				}
			case " ":
				if m.filterCursor < len(items) {
					item := items[m.filterCursor]
					if item.section == 0 {
						if m.filterTypes[item.label] {
							delete(m.filterTypes, item.label)
						} else {
							m.filterTypes[item.label] = true
						}
					} else {
						if m.filterStats[item.label] {
							delete(m.filterStats, item.label)
						} else {
							m.filterStats[item.label] = true
						}
					}
					m.selectedIdx = 0
					m.treeScroll = 0
					m.syncSelectedBead()
				}
			}
			return m, nil
		}

		// In search mode, capture input
		if m.searching {
			switch msg.Type {
			case tea.KeyEsc:
				m.searching = false
				m.searchQuery = ""
				m.selectedIdx = 0
				m.treeScroll = 0
				m.syncSelectedBead()
			case tea.KeyEnter:
				m.searching = false
				// Keep the search query active, just exit input mode
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.selectedIdx = 0
					m.treeScroll = 0
					m.syncSelectedBead()
				}
			case tea.KeyRunes:
				m.searchQuery += string(msg.Runes)
				m.selectedIdx = 0
				m.treeScroll = 0
				m.syncSelectedBead()
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.selectedIdx = 0
				m.treeScroll = 0
				m.syncSelectedBead()
			} else if m.hasActiveFilters() {
				m.filterTypes = make(map[string]bool)
				m.filterStats = make(map[string]bool)
				m.selectedIdx = 0
				m.treeScroll = 0
				m.syncSelectedBead()
			}
		case "/":
			m.searching = true
			return m, nil
		case "?":
			m.showHelp = true
			return m, nil
		case "r":
			if m.fetcher != nil {
				return m, m.fetchBeadsCmd()
			}
			return m, nil
		case "f":
			if m.focusedPane == treePane {
				m.filtering = true
				m.filterCursor = 0
				return m, nil
			}
		case "tab":
			if !m.isNarrow() {
				if m.focusedPane == treePane {
					m.focusedPane = detailPane
				} else {
					m.focusedPane = treePane
				}
			}
		case "j", "down":
			if m.focusedPane == detailPane {
				m.detailScroll++
			} else {
				m.moveSelectionDown()
			}
		case "k", "up":
			if m.focusedPane == detailPane {
				if m.detailScroll > 0 {
					m.detailScroll--
				}
			} else {
				m.moveSelectionUp()
			}
		case "enter":
			if m.focusedPane == treePane {
				if m.isNarrow() && m.selectedBead != nil {
					m.showOverlay = true
					m.detailScroll = 0
				} else {
					m.expandSelected()
				}
			}
		case "right":
			if m.focusedPane == treePane {
				m.expandSelected()
			}
		case "left":
			if m.focusedPane == treePane {
				m.collapseOrMoveToParent()
			}
		case "g":
			if m.focusedPane == treePane {
				m.goToTop()
			}
		case "G":
			if m.focusedPane == treePane {
				m.goToBottom()
			}
		case "e":
			if m.focusedPane == treePane {
				m.expandAllNodes()
			}
		case "c":
			if m.focusedPane == treePane {
				m.collapseAllNodes()
			}
		case "y":
			if m.focusedPane == treePane && m.selectedBead != nil {
				id := m.selectedBead.ID
				m.statusMsg = fmt.Sprintf("Copied: %s", id)
				return m, tea.Batch(
					copyToClipboardCmd(id),
					tea.Tick(3*time.Second, func(time.Time) tea.Msg {
						return clearStatusMsg{}
					}),
				)
			}
		}
	case tea.MouseMsg:
		// Only handle left-click press events
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			// Ignore clicks in overlay/filter/help/search modes
			if m.showHelp || m.showOverlay || m.filtering || m.searching {
				return m, nil
			}

			inTreePanel := false
			if m.isNarrow() {
				// In narrow mode, tree takes full width
				inTreePanel = true
			} else {
				tw := m.treeWidth()
				if msg.X < tw {
					inTreePanel = true
				} else {
					m.focusedPane = detailPane
				}
			}

			if inTreePanel {
				m.focusedPane = treePane
				// Calculate clicked row index: subtract 1 for header, add scroll offset
				clickedRow := msg.Y - 1 + m.treeScroll
				if msg.Y >= 1 { // Ignore clicks on header row
					visible := m.visibleNodes()
					if clickedRow >= 0 && clickedRow < len(visible) {
						m.selectedIdx = clickedRow
						m.syncSelectedBead()
					}
				}
			}
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	case beadsLoadedMsg:
		if msg.err == nil {
			m.applyRefresh(msg.beads)
		}
	case tickMsg:
		var cmds []tea.Cmd
		if m.fetcher != nil {
			cmds = append(cmds, m.fetchBeadsCmd())
		}
		cmds = append(cmds, m.tickCmd())
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// hasActiveFilters returns true if any type or status filters are set.
func (m *Model) hasActiveFilters() bool {
	return len(m.filterTypes) > 0 || len(m.filterStats) > 0
}

// visibleNodes returns the current visible node list, filtered by search query and type/status filters.
func (m *Model) visibleNodes() []*tree.Node {
	if m.tree == nil {
		return nil
	}
	visible := m.tree.FlattenVisible()
	if m.hasActiveFilters() {
		visible = m.filterByTypeStatus(visible)
	}
	if m.searchQuery != "" {
		visible = m.filterBySearch(visible)
	}
	return visible
}

// filterByTypeStatus returns nodes matching the active type/status filters plus their ancestors.
func (m *Model) filterByTypeStatus(visible []*tree.Node) []*tree.Node {
	// Find matching nodes
	matchIDs := make(map[string]bool)
	for _, node := range visible {
		typeMatch := len(m.filterTypes) == 0 || m.filterTypes[node.Bead.IssueType]
		statusMatch := len(m.filterStats) == 0 || m.filterStats[node.Bead.Status]
		if typeMatch && statusMatch {
			matchIDs[node.Bead.ID] = true
		}
	}

	// Collect ancestor IDs
	ancestorIDs := make(map[string]bool)
	for id := range matchIDs {
		if node, ok := m.tree.ByID[id]; ok {
			current := node
			for current.Bead.Parent != "" {
				ancestorIDs[current.Bead.Parent] = true
				if parent, ok := m.tree.ByID[current.Bead.Parent]; ok {
					current = parent
				} else {
					break
				}
			}
		}
	}

	var filtered []*tree.Node
	for _, node := range visible {
		if matchIDs[node.Bead.ID] || ancestorIDs[node.Bead.ID] {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// filterBySearch returns only nodes that match the search query or are ancestors
// of matching nodes, preserving hierarchy context.
func (m *Model) filterBySearch(visible []*tree.Node) []*tree.Node {
	query := strings.ToLower(m.searchQuery)

	// First pass: find all matching bead IDs
	matchIDs := make(map[string]bool)
	for _, node := range visible {
		b := node.Bead
		if strings.Contains(strings.ToLower(b.ID), query) ||
			strings.Contains(strings.ToLower(b.Title), query) ||
			strings.Contains(strings.ToLower(b.Description), query) ||
			strings.Contains(strings.ToLower(b.Design), query) ||
			strings.Contains(strings.ToLower(b.AcceptanceCriteria), query) ||
			strings.Contains(strings.ToLower(b.Notes), query) {
			matchIDs[node.Bead.ID] = true
		}
	}

	// Second pass: collect ancestor IDs of all matches
	ancestorIDs := make(map[string]bool)
	for id := range matchIDs {
		if node, ok := m.tree.ByID[id]; ok {
			current := node
			for current.Bead.Parent != "" {
				ancestorIDs[current.Bead.Parent] = true
				if parent, ok := m.tree.ByID[current.Bead.Parent]; ok {
					current = parent
				} else {
					break
				}
			}
		}
	}

	// Third pass: filter visible to only matches + ancestors
	var filtered []*tree.Node
	for _, node := range visible {
		if matchIDs[node.Bead.ID] || ancestorIDs[node.Bead.ID] {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// syncSelectedBead updates the selected bead from the current tree selection
// and adjusts treeScroll to keep the selection visible.
func (m *Model) syncSelectedBead() {
	visible := m.visibleNodes()
	if len(visible) == 0 {
		m.selectedBead = nil
		m.dependents = nil
		return
	}
	if m.selectedIdx >= len(visible) {
		m.selectedIdx = len(visible) - 1
	}
	if m.selectedIdx < 0 {
		m.selectedIdx = 0
	}
	b := visible[m.selectedIdx].Bead
	m.selectedBead = &b
	m.dependents = nil
	m.detailScroll = 0
	m.ensureSelectedVisible()
}

// ensureSelectedVisible adjusts treeScroll so the selected index is in the viewport.
func (m *Model) ensureSelectedVisible() {
	viewportHeight := m.height - 1 // subtract header line
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	if m.selectedIdx < m.treeScroll {
		m.treeScroll = m.selectedIdx
	}
	if m.selectedIdx >= m.treeScroll+viewportHeight {
		m.treeScroll = m.selectedIdx - viewportHeight + 1
	}
}

func (m *Model) moveSelectionDown() {
	visible := m.visibleNodes()
	if len(visible) == 0 {
		return
	}
	if m.selectedIdx < len(visible)-1 {
		m.selectedIdx++
		m.syncSelectedBead()
	}
}

func (m *Model) moveSelectionUp() {
	visible := m.visibleNodes()
	if len(visible) == 0 {
		return
	}
	if m.selectedIdx > 0 {
		m.selectedIdx--
		m.syncSelectedBead()
	}
}

func (m *Model) expandSelected() {
	visible := m.visibleNodes()
	if len(visible) == 0 || m.selectedIdx >= len(visible) {
		return
	}
	node := visible[m.selectedIdx]
	if len(node.Children) > 0 && !node.Expanded {
		m.tree.ToggleExpand(node.Bead.ID)
		m.persistExpandState()
	}
}

func (m *Model) collapseOrMoveToParent() {
	visible := m.visibleNodes()
	if len(visible) == 0 || m.selectedIdx >= len(visible) {
		return
	}
	node := visible[m.selectedIdx]
	// If expanded parent, collapse it
	if len(node.Children) > 0 && node.Expanded {
		m.tree.ToggleExpand(node.Bead.ID)
		m.persistExpandState()
		return
	}
	// Otherwise, move to parent
	if node.Bead.Parent != "" {
		for i, n := range visible {
			if n.Bead.ID == node.Bead.Parent {
				m.selectedIdx = i
				m.syncSelectedBead()
				return
			}
		}
	}
}

func (m *Model) goToTop() {
	visible := m.visibleNodes()
	if len(visible) == 0 {
		return
	}
	if m.selectedIdx != 0 {
		m.selectedIdx = 0
		m.syncSelectedBead()
	}
}

func (m *Model) goToBottom() {
	visible := m.visibleNodes()
	if len(visible) == 0 {
		return
	}
	last := len(visible) - 1
	if m.selectedIdx != last {
		m.selectedIdx = last
		m.syncSelectedBead()
	}
}

func (m *Model) expandAllNodes() {
	if m.tree == nil {
		return
	}
	m.tree.ExpandAll()
	m.persistExpandState()
	// Clamp selectedIdx to valid range
	visible := m.visibleNodes()
	if len(visible) == 0 {
		m.selectedIdx = 0
	} else if m.selectedIdx >= len(visible) {
		m.selectedIdx = len(visible) - 1
	}
}

func (m *Model) collapseAllNodes() {
	if m.tree == nil {
		return
	}
	// Remember selected bead ID to try to stay on it or its ancestor
	visible := m.visibleNodes()
	var selectedID string
	if m.selectedIdx < len(visible) {
		selectedID = visible[m.selectedIdx].Bead.ID
	}
	m.tree.CollapseAll()
	m.persistExpandState()
	// After collapsing, find the selected bead or its nearest ancestor in visible roots
	newVisible := m.visibleNodes()
	if len(newVisible) == 0 {
		m.selectedIdx = 0
		m.syncSelectedBead()
		return
	}
	m.selectedIdx = 0
	if selectedID != "" {
		// Try to find the bead itself (it might be a root)
		for i, n := range newVisible {
			if n.Bead.ID == selectedID {
				m.selectedIdx = i
				break
			}
		}
		// If not found, try to find the ancestor root
		if m.selectedIdx == 0 && len(newVisible) > 0 {
			if node, ok := m.tree.ByID[selectedID]; ok {
				// Walk up to find a visible ancestor
				current := node
				for current.Bead.Parent != "" {
					if parent, ok := m.tree.ByID[current.Bead.Parent]; ok {
						current = parent
					} else {
						break
					}
				}
				for i, n := range newVisible {
					if n.Bead.ID == current.Bead.ID {
						m.selectedIdx = i
						break
					}
				}
			}
		}
	}
	m.syncSelectedBead()
}

// applyRefresh applies new bead data, preserving UI state (selection, scroll, expand/collapse).
func (m *Model) applyRefresh(newBeads []data.Bead) {
	now := m.nowFunc()
	diff := data.DiffBeads(m.beads, newBeads)
	m.beads = newBeads
	m.lastRefresh = now

	if !diff.HasChanges() && m.tree != nil {
		// No changes, nothing to update
		return
	}

	// Remember current selection and expand state
	firstLoad := m.tree == nil
	var selectedID string
	expandState := make(map[string]bool)
	if m.tree != nil {
		visible := m.visibleNodes()
		if m.selectedIdx < len(visible) && m.selectedIdx >= 0 {
			selectedID = visible[m.selectedIdx].Bead.ID
		}
		for id, node := range m.tree.ByID {
			expandState[id] = node.Expanded
		}
	}

	// Rebuild tree
	newTree := tree.BuildTree(newBeads, m.config.ExpandAll)

	// Restore expand state from previous tree (in-session refresh)
	for id, expanded := range expandState {
		if node, ok := newTree.ByID[id]; ok {
			node.Expanded = expanded
		}
	}

	// On first load, apply persisted expand state unless --expand-all was explicitly passed
	if firstLoad && !m.config.ExpandAllExplicit && m.config.StatePath != "" {
		if expandedIDs := LoadExpandState(m.config.StatePath); expandedIDs != nil {
			// Override BuildTree defaults with persisted state
			for _, node := range newTree.ByID {
				node.Expanded = false
			}
			idSet := make(map[string]bool, len(expandedIDs))
			for _, id := range expandedIDs {
				idSet[id] = true
			}
			for id, node := range newTree.ByID {
				if idSet[id] {
					node.Expanded = true
				}
			}
		}
	}

	m.tree = newTree

	// Restore selection
	if selectedID != "" {
		if _, ok := m.tree.ByID[selectedID]; ok {
			// Selected bead still exists — find it in visible nodes
			visible := m.visibleNodes()
			for i, node := range visible {
				if node.Bead.ID == selectedID {
					m.selectedIdx = i
					m.syncSelectedBead()
					return
				}
			}
		}
		// Selected bead was deleted — find nearest neighbor
		// Try to keep same index position, clamped to range
	}

	// Fallback: clamp selection
	visible := m.visibleNodes()
	if m.selectedIdx >= len(visible) {
		m.selectedIdx = len(visible) - 1
	}
	if m.selectedIdx < 0 {
		m.selectedIdx = 0
	}
	m.syncSelectedBead()
}

// isNarrow returns true when the terminal is too narrow for side-by-side layout.
func (m Model) isNarrow() bool {
	return m.width < 100
}

// isTooSmall returns true when the terminal is below the minimum supported size.
func (m Model) isTooSmall() bool {
	return m.width < 80 || m.height < 24
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.isTooSmall() {
		return fmt.Sprintf("Terminal too small (%dx%d). Minimum size: 80x24.", m.width, m.height)
	}

	statusBar := m.renderStatusBar()
	contentHeight := m.height - lipgloss.Height(statusBar)
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Help overlay
	if m.showHelp {
		helpPanel := m.renderHelpOverlay(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, helpPanel, statusBar)
	}

	// Filter overlay
	if m.filtering {
		filterPanel := m.renderFilterOverlay(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, filterPanel, statusBar)
	}

	// Full-screen overlay in narrow mode
	if m.showOverlay {
		detailPanel := m.renderDetailPanel(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, detailPanel, statusBar)
	}

	if m.isNarrow() {
		// Narrow mode: tree only, full width
		treePanel := m.renderTreePanel(m.width, contentHeight)
		return lipgloss.JoinVertical(lipgloss.Left, treePanel, statusBar)
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

	base := fmt.Sprintf("%s%s%s  %s  %s", prefix, expandIndicator, node.Bead.ID, typeLabel, statusIcon)

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
			base = base + "  " + titleStyle.Render(title)
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
	if m.focusedPane == detailPane || m.showOverlay {
		titleStyle = titleStyle.Foreground(colorAccentPrimary)
	}
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
		rendered := m.renderMarkdown(sec.content, contentWidth)
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

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styleName),
		glamour.WithWordWrap(width),
		glamour.WithColorProfile(lipgloss.ColorProfile()),
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
