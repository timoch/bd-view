package tree

import (
	"testing"

	"github.com/timoch/bd-view/internal/data"
)

func beads(specs ...struct {
	id, parent string
	priority   int
}) []data.Bead {
	var out []data.Bead
	for _, s := range specs {
		out = append(out, data.Bead{ID: s.id, Parent: s.parent, Priority: s.priority})
	}
	return out
}

func bead(id, parent string, priority int) struct {
	id, parent string
	priority   int
} {
	return struct {
		id, parent string
		priority   int
	}{id, parent, priority}
}

func TestBuildTree_BasicHierarchy(t *testing.T) {
	b := beads(
		bead("epic-1", "", 1),
		bead("task-1.1", "epic-1", 2),
		bead("task-1.2", "epic-1", 1),
	)

	m := BuildTree(b, false)

	if len(m.Roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(m.Roots))
	}
	if m.Roots[0].Bead.ID != "epic-1" {
		t.Errorf("expected root epic-1, got %s", m.Roots[0].Bead.ID)
	}

	children := m.Roots[0].Children
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	// Sorted by priority: task-1.2 (priority 1) before task-1.1 (priority 2)
	if children[0].Bead.ID != "task-1.2" {
		t.Errorf("expected first child task-1.2, got %s", children[0].Bead.ID)
	}
	if children[1].Bead.ID != "task-1.1" {
		t.Errorf("expected second child task-1.1, got %s", children[1].Bead.ID)
	}
}

func TestBuildTree_SortByPriorityThenID(t *testing.T) {
	b := beads(
		bead("parent", "", 0),
		bead("c", "parent", 1),
		bead("a", "parent", 1),
		bead("b", "parent", 1),
	)

	m := BuildTree(b, false)
	children := m.Roots[0].Children

	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		if children[i].Bead.ID != exp {
			t.Errorf("child[%d]: expected %s, got %s", i, exp, children[i].Bead.ID)
		}
	}
}

func TestBuildTree_OrphanedChildren(t *testing.T) {
	b := beads(
		bead("root-1", "", 1),
		bead("orphan-1", "nonexistent", 1),
	)

	m := BuildTree(b, false)

	if len(m.Roots) != 2 {
		t.Fatalf("expected 2 roots (including orphan), got %d", len(m.Roots))
	}
	ids := []string{m.Roots[0].Bead.ID, m.Roots[1].Bead.ID}
	if ids[0] != "orphan-1" || ids[1] != "root-1" {
		t.Errorf("expected roots [orphan-1, root-1] (sorted by ID at same priority), got %v", ids)
	}
}

func TestBuildTree_ExpandAll(t *testing.T) {
	b := beads(
		bead("epic-1", "", 1),
		bead("task-1.1", "epic-1", 1),
	)

	m := BuildTree(b, true)

	for id, node := range m.ByID {
		if !node.Expanded {
			t.Errorf("node %s should be expanded with expandAll=true", id)
		}
	}
}

func TestBuildTree_DefaultCollapsed(t *testing.T) {
	b := beads(
		bead("epic-1", "", 1),
		bead("task-1.1", "epic-1", 1),
	)

	m := BuildTree(b, false)

	for id, node := range m.ByID {
		if node.Expanded {
			t.Errorf("node %s should be collapsed with expandAll=false", id)
		}
	}
}

func TestBuildTree_Depth(t *testing.T) {
	b := beads(
		bead("root", "", 0),
		bead("child", "root", 0),
		bead("grandchild", "child", 0),
	)

	m := BuildTree(b, false)

	tests := map[string]int{
		"root":       0,
		"child":      1,
		"grandchild": 2,
	}
	for id, expected := range tests {
		if m.ByID[id].Depth != expected {
			t.Errorf("node %s: expected depth %d, got %d", id, expected, m.ByID[id].Depth)
		}
	}
}

func TestFlattenVisible_AllCollapsed(t *testing.T) {
	b := beads(
		bead("epic-1", "", 1),
		bead("task-1.1", "epic-1", 1),
		bead("epic-2", "", 2),
	)

	m := BuildTree(b, false)
	visible := m.FlattenVisible()

	if len(visible) != 2 {
		t.Fatalf("expected 2 visible (roots only), got %d", len(visible))
	}
	if visible[0].Bead.ID != "epic-1" || visible[1].Bead.ID != "epic-2" {
		t.Errorf("expected [epic-1, epic-2], got [%s, %s]", visible[0].Bead.ID, visible[1].Bead.ID)
	}
}

func TestFlattenVisible_Expanded(t *testing.T) {
	b := beads(
		bead("epic-1", "", 1),
		bead("task-1.1", "epic-1", 1),
		bead("task-1.2", "epic-1", 2),
		bead("epic-2", "", 2),
	)

	m := BuildTree(b, false)
	m.ByID["epic-1"].Expanded = true

	visible := m.FlattenVisible()

	if len(visible) != 4 {
		t.Fatalf("expected 4 visible, got %d", len(visible))
	}
	expected := []string{"epic-1", "task-1.1", "task-1.2", "epic-2"}
	for i, exp := range expected {
		if visible[i].Bead.ID != exp {
			t.Errorf("visible[%d]: expected %s, got %s", i, exp, visible[i].Bead.ID)
		}
	}
}

func TestFlattenVisible_MixedExpandCollapse(t *testing.T) {
	b := beads(
		bead("root", "", 0),
		bead("child-1", "root", 1),
		bead("grandchild-1", "child-1", 1),
		bead("child-2", "root", 2),
	)

	m := BuildTree(b, false)
	m.ByID["root"].Expanded = true
	// child-1 stays collapsed, so grandchild-1 should not be visible

	visible := m.FlattenVisible()

	expected := []string{"root", "child-1", "child-2"}
	if len(visible) != len(expected) {
		t.Fatalf("expected %d visible, got %d", len(expected), len(visible))
	}
	for i, exp := range expected {
		if visible[i].Bead.ID != exp {
			t.Errorf("visible[%d]: expected %s, got %s", i, exp, visible[i].Bead.ID)
		}
	}
}

func TestFlattenVisible_DeepExpand(t *testing.T) {
	b := beads(
		bead("root", "", 0),
		bead("child", "root", 0),
		bead("grandchild", "child", 0),
	)

	m := BuildTree(b, true) // all expanded

	visible := m.FlattenVisible()

	expected := []string{"root", "child", "grandchild"}
	if len(visible) != len(expected) {
		t.Fatalf("expected %d visible, got %d", len(expected), len(visible))
	}
	for i, exp := range expected {
		if visible[i].Bead.ID != exp {
			t.Errorf("visible[%d]: expected %s, got %s", i, exp, visible[i].Bead.ID)
		}
	}
}

func TestToggleExpand(t *testing.T) {
	b := beads(
		bead("root", "", 0),
		bead("child", "root", 0),
	)

	m := BuildTree(b, false)

	if m.ByID["root"].Expanded {
		t.Fatal("root should start collapsed")
	}

	ok := m.ToggleExpand("root")
	if !ok {
		t.Fatal("ToggleExpand should return true for existing node")
	}
	if !m.ByID["root"].Expanded {
		t.Error("root should be expanded after toggle")
	}

	m.ToggleExpand("root")
	if m.ByID["root"].Expanded {
		t.Error("root should be collapsed after second toggle")
	}
}

func TestToggleExpand_NotFound(t *testing.T) {
	m := BuildTree(nil, false)
	if m.ToggleExpand("nonexistent") {
		t.Error("ToggleExpand should return false for nonexistent node")
	}
}

func TestExpandAll_CollapseAll(t *testing.T) {
	b := beads(
		bead("root", "", 0),
		bead("child", "root", 0),
		bead("grandchild", "child", 0),
	)

	m := BuildTree(b, false)

	m.ExpandAll()
	for id, node := range m.ByID {
		if !node.Expanded {
			t.Errorf("after ExpandAll, %s should be expanded", id)
		}
	}

	m.CollapseAll()
	for id, node := range m.ByID {
		if node.Expanded {
			t.Errorf("after CollapseAll, %s should be collapsed", id)
		}
	}
}

func TestBuildTree_EmptyInput(t *testing.T) {
	m := BuildTree(nil, false)

	if len(m.Roots) != 0 {
		t.Errorf("expected 0 roots for nil input, got %d", len(m.Roots))
	}
	if len(m.ByID) != 0 {
		t.Errorf("expected empty ByID map, got %d entries", len(m.ByID))
	}

	visible := m.FlattenVisible()
	if len(visible) != 0 {
		t.Errorf("expected 0 visible for empty tree, got %d", len(visible))
	}
}

func TestBuildTree_MultipleRootsSorted(t *testing.T) {
	b := beads(
		bead("z-task", "", 2),
		bead("a-epic", "", 1),
		bead("m-feat", "", 1),
	)

	m := BuildTree(b, false)

	if len(m.Roots) != 3 {
		t.Fatalf("expected 3 roots, got %d", len(m.Roots))
	}
	// priority 1: a-epic, m-feat; priority 2: z-task
	expected := []string{"a-epic", "m-feat", "z-task"}
	for i, exp := range expected {
		if m.Roots[i].Bead.ID != exp {
			t.Errorf("roots[%d]: expected %s, got %s", i, exp, m.Roots[i].Bead.ID)
		}
	}
}
