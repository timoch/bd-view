package ui

import "github.com/charmbracelet/lipgloss"

// Theme colors aligned with ralph-tui's Tokyo Night palette.
// See: ralph-tui src/tui/theme.ts (default theme)

// Status colors — mapped from ralph-tui task states
var (
	colorStatusOpen       = lipgloss.Color("#565f89") // pending (muted)
	colorStatusInProgress = lipgloss.Color("#9ece6a") // active (green)
	colorStatusBlocked    = lipgloss.Color("#f7768e") // blocked/error (red)
	colorStatusDeferred   = lipgloss.Color("#565f89") // pending (muted), rendered faint
	colorStatusClosed     = lipgloss.Color("#414868") // closed (dim)
)

// Accent / focus colors — mapped from ralph-tui accent + border
var (
	colorAccentPrimary = lipgloss.Color("#7aa2f7") // active border, focus highlight
	colorBorderNormal  = lipgloss.Color("#3d4259") // inactive border
)

// Status bar colors — mapped from ralph-tui bg/fg
var (
	colorStatusBarBg = lipgloss.Color("#2f3449") // bg tertiary
	colorStatusBarFg = lipgloss.Color("#a9b1d6") // fg secondary
)

// Progress indicator colors
var (
	colorProgressDone    = lipgloss.Color("#9ece6a") // success green
	colorProgressPartial = lipgloss.Color("#e0af68") // warning yellow
)
