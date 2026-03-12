package data

import (
	"context"
	"fmt"
	"testing"
)

// mockExecutor is a test stub for CommandExecutor.
type mockExecutor struct {
	// outputs maps a command key to its response.
	outputs map[string][]byte
	// errors maps a command key to an error.
	errors map[string]error
	// calls records all invocations for assertion.
	calls [][]string
}

func (m *mockExecutor) Execute(_ context.Context, args ...string) ([]byte, error) {
	m.calls = append(m.calls, args)
	key := fmt.Sprint(args)
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if out, ok := m.outputs[key]; ok {
		return out, nil
	}
	return nil, fmt.Errorf("unexpected call: %v", args)
}

func TestListAll_Success(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string][]byte{
			fmt.Sprint([]string{"list", "--all", "--json", "--limit", "0"}): []byte(`[
				{
					"id": "bead-1",
					"title": "First bead",
					"status": "open",
					"priority": 1,
					"issue_type": "task",
					"owner": "user@test.com",
					"created_at": "2026-03-12T09:00:00Z",
					"updated_at": "2026-03-12T09:00:00Z",
					"dependencies": [],
					"dependency_count": 0,
					"dependent_count": 0,
					"comment_count": 0
				},
				{
					"id": "bead-2",
					"title": "Second bead",
					"status": "closed",
					"priority": 2,
					"issue_type": "epic",
					"parent": "bead-1",
					"created_at": "2026-03-12T10:00:00Z",
					"updated_at": "2026-03-12T10:00:00Z",
					"dependencies": [
						{
							"issue_id": "bead-2",
							"depends_on_id": "bead-1",
							"type": "parent-child"
						}
					],
					"dependency_count": 1,
					"dependent_count": 0,
					"comment_count": 0
				}
			]`),
		},
	}

	fetcher := NewFetcher(mock)
	beads, err := fetcher.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(beads) != 2 {
		t.Fatalf("expected 2 beads, got %d", len(beads))
	}

	if beads[0].ID != "bead-1" {
		t.Errorf("expected id bead-1, got %s", beads[0].ID)
	}
	if beads[0].Title != "First bead" {
		t.Errorf("expected title 'First bead', got %s", beads[0].Title)
	}
	if beads[0].Status != "open" {
		t.Errorf("expected status open, got %s", beads[0].Status)
	}
	if beads[0].Priority != 1 {
		t.Errorf("expected priority 1, got %d", beads[0].Priority)
	}

	if beads[1].Parent != "bead-1" {
		t.Errorf("expected parent bead-1, got %s", beads[1].Parent)
	}
	if len(beads[1].Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(beads[1].Dependencies))
	}
	if beads[1].Dependencies[0].DependsOnID != "bead-1" {
		t.Errorf("expected depends_on_id bead-1, got %s", beads[1].Dependencies[0].DependsOnID)
	}
}

func TestListAll_EmptyList(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string][]byte{
			fmt.Sprint([]string{"list", "--all", "--json", "--limit", "0"}): []byte(`[]`),
		},
	}

	fetcher := NewFetcher(mock)
	beads, err := fetcher.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(beads) != 0 {
		t.Fatalf("expected 0 beads, got %d", len(beads))
	}
}

func TestListAll_MalformedJSON(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string][]byte{
			fmt.Sprint([]string{"list", "--all", "--json", "--limit", "0"}): []byte(`{not valid json`),
		},
	}

	fetcher := NewFetcher(mock)
	_, err := fetcher.ListAll(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestListAll_ExecutorError(t *testing.T) {
	mock := &mockExecutor{
		errors: map[string]error{
			fmt.Sprint([]string{"list", "--all", "--json", "--limit", "0"}): fmt.Errorf("bd not found"),
		},
	}

	fetcher := NewFetcher(mock)
	_, err := fetcher.ListAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestShow_Success(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string][]byte{
			fmt.Sprint([]string{"show", "bead-1", "--json"}): []byte(`[
				{
					"id": "bead-1",
					"title": "First bead",
					"description": "A detailed description",
					"design": "Some design notes",
					"acceptance_criteria": "- [ ] Criterion 1",
					"notes": "Extra notes",
					"status": "open",
					"priority": 1,
					"issue_type": "task",
					"owner": "user@test.com",
					"created_at": "2026-03-12T09:00:00Z",
					"updated_at": "2026-03-12T09:00:00Z",
					"dependencies": [
						{
							"id": "parent-1",
							"title": "Parent bead",
							"status": "open",
							"dependency_type": "parent-child"
						}
					],
					"dependents": [
						{
							"id": "child-1",
							"title": "Child bead",
							"status": "in_progress",
							"dependency_type": "blocks"
						}
					],
					"parent": "parent-1"
				}
			]`),
		},
	}

	fetcher := NewFetcher(mock)
	detail, err := fetcher.Show(context.Background(), "bead-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if detail.ID != "bead-1" {
		t.Errorf("expected id bead-1, got %s", detail.ID)
	}
	if detail.Description != "A detailed description" {
		t.Errorf("unexpected description: %s", detail.Description)
	}
	if detail.Design != "Some design notes" {
		t.Errorf("unexpected design: %s", detail.Design)
	}
	if detail.AcceptanceCriteria != "- [ ] Criterion 1" {
		t.Errorf("unexpected acceptance_criteria: %s", detail.AcceptanceCriteria)
	}
	if detail.Notes != "Extra notes" {
		t.Errorf("unexpected notes: %s", detail.Notes)
	}
	if len(detail.Dependents) != 1 {
		t.Fatalf("expected 1 dependent, got %d", len(detail.Dependents))
	}
	if detail.Dependents[0].ID != "child-1" {
		t.Errorf("expected dependent id child-1, got %s", detail.Dependents[0].ID)
	}
	if detail.Dependents[0].DependencyType != "blocks" {
		t.Errorf("expected dependency_type blocks, got %s", detail.Dependents[0].DependencyType)
	}
}

func TestShow_EmptyResult(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string][]byte{
			fmt.Sprint([]string{"show", "nonexistent", "--json"}): []byte(`[]`),
		},
	}

	fetcher := NewFetcher(mock)
	_, err := fetcher.Show(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for empty result, got nil")
	}
}

func TestShow_MalformedJSON(t *testing.T) {
	mock := &mockExecutor{
		outputs: map[string][]byte{
			fmt.Sprint([]string{"show", "bead-1", "--json"}): []byte(`not json`),
		},
	}

	fetcher := NewFetcher(mock)
	_, err := fetcher.Show(context.Background(), "bead-1")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}
