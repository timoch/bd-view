package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/timoch/bd-view/internal/data"
	"github.com/timoch/bd-view/internal/ui"
)

func main() {
	var cfg ui.Config

	rootCmd := &cobra.Command{
		Use:   "bd-view",
		Short: "Terminal UI viewer for .beads databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.NoColor {
				lipgloss.SetColorProfile(termenv.Ascii)
			}

			// Derive state file path from database path
			if cfg.DBPath != "" {
				cfg.StatePath = filepath.Join(filepath.Dir(cfg.DBPath), "bd-view-state.json")
			}
			cfg.ExpandAllExplicit = cmd.Flags().Changed("expand-all")

			executor := &data.BdExecutor{DBPath: cfg.DBPath}
			fetcher := data.NewFetcher(executor)

			m := ui.New(cfg)
			m.SetFetcher(fetcher)
			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running TUI: %w", err)
			}
			return nil
		},
	}

	rootCmd.Flags().StringVar(&cfg.DBPath, "db", "", "Path to .beads database")
	rootCmd.Flags().IntVar(&cfg.Refresh, "refresh", 2, "Refresh interval in seconds")
	rootCmd.Flags().BoolVar(&cfg.ExpandAll, "expand-all", false, "Start with all tree nodes expanded")
	rootCmd.Flags().BoolVar(&cfg.NoColor, "no-color", false, "Disable color output")
	rootCmd.Flags().StringSliceVar(&cfg.FilterTypes, "type", nil, "Filter by bead type (can specify multiple)")
	rootCmd.Flags().StringSliceVar(&cfg.FilterStatuses, "status", nil, "Filter by bead status (can specify multiple)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
