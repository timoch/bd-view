// Package testutil provides shared test helpers for bd-view tests.
package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/timoch/bd-view/internal/data"
)

// MockExecutor is a test stub for data.CommandExecutor.
type MockExecutor struct {
	// Outputs maps a command key (fmt.Sprint(args)) to its response.
	Outputs map[string][]byte
	// Errors maps a command key to an error.
	Errors map[string]error
	// Calls records all invocations for assertion.
	Calls [][]string
}

func (m *MockExecutor) Execute(_ context.Context, args ...string) ([]byte, error) {
	m.Calls = append(m.Calls, args)
	key := fmt.Sprint(args)
	if err, ok := m.Errors[key]; ok {
		return nil, err
	}
	if out, ok := m.Outputs[key]; ok {
		return out, nil
	}
	return nil, fmt.Errorf("unexpected call: %v", args)
}

// SampleBeads returns a set of beads useful for view snapshot testing.
// It includes an epic parent with 3 task children, covering various statuses.
func SampleBeads() []data.Bead {
	t1 := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)
	return []data.Bead{
		{
			ID:        "proj-1",
			Title:     "Epic: Project setup",
			IssueType: "epic",
			Status:    "in_progress",
			Priority:  0,
			Owner:     "dev@test.com",
			CreatedAt: &t1,
			UpdatedAt: &t2,
		},
		{
			ID:        "proj-1.1",
			Title:     "Set up CI pipeline",
			IssueType: "task",
			Status:    "closed",
			Priority:  1,
			Owner:     "dev@test.com",
			Parent:    "proj-1",
			CreatedAt: &t1,
			UpdatedAt: &t2,
			ClosedAt:  &t2,
		},
		{
			ID:        "proj-1.2",
			Title:     "Configure linting",
			IssueType: "task",
			Status:    "open",
			Priority:  2,
			Owner:     "dev@test.com",
			Parent:    "proj-1",
			CreatedAt: &t1,
			UpdatedAt: &t1,
		},
		{
			ID:        "proj-1.3",
			Title:     "Write integration tests",
			IssueType: "task",
			Status:    "blocked",
			Priority:  3,
			Parent:    "proj-1",
			CreatedAt: &t1,
			UpdatedAt: &t1,
		},
		{
			ID:        "proj-2",
			Title:     "Feature: User auth",
			IssueType: "feature",
			Status:    "open",
			Priority:  1,
			Owner:     "dev@test.com",
			CreatedAt: &t1,
			UpdatedAt: &t1,
		},
	}
}

// SampleBeadDetail returns a BeadDetail with all fields populated for detail pane testing.
func SampleBeadDetail() *data.BeadDetail {
	t1 := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)
	return &data.BeadDetail{
		Bead: data.Bead{
			ID:                 "proj-1.1",
			Title:              "Set up CI pipeline",
			Description:        "Configure GitHub Actions for CI.\n\n**Requirements:**\n- Build on push\n- Run tests\n- Lint check",
			Design:             "Use `.github/workflows/ci.yml`.\n```yaml\nname: CI\non: push\n```",
			AcceptanceCriteria: "- [ ] CI runs on push\n- [ ] Tests pass\n- [ ] Lint passes",
			Notes:              "See docs/ci.md for reference.",
			IssueType:          "task",
			Status:             "closed",
			Priority:           1,
			Owner:              "dev@test.com",
			Parent:             "proj-1",
			CreatedAt:          &t1,
			UpdatedAt:          &t2,
			ClosedAt:           &t2,
			Dependencies: []data.Dependency{
				{IssueID: "proj-1.1", DependsOnID: "proj-1", Type: "parent-child"},
			},
		},
		Dependents: []data.RelatedBead{
			{ID: "proj-1.2", Title: "Configure linting", Status: "open", DependencyType: "blocks"},
		},
	}
}

// SampleListJSON returns JSON output matching SampleBeads(), as bd list --all --json would produce.
func SampleListJSON() []byte {
	return []byte(`[
	{"id":"proj-1","title":"Epic: Project setup","issue_type":"epic","status":"in_progress","priority":0,"owner":"dev@test.com","created_at":"2026-03-10T00:00:00Z","updated_at":"2026-03-11T00:00:00Z","dependencies":[],"dependency_count":0,"dependent_count":0,"comment_count":0},
	{"id":"proj-1.1","title":"Set up CI pipeline","issue_type":"task","status":"closed","priority":1,"owner":"dev@test.com","parent":"proj-1","created_at":"2026-03-10T00:00:00Z","updated_at":"2026-03-11T00:00:00Z","closed_at":"2026-03-11T00:00:00Z","dependencies":[],"dependency_count":0,"dependent_count":0,"comment_count":0},
	{"id":"proj-1.2","title":"Configure linting","issue_type":"task","status":"open","priority":2,"owner":"dev@test.com","parent":"proj-1","created_at":"2026-03-10T00:00:00Z","updated_at":"2026-03-10T00:00:00Z","dependencies":[],"dependency_count":0,"dependent_count":0,"comment_count":0},
	{"id":"proj-1.3","title":"Write integration tests","issue_type":"task","status":"blocked","priority":3,"parent":"proj-1","created_at":"2026-03-10T00:00:00Z","updated_at":"2026-03-10T00:00:00Z","dependencies":[],"dependency_count":0,"dependent_count":0,"comment_count":0},
	{"id":"proj-2","title":"Feature: User auth","issue_type":"feature","status":"open","priority":1,"owner":"dev@test.com","created_at":"2026-03-10T00:00:00Z","updated_at":"2026-03-10T00:00:00Z","dependencies":[],"dependency_count":0,"dependent_count":0,"comment_count":0}
]`)
}
