package ui

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/testutil"
)

func init() {
	noColorHighlight = true
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
	tm := testutil.NewTestModel(t, newIntegrationModel(), testutil.WithInitialTermSize(120, 40))

	// Wait for the tree to appear with our fixture beads
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1", "proj-1.1", "proj-1.2", "proj-2", "Beads")
	}, testutil.WithDuration(5*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, testutil.WithFinalTimeout(3*time.Second))
}

func TestIntegration_NavigationUpdatesDetail(t *testing.T) {
	tm := testutil.NewTestModel(t, newIntegrationModel(), testutil.WithInitialTermSize(120, 40))

	// Wait for tree to load
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1", "Beads")
	}, testutil.WithDuration(5*time.Second))

	// Navigate down to proj-1.1
	tm.Send(tea.KeyPressMsg{Code: 'j', Text: "j"})

	// Wait for detail pane to show proj-1.1 info
	// Note: bubbletea v2 uses differential rendering with cursor positioning,
	// so IDs like "proj-1.1" may be split across escape codes. Check for
	// contiguous strings that appear in single writes.
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "Set up CI pipeline")
	}, testutil.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, testutil.WithFinalTimeout(3*time.Second))
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

	tm := testutil.NewTestModel(t, m, testutil.WithInitialTermSize(120, 40))

	// Wait for tree to load — children should not be visible initially
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1", "▶") // collapsed indicator
	}, testutil.WithDuration(5*time.Second))

	// Expand proj-1 with Right arrow
	tm.Send(tea.KeyPressMsg{Code: tea.KeyRight})

	// Children should now be visible
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1.1", "proj-1.2", "proj-1.3", "▼")
	}, testutil.WithDuration(3*time.Second))

	// Collapse with Left arrow
	tm.Send(tea.KeyPressMsg{Code: tea.KeyLeft})

	// Children should be hidden again
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "▶") && !containsAll(s, "proj-1.1")
	}, testutil.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, testutil.WithFinalTimeout(3*time.Second))
}

func TestIntegration_Search(t *testing.T) {
	tm := testutil.NewTestModel(t, newIntegrationModel(), testutil.WithInitialTermSize(120, 40))

	// Wait for tree
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1", "proj-2")
	}, testutil.WithDuration(5*time.Second))

	// Enter search mode
	tm.Send(tea.KeyPressMsg{Code: '/', Text: "/"})

	// Type search query
	tm.Send(tea.KeyPressMsg{Code: 'a', Text: "auth"})

	// Should filter to only show proj-2 (Feature: User auth)
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-2", "Search:")
	}, testutil.WithDuration(3*time.Second))

	// Escape clears search
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEscape})

	// All beads should reappear
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1", "proj-2")
	}, testutil.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, testutil.WithFinalTimeout(3*time.Second))
}

func TestIntegration_Resize(t *testing.T) {
	tm := testutil.NewTestModel(t, newIntegrationModel(), testutil.WithInitialTermSize(120, 40))

	// Wait for tree
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return containsAll(string(bts), "proj-1")
	}, testutil.WithDuration(5*time.Second))

	// Resize to narrow mode
	tm.Send(tea.WindowSizeMsg{Width: 90, Height: 30})

	// In narrow mode the detail pane should be hidden — tree should still show
	testutil.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := string(bts)
		return containsAll(s, "proj-1", "Beads")
	}, testutil.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, testutil.WithFinalTimeout(3*time.Second))
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
