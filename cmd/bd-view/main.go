package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

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

			m := ui.New(cfg)
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
