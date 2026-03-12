package tree

import (
	"sort"

	"github.com/timoch/bd-view/internal/data"
)

// Node represents a bead in the tree with UI state.
type Node struct {
	Bead     data.Bead
	Children []*Node
	Expanded bool
	Depth    int
}

// Model holds the tree structure built from a flat list of beads.
type Model struct {
	Roots    []*Node
	ByID     map[string]*Node
	allBeads []data.Bead
}

// BuildTree constructs a tree from a flat slice of beads.
// Top-level beads (no parent) become root nodes. Children are sorted
// by priority (ascending) then ID (alphabetical). Orphaned children
// (parent ID not found) are promoted to root nodes.
// If expandAll is true, all nodes start expanded; otherwise collapsed.
func BuildTree(beads []data.Bead, expandAll bool) *Model {
	m := &Model{
		ByID:     make(map[string]*Node, len(beads)),
		allBeads: beads,
	}

	// Create all nodes first.
	for _, b := range beads {
		m.ByID[b.ID] = &Node{
			Bead:     b,
			Expanded: expandAll,
		}
	}

	// Build parent-child relationships.
	for _, b := range beads {
		node := m.ByID[b.ID]
		if b.Parent != "" {
			if parent, ok := m.ByID[b.Parent]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
			// Orphan: parent ID not found, fall through to root.
		}
		m.Roots = append(m.Roots, node)
	}

	// Sort children and roots.
	sortNodes(m.Roots)
	for _, node := range m.ByID {
		if len(node.Children) > 0 {
			sortNodes(node.Children)
		}
	}

	// Set depths.
	for _, root := range m.Roots {
		setDepth(root, 0)
	}

	return m
}

// FlattenVisible returns an ordered slice of nodes that are currently visible,
// respecting expand/collapse state. A node's children are visible only if the
// node is expanded.
func (m *Model) FlattenVisible() []*Node {
	var result []*Node
	for _, root := range m.Roots {
		flattenNode(root, &result)
	}
	return result
}

// ToggleExpand toggles the expand/collapse state of the node with the given ID.
// Returns true if the node was found.
func (m *Model) ToggleExpand(id string) bool {
	node, ok := m.ByID[id]
	if !ok {
		return false
	}
	node.Expanded = !node.Expanded
	return true
}

// ExpandAll expands all nodes in the tree.
func (m *Model) ExpandAll() {
	for _, node := range m.ByID {
		node.Expanded = true
	}
}

// CollapseAll collapses all nodes in the tree.
func (m *Model) CollapseAll() {
	for _, node := range m.ByID {
		node.Expanded = false
	}
}

func flattenNode(n *Node, result *[]*Node) {
	*result = append(*result, n)
	if !n.Expanded {
		return
	}
	for _, child := range n.Children {
		flattenNode(child, result)
	}
}

// sortNodes sorts sibling nodes by dependency-aware topological order.
// Nodes with no incoming "blocks" edges from other siblings come first.
// Ties are broken by priority (ascending) then ID (alphabetical).
// Cycles fall back to priority+ID sort for the affected nodes.
func sortNodes(nodes []*Node) {
	if len(nodes) <= 1 {
		return
	}

	// Build set of sibling IDs for quick lookup.
	siblingSet := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		siblingSet[n.Bead.ID] = true
	}

	// Build adjacency: blockedBy[A] = {B, C} means B and C block A (A depends on B and C).
	// Only consider "blocks" dependencies where both ends are siblings.
	blockedBy := make(map[string]map[string]bool)
	for _, n := range nodes {
		for _, dep := range n.Bead.Dependencies {
			if dep.Type == "blocks" && siblingSet[dep.DependsOnID] {
				if blockedBy[n.Bead.ID] == nil {
					blockedBy[n.Bead.ID] = make(map[string]bool)
				}
				blockedBy[n.Bead.ID][dep.DependsOnID] = true
			}
		}
	}

	// If no dependency relationships among siblings, use simple sort.
	if len(blockedBy) == 0 {
		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].Bead.Priority != nodes[j].Bead.Priority {
				return nodes[i].Bead.Priority < nodes[j].Bead.Priority
			}
			return nodes[i].Bead.ID < nodes[j].Bead.ID
		})
		return
	}

	// Topological sort (Kahn's algorithm) with priority+ID tie-breaking.
	// inDegree counts how many sibling blockers each node has.
	inDegree := make(map[string]int, len(nodes))
	for _, n := range nodes {
		if _, ok := inDegree[n.Bead.ID]; !ok {
			inDegree[n.Bead.ID] = 0
		}
		if deps, ok := blockedBy[n.Bead.ID]; ok {
			inDegree[n.Bead.ID] = len(deps)
		}
	}

	// Collect nodes with no incoming edges, sorted by priority+ID.
	var ready []*Node
	for _, n := range nodes {
		if inDegree[n.Bead.ID] == 0 {
			ready = append(ready, n)
		}
	}
	sort.Slice(ready, func(i, j int) bool {
		if ready[i].Bead.Priority != ready[j].Bead.Priority {
			return ready[i].Bead.Priority < ready[j].Bead.Priority
		}
		return ready[i].Bead.ID < ready[j].Bead.ID
	})

	// Build forward edges: blocks[B] = {A, C} means B blocks A and C.
	blocks := make(map[string][]string)
	for blocked, blockers := range blockedBy {
		for blocker := range blockers {
			blocks[blocker] = append(blocks[blocker], blocked)
		}
	}

	// Process queue.
	nodeByID := make(map[string]*Node, len(nodes))
	for _, n := range nodes {
		nodeByID[n.Bead.ID] = n
	}

	var result []*Node
	for len(ready) > 0 {
		// Pop first (highest priority / earliest ID).
		n := ready[0]
		ready = ready[1:]
		result = append(result, n)

		// Release nodes blocked by this one.
		for _, blockedID := range blocks[n.Bead.ID] {
			inDegree[blockedID]--
			if inDegree[blockedID] == 0 {
				ready = append(ready, nodeByID[blockedID])
				// Re-sort ready list to maintain priority+ID ordering.
				sort.Slice(ready, func(i, j int) bool {
					if ready[i].Bead.Priority != ready[j].Bead.Priority {
						return ready[i].Bead.Priority < ready[j].Bead.Priority
					}
					return ready[i].Bead.ID < ready[j].Bead.ID
				})
			}
		}
	}

	// Cycle detection: if result is shorter than nodes, some nodes are in a cycle.
	// Append remaining nodes sorted by priority+ID (graceful fallback).
	if len(result) < len(nodes) {
		placed := make(map[string]bool, len(result))
		for _, n := range result {
			placed[n.Bead.ID] = true
		}
		var remaining []*Node
		for _, n := range nodes {
			if !placed[n.Bead.ID] {
				remaining = append(remaining, n)
			}
		}
		sort.Slice(remaining, func(i, j int) bool {
			if remaining[i].Bead.Priority != remaining[j].Bead.Priority {
				return remaining[i].Bead.Priority < remaining[j].Bead.Priority
			}
			return remaining[i].Bead.ID < remaining[j].Bead.ID
		})
		result = append(result, remaining...)
	}

	// Copy result back into the original slice.
	copy(nodes, result)
}

func setDepth(n *Node, depth int) {
	n.Depth = depth
	for _, child := range n.Children {
		setDepth(child, depth+1)
	}
}
