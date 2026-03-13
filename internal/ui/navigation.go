package ui

import (
	"strings"

	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/tree"
)

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
// Resets detailScroll to 0 only when the selected bead changes;
// preserves it (clamped) when the same bead is re-selected (e.g., on refresh).
func (m *Model) syncSelectedBead() {
	visible := m.visibleNodes()
	if len(visible) == 0 {
		m.selectedBead = nil
		m.dependents = nil
		m.lastSelectedBeadID = ""
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
	if b.ID != m.lastSelectedBeadID {
		m.detailScroll = 0
		m.lastSelectedBeadID = b.ID
	}
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
