# VHS demo tape for bd-view
# Run: vhs demo.tape
# Requires: bd-view binary in PATH, a .beads database with data, bd CLI in PATH

Output "demo.gif"

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "TokyoNight"

# Ensure bd and bd-view are in PATH
Type "export PATH=$HOME/.local/bin:$HOME/go/bin:$PATH"
Enter
Sleep 500ms

# Launch bd-view with the project database
Type "bd-view"
Enter
Sleep 3s

# Navigate down through the tree
Sleep 500ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Expand a node
Right
Sleep 500ms
Type "j"
Sleep 400ms

# Switch to detail panel
Tab
Sleep 1s

# Scroll detail panel
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 500ms

# Switch back to tree
Tab
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
