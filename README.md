# bd-view

A terminal UI for browsing `.beads` databases.

<!-- To regenerate the demo GIF: vhs demo.tape -->
![bd-view demo](demo.gif)

## Features

- **Tree navigation** -- hierarchical parent-child view with expand/collapse
- **Split-panel layout** -- tree on the left, detail on the right (auto-stacks in narrow terminals)
- **Rich detail view** -- markdown-rendered description, design, acceptance criteria, and notes
- **Search** -- fuzzy find by ID, title, or content with match highlighting
- **Filter** -- narrow by bead type and status via interactive menu
- **Real-time refresh** -- polls for changes with configurable interval, preserves UI state
- **Clipboard** -- copy bead IDs or select and copy text from the detail panel
- **Read-only** -- safe to use on any database, never writes
- **Vim keybindings** -- `j`/`k`, `g`/`G`, `/` search, and more
- **Mouse support** -- click to focus, scroll, drag to select text

## Installation

### Download pre-built binary

Download the latest release for your platform from [GitHub Releases](https://github.com/timoch/bd-view/releases/latest).

```bash
# Linux (amd64)
curl -Lo bd-view.tar.gz https://github.com/timoch/bd-view/releases/latest/download/bd-view_Linux_x86_64.tar.gz

# Linux (arm64)
curl -Lo bd-view.tar.gz https://github.com/timoch/bd-view/releases/latest/download/bd-view_Linux_arm64.tar.gz

# macOS (Apple Silicon)
curl -Lo bd-view.tar.gz https://github.com/timoch/bd-view/releases/latest/download/bd-view_Darwin_arm64.tar.gz

# macOS (Intel)
curl -Lo bd-view.tar.gz https://github.com/timoch/bd-view/releases/latest/download/bd-view_Darwin_x86_64.tar.gz
```

Extract and install to `~/.local/bin/` (no sudo required):

```bash
tar xzf bd-view.tar.gz
mv bd-view ~/.local/bin/
```

For system-wide install:

```bash
sudo mv bd-view /usr/local/bin/
```

#### Verify checksums

Each release includes a `checksums.txt` file for verification:

```bash
curl -Lo checksums.txt https://github.com/timoch/bd-view/releases/latest/download/checksums.txt
sha256sum -c checksums.txt --ignore-missing
```

### Go install

```bash
go install github.com/timoch/bd-view/cmd/bd-view@latest
```

### Build from source

```bash
git clone https://github.com/timoch/bd-view.git
cd bd-view
make build
# or: go build -o bd-view ./cmd/bd-view
```

Install to `~/.local/bin/` (default prefix):

```bash
make install
```

To install elsewhere:

```bash
make install PREFIX=/usr/local
```

## Usage

Run from a directory containing a `.beads` directory (or any parent):

```bash
bd-view
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--refresh` | `2` | Refresh interval in seconds |
| `--expand-all` | `false` | Start with all tree nodes expanded |
| `--no-color` | `false` | Disable color output |
| `--type` | | Filter by bead type (repeatable) |
| `--status` | | Filter by bead status (repeatable) |
| `--version` | | Show version information |

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `Down` | Move selection down |
| `k` / `Up` | Move selection up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Tab` | Switch focus (tree / detail) |
| `Click` | Switch focus to clicked panel |

### Tree

| Key | Action |
|-----|--------|
| `Enter` | Expand node / open overlay (narrow mode) |
| `Right` | Expand node |
| `Left` | Collapse node / go to parent |
| `e` | Expand all nodes |
| `c` | Collapse all nodes |

### Search & Filter

| Key | Action |
|-----|--------|
| `/` | Search by ID, title, or content |
| `f` | Open filter menu |
| `Esc` | Clear search/filter, close overlay |

### Other

| Key | Action |
|-----|--------|
| `y` | Copy bead ID to clipboard |
| `Drag` | Select text in detail panel |
| `Right-click` | Copy selection and clear |
| `Shift+Click` | Terminal-native text selection |
| `r` | Force refresh |
| `?` | Show help overlay |
| `q` / `Ctrl+C` | Quit |

## Requirements

- **Go 1.25+** (for building from source)
- **[bd](https://github.com/timoch/bd)** CLI installed and in `PATH`
- Terminal with Unicode support and 256-color capability (minimum 80x24)

## License

[MIT](LICENSE)
