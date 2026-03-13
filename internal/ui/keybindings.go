package ui

// Keybinding defines a single keybinding entry for the help overlay.
type Keybinding struct {
	Keys        string // display string: "j / Down", "Tab", "y"
	Description string // "Move selection down", "Copy bead ID"
	Section     string // "NAVIGATION", "TREE", "SEARCH & FILTER", "OTHER"
}

// keybindingRegistry is the single source of truth for all keybindings.
// The help overlay is generated from this registry.
var keybindingRegistry = []Keybinding{
	// Navigation
	{Keys: "j / Down", Description: "Move selection down", Section: "NAVIGATION"},
	{Keys: "k / Up", Description: "Move selection up", Section: "NAVIGATION"},
	{Keys: "g", Description: "Go to top", Section: "NAVIGATION"},
	{Keys: "G", Description: "Go to bottom", Section: "NAVIGATION"},
	{Keys: "Tab", Description: "Switch focus (tree / detail)", Section: "NAVIGATION"},
	{Keys: "Click", Description: "Switch focus to clicked panel", Section: "NAVIGATION"},

	// Tree
	{Keys: "Enter", Description: "Expand node / open overlay (narrow)", Section: "TREE"},
	{Keys: "Right", Description: "Expand node", Section: "TREE"},
	{Keys: "Left", Description: "Collapse node / go to parent", Section: "TREE"},
	{Keys: "e", Description: "Expand all nodes", Section: "TREE"},
	{Keys: "c", Description: "Collapse all nodes", Section: "TREE"},

	// Search & Filter
	{Keys: "/", Description: "Search by ID, title, or content", Section: "SEARCH & FILTER"},
	{Keys: "f", Description: "Open filter menu", Section: "SEARCH & FILTER"},
	{Keys: "Esc", Description: "Clear search/filter, close overlay", Section: "SEARCH & FILTER"},

	// Other
	{Keys: "y", Description: "Copy bead ID to clipboard", Section: "OTHER"},
	{Keys: "Drag", Description: "Select text in detail panel", Section: "OTHER"},
	{Keys: "Right-click", Description: "Copy selection and clear", Section: "OTHER"},
	{Keys: "Shift+Click", Description: "Terminal-native text selection (passthrough)", Section: "OTHER"},
	{Keys: "r", Description: "Force refresh", Section: "OTHER"},
	{Keys: "?", Description: "Show this help", Section: "OTHER"},
	{Keys: "q", Description: "Quit", Section: "OTHER"},
}

// sectionOrder defines the display order of sections in the help overlay.
var sectionOrder = []string{"NAVIGATION", "TREE", "SEARCH & FILTER", "OTHER"}
