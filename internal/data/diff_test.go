package data

import (
	"testing"
	"time"
)

func TestDiffBeads_NoChanges(t *testing.T) {
	now := time.Now()
	beads := []Bead{
		{ID: "a", UpdatedAt: &now},
		{ID: "b", UpdatedAt: &now},
	}
	result := DiffBeads(beads, beads)
	if result.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestDiffBeads_Added(t *testing.T) {
	now := time.Now()
	old := []Bead{{ID: "a", UpdatedAt: &now}}
	new := []Bead{{ID: "a", UpdatedAt: &now}, {ID: "b", UpdatedAt: &now}}
	result := DiffBeads(old, new)
	if len(result.Added) != 1 || result.Added[0].ID != "b" {
		t.Errorf("expected 1 added bead 'b', got %v", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(result.Removed))
	}
	if len(result.Updated) != 0 {
		t.Errorf("expected 0 updated, got %d", len(result.Updated))
	}
}

func TestDiffBeads_Removed(t *testing.T) {
	now := time.Now()
	old := []Bead{{ID: "a", UpdatedAt: &now}, {ID: "b", UpdatedAt: &now}}
	new := []Bead{{ID: "a", UpdatedAt: &now}}
	result := DiffBeads(old, new)
	if len(result.Removed) != 1 || result.Removed[0].ID != "b" {
		t.Errorf("expected 1 removed bead 'b', got %v", result.Removed)
	}
	if len(result.Added) != 0 {
		t.Errorf("expected 0 added, got %d", len(result.Added))
	}
}

func TestDiffBeads_Updated(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Second)
	old := []Bead{{ID: "a", UpdatedAt: &t1}}
	new := []Bead{{ID: "a", UpdatedAt: &t2}}
	result := DiffBeads(old, new)
	if len(result.Updated) != 1 || result.Updated[0].ID != "a" {
		t.Errorf("expected 1 updated bead 'a', got %v", result.Updated)
	}
	if len(result.Added) != 0 || len(result.Removed) != 0 {
		t.Error("expected no added or removed")
	}
}

func TestDiffBeads_SameTimestamp(t *testing.T) {
	now := time.Now()
	old := []Bead{{ID: "a", UpdatedAt: &now}}
	new := []Bead{{ID: "a", UpdatedAt: &now}}
	result := DiffBeads(old, new)
	if result.HasChanges() {
		t.Error("expected no changes when timestamps are equal")
	}
}

func TestDiffBeads_NilTimestamps(t *testing.T) {
	old := []Bead{{ID: "a", UpdatedAt: nil}}
	new := []Bead{{ID: "a", UpdatedAt: nil}}
	result := DiffBeads(old, new)
	if result.HasChanges() {
		t.Error("expected no changes when both timestamps are nil")
	}
}

func TestDiffBeads_NilToNonNil(t *testing.T) {
	now := time.Now()
	old := []Bead{{ID: "a", UpdatedAt: nil}}
	new := []Bead{{ID: "a", UpdatedAt: &now}}
	result := DiffBeads(old, new)
	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated when nil->non-nil, got %d", len(result.Updated))
	}
}

func TestDiffBeads_Mixed(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Second)
	old := []Bead{
		{ID: "a", UpdatedAt: &t1},
		{ID: "b", UpdatedAt: &t1},
		{ID: "c", UpdatedAt: &t1},
	}
	new := []Bead{
		{ID: "a", UpdatedAt: &t1}, // unchanged
		{ID: "b", UpdatedAt: &t2}, // updated
		{ID: "d", UpdatedAt: &t1}, // added
		// c removed
	}
	result := DiffBeads(old, new)
	if len(result.Added) != 1 || result.Added[0].ID != "d" {
		t.Errorf("expected added=[d], got %v", result.Added)
	}
	if len(result.Removed) != 1 || result.Removed[0].ID != "c" {
		t.Errorf("expected removed=[c], got %v", result.Removed)
	}
	if len(result.Updated) != 1 || result.Updated[0].ID != "b" {
		t.Errorf("expected updated=[b], got %v", result.Updated)
	}
}

func TestDiffBeads_EmptyOld(t *testing.T) {
	now := time.Now()
	new := []Bead{{ID: "a", UpdatedAt: &now}}
	result := DiffBeads(nil, new)
	if len(result.Added) != 1 {
		t.Errorf("expected 1 added from empty old, got %d", len(result.Added))
	}
}

func TestDiffBeads_EmptyNew(t *testing.T) {
	now := time.Now()
	old := []Bead{{ID: "a", UpdatedAt: &now}}
	result := DiffBeads(old, nil)
	if len(result.Removed) != 1 {
		t.Errorf("expected 1 removed when new is empty, got %d", len(result.Removed))
	}
}

func TestDiffBeads_BothEmpty(t *testing.T) {
	result := DiffBeads(nil, nil)
	if result.HasChanges() {
		t.Error("expected no changes when both empty")
	}
}
