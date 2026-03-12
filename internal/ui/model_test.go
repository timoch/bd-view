package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/data"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func TestDetailPanel_EmptyState(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	output := m.View()
	if !strings.Contains(output, "Select a bead to view details") {
		t.Error("expected empty state message when no bead selected")
	}
}

func TestDetailPanel_ShowsBeadHeader(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	created := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	closed := time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)
	bead := &data.Bead{
		ID:        "hep-ws-f6.3",
		Title:     "US-003: Install SignalR client",
		IssueType: "task",
		Status:    "closed",
		Priority:  1,
		Owner:     "timoch@timoch.com",
		Parent:    "hep-ws-f6",
		CreatedAt: &created,
		ClosedAt:  &closed,
	}
	m.SetSelectedBead(bead)

	output := m.View()

	checks := []struct {
		label string
		want  string
	}{
		{"bead ID", "hep-ws-f6.3"},
		{"title", "US-003: Install SignalR client"},
		{"type", "task"},
		{"status", "closed"},
		{"priority", "1"},
		{"owner", "timoch@timoch.com"},
		{"parent", "hep-ws-f6"},
		{"created date", "2026-03-10"},
		{"closed date", "2026-03-11"},
	}
	for _, c := range checks {
		if !strings.Contains(output, c.want) {
			t.Errorf("expected %s %q in output", c.label, c.want)
		}
	}
}

func TestDetailPanel_HidesEmptyDates(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Test bead",
		IssueType: "task",
		Status:    "open",
	}
	m.SetSelectedBead(bead)

	output := m.View()
	if strings.Contains(output, "Created:") {
		t.Error("should not show Created when CreatedAt is nil")
	}
	if strings.Contains(output, "Closed:") {
		t.Error("should not show Closed when ClosedAt is nil")
	}
}

func TestDetailPanel_HidesParentWhenEmpty(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Root bead",
		IssueType: "epic",
		Status:    "open",
	}
	m.SetSelectedBead(bead)

	output := m.View()
	if strings.Contains(output, "Parent:") {
		t.Error("should not show Parent when parent is empty")
	}
}

func TestDetailPanel_HasSeparators(t *testing.T) {
	m := New(Config{Refresh: 2})
	m.width = 120
	m.height = 40
	m.ready = true

	bead := &data.Bead{
		ID:        "test-1",
		Title:     "Test",
		IssueType: "task",
		Status:    "open",
	}
	m.SetSelectedBead(bead)

	output := m.View()
	// Should have horizontal separators (─ characters)
	if !strings.Contains(output, "─") {
		t.Error("expected horizontal separator lines")
	}
}

func TestColorStatus_AllStatuses(t *testing.T) {
	m := Model{}
	statuses := []string{"open", "in_progress", "blocked", "deferred", "closed", "unknown"}
	for _, s := range statuses {
		result := m.colorStatus(s)
		if !strings.Contains(result, s) {
			t.Errorf("colorStatus(%q) = %q, should contain status text", s, result)
		}
	}
}
