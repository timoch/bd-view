package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil
	case tea.KeyPressMsg:
		// Clear any temporary status message and selection on keypress
		m.statusMsg = ""
		m.hasSelection = false
		m.selecting = false

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
			case "space":
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
			switch msg.Code {
			case tea.KeyEscape:
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
			default:
				if msg.Text != "" {
					m.searchQuery += msg.Text
					m.selectedIdx = 0
					m.treeScroll = 0
					m.syncSelectedBead()
				}
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
	case tea.MouseWheelMsg:
		// Handle mouse wheel scroll events
		// Clear selection on scroll
		m.hasSelection = false
		m.selecting = false

		scrollStep := 3
		scrollUp := msg.Button == tea.MouseWheelUp

		// Help overlay: scroll help content
		if m.showHelp {
			if scrollUp {
				m.helpScroll -= scrollStep
				if m.helpScroll < 0 {
					m.helpScroll = 0
				}
			} else {
				m.helpScroll += scrollStep
			}
			return m, nil
		}

		// Filter overlay: ignore scroll (no scrollable content)
		if m.filtering {
			return m, nil
		}

		// Determine target panel
		scrollTree := false
		if m.showOverlay {
			// Overlay mode: scroll detail content
			scrollTree = false
		} else if m.isNarrow() {
			// Narrow mode: only tree visible
			scrollTree = true
		} else {
			tw := m.treeWidth()
			scrollTree = msg.X < tw
		}

		if scrollTree {
			visible := m.visibleNodes()
			viewportHeight := m.height - 1 // subtract header
			if viewportHeight < 1 {
				viewportHeight = 1
			}
			maxScroll := len(visible) - viewportHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if scrollUp {
				m.treeScroll -= scrollStep
				if m.treeScroll < 0 {
					m.treeScroll = 0
				}
			} else {
				m.treeScroll += scrollStep
				if m.treeScroll > maxScroll {
					m.treeScroll = maxScroll
				}
			}
		} else {
			if scrollUp {
				m.detailScroll -= scrollStep
				if m.detailScroll < 0 {
					m.detailScroll = 0
				}
			} else {
				m.detailScroll += scrollStep
			}
		}
		return m, nil

	case tea.MouseReleaseMsg:
		// Refresh detail lines for text extraction (View() value receiver loses them)
		m.refreshDetailLines()
		// Handle mouse button release — finalize selection (copy happens on right-click)
		if m.selecting {
			m.selecting = false
			startRow, startCol, endRow, endCol := m.selectionNormalized()
			if startRow != endRow || startCol != endCol {
				m.hasSelection = true
			}
		}
		return m, nil

	case tea.MouseMotionMsg:
		// Handle mouse drag (motion with left button held) — text selection in detail panel
		if m.selecting {
			row, col := m.screenToDetailCoord(msg.X, msg.Y)
			m.selEndRow = row
			m.selEndCol = col
		}
		return m, nil

	case tea.MouseClickMsg:
		// Handle right-click — copy selection to clipboard and clear it
		if msg.Button == tea.MouseRight && m.hasSelection {
			m.refreshDetailLines()
			text := m.extractSelectedText()
			m.hasSelection = false
			m.selecting = false
			if text != "" {
				lineCount := strings.Count(text, "\n") + 1
				m.statusMsg = fmt.Sprintf("Copied %d line(s)", lineCount)
				return m, tea.Batch(
					copyToClipboardCmd(text),
					tea.Tick(3*time.Second, func(time.Time) tea.Msg {
						return clearStatusMsg{}
					}),
				)
			}
			return m, nil
		}

		// Handle left-click press events
		if msg.Button == tea.MouseLeft {
			// Clear previous selection on any click
			m.hasSelection = false
			m.selecting = false

			// Ignore clicks in filter/help/search modes
			if m.showHelp || m.filtering || m.searching {
				return m, nil
			}

			// Determine if click is in detail panel for text selection
			inDetailPanel := false
			if m.showOverlay {
				// Overlay mode: detail takes full width
				inDetailPanel = true
			} else if !m.isNarrow() {
				tw := m.treeWidth()
				if msg.X >= tw+1 { // +1 for border
					inDetailPanel = true
				}
			}

			if inDetailPanel && m.selectedBead != nil {
				// Start text selection in detail panel
				m.selecting = true
				row, col := m.screenToDetailCoord(msg.X, msg.Y)
				m.selStartRow = row
				m.selStartCol = col
				m.selEndRow = row
				m.selEndCol = col
				if !m.showOverlay {
					m.focusedPane = detailPane
				}
				return m, nil
			}

			// In overlay mode with no bead, ignore click
			if m.showOverlay {
				return m, nil
			}

			// Tree panel click handling
			inTreePanel := false
			if m.isNarrow() {
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
				clickedRow := msg.Y - 1 + m.treeScroll
				if msg.Y >= 1 {
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
