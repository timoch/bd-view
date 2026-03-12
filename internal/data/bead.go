package data

import "time"

// Bead represents a single bead from the bd CLI JSON output.
type Bead struct {
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	Description        string       `json:"description"`
	Design             string       `json:"design"`
	AcceptanceCriteria string       `json:"acceptance_criteria"`
	Notes              string       `json:"notes"`
	Status             string       `json:"status"`
	Priority           int          `json:"priority"`
	IssueType          string       `json:"issue_type"`
	Owner              string       `json:"owner"`
	Parent             string       `json:"parent"`
	Dependencies       []Dependency `json:"dependencies"`
	Labels             []string     `json:"labels"`
	CreatedAt          *time.Time   `json:"created_at"`
	CreatedBy          string       `json:"created_by"`
	UpdatedAt          *time.Time   `json:"updated_at"`
	ClosedAt           *time.Time   `json:"closed_at"`
	CloseReason        string       `json:"close_reason"`
	ExternalRef        string       `json:"external_ref"`
	DependencyCount    int          `json:"dependency_count"`
	DependentCount     int          `json:"dependent_count"`
	CommentCount       int          `json:"comment_count"`
}

// Dependency represents a dependency relationship from bd list output.
type Dependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	Metadata    string `json:"metadata"`
}

// BeadDetail is the richer representation from bd show, which includes
// full dependency/dependent bead info.
type BeadDetail struct {
	Bead
	Dependents []RelatedBead `json:"dependents"`
}

// RelatedBead is a bead reference returned in the dependencies/dependents
// arrays of bd show output.
type RelatedBead struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       int        `json:"priority"`
	IssueType      string     `json:"issue_type"`
	Owner          string     `json:"owner"`
	CreatedAt      *time.Time `json:"created_at"`
	CreatedBy      string     `json:"created_by"`
	UpdatedAt      *time.Time `json:"updated_at"`
	ExternalRef    string     `json:"external_ref"`
	DependencyType string     `json:"dependency_type"`
}
