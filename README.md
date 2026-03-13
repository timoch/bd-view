# bd-view

Terminal UI viewer for `.beads` databases. Provides an interactive tree-based interface for browsing and inspecting beads issues.

## Prerequisites

- **[bd](https://github.com/timoch/bd)** CLI must be installed and available in `PATH`
- Terminal with Unicode support and 256-color capability

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

```bash
bd-view --db path/to/.beads/beads.db
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--db` | | Path to `.beads` database |
| `--refresh` | `2` | Refresh interval in seconds |
| `--expand-all` | `false` | Start with all tree nodes expanded |
| `--no-color` | `false` | Disable color output |
| `--type` | | Filter by bead type (repeatable) |
| `--status` | | Filter by bead status (repeatable) |
| `--version` | | Show version information |
