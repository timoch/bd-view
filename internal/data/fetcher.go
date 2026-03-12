package data

import (
	"context"
	"encoding/json"
	"fmt"
)

// Fetcher provides methods to fetch bead data via the bd CLI.
type Fetcher struct {
	Executor CommandExecutor
}

// NewFetcher creates a Fetcher with the given executor.
func NewFetcher(executor CommandExecutor) *Fetcher {
	return &Fetcher{Executor: executor}
}

// ListAll fetches all beads (open + closed) by running bd list --all --json --limit 0.
func (f *Fetcher) ListAll(ctx context.Context) ([]Bead, error) {
	out, err := f.Executor.Execute(ctx, "list", "--all", "--json", "--limit", "0")
	if err != nil {
		return nil, fmt.Errorf("list beads: %w", err)
	}

	var beads []Bead
	if err := json.Unmarshal(out, &beads); err != nil {
		return nil, fmt.Errorf("parse bead list: %w", err)
	}
	return beads, nil
}

// Show fetches full detail for a single bead by running bd show <id> --json.
// The bd show command returns a JSON array; this returns the first element.
func (f *Fetcher) Show(ctx context.Context, id string) (*BeadDetail, error) {
	out, err := f.Executor.Execute(ctx, "show", id, "--json")
	if err != nil {
		return nil, fmt.Errorf("show bead %s: %w", id, err)
	}

	var beads []BeadDetail
	if err := json.Unmarshal(out, &beads); err != nil {
		return nil, fmt.Errorf("parse bead detail: %w", err)
	}
	if len(beads) == 0 {
		return nil, fmt.Errorf("bead %s not found", id)
	}
	return &beads[0], nil
}
