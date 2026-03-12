package data

// DiffResult describes what changed between two bead lists.
type DiffResult struct {
	Added   []Bead // beads in new but not old
	Removed []Bead // beads in old but not new
	Updated []Bead // beads in both but with different updated_at
}

// DiffBeads compares old and new bead slices, detecting added, removed, and updated beads.
// A bead is "updated" if its updated_at timestamp changed.
func DiffBeads(old, new []Bead) DiffResult {
	oldMap := make(map[string]Bead, len(old))
	for _, b := range old {
		oldMap[b.ID] = b
	}

	newMap := make(map[string]Bead, len(new))
	for _, b := range new {
		newMap[b.ID] = b
	}

	var result DiffResult

	// Find added and updated
	for _, b := range new {
		oldBead, exists := oldMap[b.ID]
		if !exists {
			result.Added = append(result.Added, b)
		} else {
			// Compare updated_at timestamps
			oldTime := oldBead.UpdatedAt
			newTime := b.UpdatedAt
			if oldTime == nil && newTime != nil {
				result.Updated = append(result.Updated, b)
			} else if oldTime != nil && newTime == nil {
				result.Updated = append(result.Updated, b)
			} else if oldTime != nil && newTime != nil && !oldTime.Equal(*newTime) {
				result.Updated = append(result.Updated, b)
			}
		}
	}

	// Find removed
	for _, b := range old {
		if _, exists := newMap[b.ID]; !exists {
			result.Removed = append(result.Removed, b)
		}
	}

	return result
}

// HasChanges returns true if any beads were added, removed, or updated.
func (d DiffResult) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Updated) > 0
}
