package ui

// Layout constants centralise all margin, padding, and border values used
// across the UI subsystems.  Every panel/overlay derives its content width
// from these constants so that wrapping, selection-coordinate mapping, and
// rendering stay in sync.

const (
	// Tree panel
	treeBorderRight = 1 // right border separating tree from detail

	// Detail panel
	detailPaddingLeft = 1

	// Overlay panels (help, filter)
	overlayPaddingLeft = 2

	// Glamour (markdown) renderer – document margin (left only; right is 0).
	glamourMarginLeft = 0

	// Terminal size thresholds
	minTermWidth  = 80
	minTermHeight = 24
	narrowWidth   = 100 // below this, single-panel mode

	// Tree sizing
	treeWidthRatio = 2  // numerator: tree gets treeWidthRatio/treeWidthDiv of width
	treeWidthDiv   = 5  // denominator
	minTreeWidth   = 20 // floor for tree panel width
)

// detailContentWidth returns the usable content width inside the detail panel
// (total panel width minus its left padding).
func detailContentWidth(panelWidth int) int {
	return panelWidth - detailPaddingLeft
}

// overlayContentWidth returns the usable content width inside an overlay panel.
func overlayContentWidth(panelWidth int) int {
	return panelWidth - overlayPaddingLeft
}
