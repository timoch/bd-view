package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/cobra"

	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/ui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var cfg ui.Config

	rootCmd := &cobra.Command{
		Use:   "bd-view",
		Short: "Terminal UI viewer for .beads databases",
		Long:  "Run in a directory containing a .beads directory (or any parent).",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Derive state file path from .beads directory
			if beadsDir, err := findBeadsDir(); err == nil {
				cfg.StatePath = filepath.Join(beadsDir, "bd-view-state.json")
			}
			cfg.ExpandAllExplicit = cmd.Flags().Changed("expand-all")

			executor := &data.BdExecutor{}
			fetcher := data.NewFetcher(executor)

			m := ui.New(cfg)
			m.SetFetcher(fetcher)

			opts := []tea.ProgramOption{}
			if cfg.NoColor {
				opts = append(opts, tea.WithColorProfile(colorprofile.Ascii))
			}

			p := tea.NewProgram(m, opts...)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running TUI: %w", err)
			}
			return nil
		},
	}

	rootCmd.Flags().IntVar(&cfg.Refresh, "refresh", 2, "Refresh interval in seconds")
	rootCmd.Flags().BoolVar(&cfg.ExpandAll, "expand-all", false, "Start with all tree nodes expanded")
	rootCmd.Flags().BoolVar(&cfg.NoColor, "no-color", false, "Disable color output")
	rootCmd.Flags().StringSliceVar(&cfg.FilterTypes, "type", nil, "Filter by bead type (can specify multiple)")
	rootCmd.Flags().StringSliceVar(&cfg.FilterStatuses, "status", nil, "Filter by bead status (can specify multiple)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// findBeadsDir walks up from the current directory looking for a .beads directory.
func findBeadsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, ".beads")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf(".beads directory not found")
		}
		dir = parent
	}
}
