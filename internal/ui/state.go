package ui

import (
	"encoding/json"
	"os"
	"sort"
)

// expandStateFile represents the persisted expand/collapse state.
type expandStateFile struct {
	ExpandedNodes []string `json:"expanded_nodes"`
}

// LoadExpandState reads the expand state from the given file path.
// Returns nil if the file doesn't exist, is unreadable, or contains invalid JSON.
func LoadExpandState(path string) []string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var state expandStateFile
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil
	}
	return state.ExpandedNodes
}

// SaveExpandState writes the expand state to the given file path.
// Errors are silently ignored to avoid blocking the UI.
func SaveExpandState(path string, expandedIDs []string) {
	sort.Strings(expandedIDs)
	state := expandStateFile{ExpandedNodes: expandedIDs}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, raw, 0644)
}

// collectExpandedIDs returns the IDs of all currently expanded nodes.
func (m *Model) collectExpandedIDs() []string {
	if m.tree == nil {
		return nil
	}
	ids := make([]string, 0)
	for id, node := range m.tree.ByID {
		if node.Expanded {
			ids = append(ids, id)
		}
	}
	return ids
}

// persistExpandState saves the current expand state to disk asynchronously.
func (m *Model) persistExpandState() {
	if m.config.StatePath == "" || m.tree == nil {
		return
	}
	ids := m.collectExpandedIDs()
	path := m.config.StatePath
	go SaveExpandState(path, ids)
}
