# VHS demo tape for bd-view
# Run: vhs demo.tape
# Requires: bd-view binary, a .beads database with sample data

Output demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Tokyo Night"

# Launch bd-view with sample database
Type "bd-view --db .beads/beads.db"
Enter
Sleep 2s

# Navigate down through the tree
Sleep 500ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Expand a node
Type "l"
Sleep 500ms
Type "j"
Sleep 400ms

# Switch to detail panel
Type@100ms "	"
Sleep 1s

# Scroll detail panel
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 500ms

# Switch back to tree
Type@100ms "	"
Sleep 500ms

# Expand all nodes
Type "e"
Sleep 1s

# Search for something
Type "/"
Sleep 300ms
Type@80ms "feature"
Sleep 1s
Escape
Sleep 500ms

# Open filter menu
Type "f"
Sleep 1s
Escape
Sleep 500ms

# Show help
Type "?"
Sleep 2s
Escape
Sleep 500ms

# Quit
Type "q"
Sleep 500ms
