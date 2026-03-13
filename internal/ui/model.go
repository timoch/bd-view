package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
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
	statusMsg          string // temporary status bar message (e.g., "Copied: bd-view-0ny.3")
	lastSelectedBeadID string // tracks bead ID to detect selection changes

	// Text selection state for detail panel copy
	selecting      bool // true while mouse drag is in progress
	selStartRow    int  // selection start row (in detail content lines, 0-based)
	selStartCol    int  // selection start column (in visible text, 0-based)
	selEndRow      int  // selection end row
	selEndCol      int  // selection end column
	hasSelection   bool // true when a completed selection exists to render
	detailLines    []string // rendered detail content lines (plain, post-scroll)
	detailPanelX   int  // x offset of detail panel in the terminal
	detailPanelY   int  // y offset of detail panel content (after header)
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
	if b != nil {
		m.lastSelectedBeadID = b.ID
	} else {
		m.lastSelectedBeadID = ""
	}
}

// SetSelectedBeadDetail sets the bead and its dependents for the detail pane.
func (m *Model) SetSelectedBeadDetail(detail *data.BeadDetail) {
	if detail == nil {
		m.selectedBead = nil
		m.dependents = nil
		m.lastSelectedBeadID = ""
	} else {
		m.selectedBead = &detail.Bead
		m.dependents = detail.Dependents
		m.lastSelectedBeadID = detail.Bead.ID
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
// using the native Bubble Tea v2 clipboard support (OSC 52).
func copyToClipboardCmd(text string) tea.Cmd {
	return tea.SetClipboard(text)
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

// isNarrow returns true when the terminal is too narrow for side-by-side layout.
func (m Model) isNarrow() bool {
	return m.width < 100
}

// isTooSmall returns true when the terminal is below the minimum supported size.
func (m Model) isTooSmall() bool {
	return m.width < 80 || m.height < 24
}
