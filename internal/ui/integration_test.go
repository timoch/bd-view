package ui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/testutil"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// newIntegrationModel creates a Model wired to a mock executor for integration testing.
func newIntegrationModel() Model {
	mock := &testutil.MockExecutor{
		Outputs: map[string][]byte{
			fmt.Sprint([]string{"list", "--all", "--json", "--limit", "0"}): testutil.SampleListJSON(),
		},
	}
	fetcher := data.NewFetcher(mock)

	m := New(Config{Refresh: 60, ExpandAll: true}) // long refresh to avoid extra fetches
	m.SetFetcher(fetcher)
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }
	return m
}

func TestIntegration_StartupShowsTree(t *testing.T) {
	tm := teatest.NewTestModel(t, newIntegrationModel(), teatest.WithInitialTermSize(120, 40))

	// Wait for the tree to appear with our fixture beads
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1", "proj-1.1", "proj-1.2", "proj-2", "Beads")
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestIntegration_NavigationUpdatesDetail(t *testing.T) {
	tm := teatest.NewTestModel(t, newIntegrationModel(), teatest.WithInitialTermSize(120, 40))

	// Wait for tree to load
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1", "Beads")
	}, teatest.WithDuration(5*time.Second))

	// Navigate down to proj-1.1
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Wait for detail pane to show proj-1.1 info
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1.1", "Set up CI pipeline")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestIntegration_ExpandCollapse(t *testing.T) {
	// Start with collapsed tree
	mock := &testutil.MockExecutor{
		Outputs: map[string][]byte{
			fmt.Sprint([]string{"list", "--all", "--json", "--limit", "0"}): testutil.SampleListJSON(),
		},
	}
	fetcher := data.NewFetcher(mock)
	m := New(Config{Refresh: 60, ExpandAll: false})
	m.SetFetcher(fetcher)
	m.nowFunc = func() time.Time { return time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) }

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 40))

	// Wait for tree to load — children should not be visible initially
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1", "[+]") // collapsed indicator
	}, teatest.WithDuration(5*time.Second))

	// Expand proj-1 with Right arrow
	tm.Send(tea.KeyMsg{Type: tea.KeyRight})

	// Children should now be visible
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1.1", "proj-1.2", "proj-1.3", "[-]")
	}, teatest.WithDuration(3*time.Second))

	// Collapse with Left arrow
	tm.Send(tea.KeyMsg{Type: tea.KeyLeft})

	// Children should be hidden again
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "[+]") && !containsAll(s, "proj-1.1")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestIntegration_Search(t *testing.T) {
	tm := teatest.NewTestModel(t, newIntegrationModel(), teatest.WithInitialTermSize(120, 40))

	// Wait for tree
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1", "proj-2")
	}, teatest.WithDuration(5*time.Second))

	// Enter search mode
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	// Type search query
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("auth")})

	// Should filter to only show proj-2 (Feature: User auth)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-2", "Search:")
	}, teatest.WithDuration(3*time.Second))

	// Escape clears search
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// All beads should reappear
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1", "proj-2")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestIntegration_Resize(t *testing.T) {
	tm := teatest.NewTestModel(t, newIntegrationModel(), teatest.WithInitialTermSize(120, 40))

	// Wait for tree
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1")
	}, teatest.WithDuration(5*time.Second))

	// Resize to narrow mode
	tm.Send(tea.WindowSizeMsg{Width: 90, Height: 30})

	// In narrow mode the detail pane should be hidden — tree should still show
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1", "Beads")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// containsAll checks if s contains all the given substrings.
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
