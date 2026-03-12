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

// --- Dependency-aware sort tests ---

func TestSortNodes_IndependentFirst(t *testing.T) {
	// B depends on A (A blocks B). A should come first.
	b := []data.Bead{
		{ID: "b", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "b", DependsOnID: "a", Type: "blocks"},
		}},
		{ID: "a", Parent: "p", Priority: 1},
	}

	m := BuildTree(append([]data.Bead{{ID: "p", Priority: 0}}, b...), false)
	children := m.Roots[0].Children

	if children[0].Bead.ID != "a" {
		t.Errorf("expected independent node 'a' first, got %s", children[0].Bead.ID)
	}
	if children[1].Bead.ID != "b" {
		t.Errorf("expected dependent node 'b' second, got %s", children[1].Bead.ID)
	}
}

func TestSortNodes_ChainOrder(t *testing.T) {
	// Chain: c depends on b, b depends on a. Expected: a, b, c.
	b := []data.Bead{
		{ID: "p", Priority: 0},
		{ID: "c", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "c", DependsOnID: "b", Type: "blocks"},
		}},
		{ID: "a", Parent: "p", Priority: 1},
		{ID: "b", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "b", DependsOnID: "a", Type: "blocks"},
		}},
	}

	m := BuildTree(b, false)
	children := m.Roots[0].Children

	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		if children[i].Bead.ID != exp {
			t.Errorf("child[%d]: expected %s, got %s", i, exp, children[i].Bead.ID)
		}
	}
}

func TestSortNodes_CycleFallback(t *testing.T) {
	// a depends on b, b depends on a — cycle. Should fall back to priority+ID.
	b := []data.Bead{
		{ID: "p", Priority: 0},
		{ID: "b", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "b", DependsOnID: "a", Type: "blocks"},
		}},
		{ID: "a", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "a", DependsOnID: "b", Type: "blocks"},
		}},
	}

	m := BuildTree(b, false)
	children := m.Roots[0].Children

	// Both in cycle, fall back to ID sort: a before b.
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0].Bead.ID != "a" || children[1].Bead.ID != "b" {
		t.Errorf("expected [a, b] fallback order, got [%s, %s]", children[0].Bead.ID, children[1].Bead.ID)
	}
}

func TestSortNodes_MixedDepsAndNoDeps(t *testing.T) {
	// d depends on b (blocks). a, b, c are independent.
	// Expected: independent nodes first by priority+ID (a, b, c), then d.
	b := []data.Bead{
		{ID: "p", Priority: 0},
		{ID: "d", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "d", DependsOnID: "b", Type: "blocks"},
		}},
		{ID: "c", Parent: "p", Priority: 1},
		{ID: "a", Parent: "p", Priority: 1},
		{ID: "b", Parent: "p", Priority: 1},
	}

	m := BuildTree(b, false)
	children := m.Roots[0].Children

	expected := []string{"a", "b", "c", "d"}
	for i, exp := range expected {
		if children[i].Bead.ID != exp {
			t.Errorf("child[%d]: expected %s, got %s", i, exp, children[i].Bead.ID)
		}
	}
}

func TestSortNodes_CrossParentDepsIgnored(t *testing.T) {
	// task-2 (under epic-2) depends on task-1 (under epic-1).
	// This cross-parent dep should NOT affect sort within epic-2's children.
	b := []data.Bead{
		{ID: "epic-1", Priority: 1},
		{ID: "task-1", Parent: "epic-1", Priority: 1},
		{ID: "epic-2", Priority: 2},
		{ID: "task-2", Parent: "epic-2", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "task-2", DependsOnID: "task-1", Type: "blocks"},
		}},
		{ID: "task-3", Parent: "epic-2", Priority: 1},
	}

	m := BuildTree(b, false)
	epic2 := m.ByID["epic-2"]

	// task-1 is not a sibling of task-2, so dep is ignored.
	// Sort by ID: task-2 before task-3.
	if epic2.Children[0].Bead.ID != "task-2" {
		t.Errorf("expected task-2 first (cross-parent dep ignored), got %s", epic2.Children[0].Bead.ID)
	}
	if epic2.Children[1].Bead.ID != "task-3" {
		t.Errorf("expected task-3 second, got %s", epic2.Children[1].Bead.ID)
	}
}

func TestSortNodes_NonBlocksTypeIgnored(t *testing.T) {
	// b has a "relates_to" dependency on a — should be ignored for sort.
	b := []data.Bead{
		{ID: "p", Priority: 0},
		{ID: "b", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "b", DependsOnID: "a", Type: "relates_to"},
		}},
		{ID: "a", Parent: "p", Priority: 1},
	}

	m := BuildTree(b, false)
	children := m.Roots[0].Children

	// No blocks deps, so sort by ID: a, b.
	if children[0].Bead.ID != "a" || children[1].Bead.ID != "b" {
		t.Errorf("expected [a, b], got [%s, %s]", children[0].Bead.ID, children[1].Bead.ID)
	}
}

func TestSortNodes_PriorityBreaksTiesAmongIndependent(t *testing.T) {
	// c (priority 1) depends on both a and b. a has priority 2, b has priority 1.
	// Among independent nodes, b (priority 1) comes before a (priority 2).
	b := []data.Bead{
		{ID: "p", Priority: 0},
		{ID: "a", Parent: "p", Priority: 2},
		{ID: "b", Parent: "p", Priority: 1},
		{ID: "c", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "c", DependsOnID: "a", Type: "blocks"},
			{IssueID: "c", DependsOnID: "b", Type: "blocks"},
		}},
	}

	m := BuildTree(b, false)
	children := m.Roots[0].Children

	expected := []string{"b", "a", "c"}
	for i, exp := range expected {
		if children[i].Bead.ID != exp {
			t.Errorf("child[%d]: expected %s, got %s", i, exp, children[i].Bead.ID)
		}
	}
}

func TestSortNodes_RootLevelDeps(t *testing.T) {
	// Root-level nodes also get dependency-aware sort.
	b := []data.Bead{
		{ID: "z", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "z", DependsOnID: "a", Type: "blocks"},
		}},
		{ID: "a", Priority: 1},
	}

	m := BuildTree(b, false)

	if m.Roots[0].Bead.ID != "a" {
		t.Errorf("expected root 'a' first (blocks z), got %s", m.Roots[0].Bead.ID)
	}
	if m.Roots[1].Bead.ID != "z" {
		t.Errorf("expected root 'z' second, got %s", m.Roots[1].Bead.ID)
	}
}

func TestSortNodes_PartialCycle(t *testing.T) {
	// a and b form a cycle. c depends on a. d is independent.
	// Expected: d first (independent), then a, b (cycle fallback by ID), then c.
	b := []data.Bead{
		{ID: "p", Priority: 0},
		{ID: "c", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "c", DependsOnID: "a", Type: "blocks"},
		}},
		{ID: "b", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "b", DependsOnID: "a", Type: "blocks"},
		}},
		{ID: "a", Parent: "p", Priority: 1, Dependencies: []data.Dependency{
			{IssueID: "a", DependsOnID: "b", Type: "blocks"},
		}},
		{ID: "d", Parent: "p", Priority: 1},
	}

	m := BuildTree(b, false)
	children := m.Roots[0].Children

	// d is independent, comes first. a and b are in a cycle, appended by ID.
	// c depends on a (in cycle), also appended.
	if children[0].Bead.ID != "d" {
		t.Errorf("expected 'd' first (independent), got %s", children[0].Bead.ID)
	}
	// Remaining 3 (a, b, c) are all in/dependent on cycle — sorted by priority+ID.
	if len(children) != 4 {
		t.Fatalf("expected 4 children, got %d", len(children))
	}
	// a, b, c all have priority 1, so sorted by ID.
	if children[1].Bead.ID != "a" || children[2].Bead.ID != "b" || children[3].Bead.ID != "c" {
		t.Errorf("expected [a, b, c] for cycle+dependent fallback, got [%s, %s, %s]",
			children[1].Bead.ID, children[2].Bead.ID, children[3].Bead.ID)
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
