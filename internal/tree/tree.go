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

func sortNodes(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Bead.Priority != nodes[j].Bead.Priority {
			return nodes[i].Bead.Priority < nodes[j].Bead.Priority
		}
		return nodes[i].Bead.ID < nodes[j].Bead.ID
	})
}

func setDepth(n *Node, depth int) {
	n.Depth = depth
	for _, child := range n.Children {
		setDepth(child, depth+1)
	}
}
