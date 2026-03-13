# PRD: Beads TUI Viewer

## Overview

A terminal user interface (TUI) application for browsing a `.beads` database in real time. The viewer presents a navigable tree of beads on the left panel with type/status indicators, and a scrollable detail pane on the right showing all fields of the selected bead. The display refreshes automatically as the underlying database changes, giving the user a live dashboard while they edit beads via `bd` or other tools in a separate terminal.

## Problem Statement

Working with beads today requires running individual `bd` commands (`bd list`, `bd show`, `bd graph`) one at a time. There is no persistent, unified view that lets you see the full tree structure and drill into details simultaneously. When editing beads in another terminal, you have to re-run commands to see the effect. A real-time TUI solves this by providing a always-up-to-date, interactive view.

## Goals

1. **Real-time tree view** -- Display all beads organized by parent-child hierarchy, updating live as the database changes.
2. **Rich detail pane** -- Show every field of the selected bead (description, design, acceptance criteria, notes, dependencies, metadata) with proper formatting.
3. **Fast navigation** -- Vim-style keybindings and search/filter for quick traversal of large backlogs.
4. **Zero-write guarantee** -- The viewer is strictly read-only. It never modifies the beads database.

## Non-Goals

- Editing beads from within the TUI (use `bd` CLI directly).
- Replacing the `bd` CLI -- this is a complementary viewer.
- Supporting multiple `.beads` databases simultaneously.
- Remote/networked access to beads databases.

## Data Source

All data is read via the `bd` CLI with `--json` output. The TUI does **not** access the Dolt database or SQLite files directly. This keeps the viewer decoupled from storage internals and compatible with future backend changes.

### Key Commands Used

| Command | Purpose |
|---|---|
| `bd list --all --json --limit 0` | Fetch all beads (open + closed) with full metadata |
| `bd show <id> --json` | Fetch full detail for a single bead (on selection) |
| `bd children <id> --json` | Fetch children of a parent bead |
| `bd types` | Discover available bead types |
| `bd status` | Database health/overview |

### Bead Data Model (from JSON output)

Each bead exposes:

| Field | Type | Notes |
|---|---|---|
| `id` | string | e.g. `hep-ws-f6.3` |
| `title` | string | Short summary |
| `description` | string | Full description (may be multi-paragraph markdown) |
| `design` | string | Implementation design notes |
| `acceptance_criteria` | string | Checklist-style criteria |
| `notes` | string | Additional context |
| `status` | enum | `open`, `in_progress`, `blocked`, `deferred`, `closed` |
| `priority` | int | 0-4 |
| `issue_type` | enum | `task`, `bug`, `feature`, `chore`, `epic`, `decision` |
| `owner` | string | |
| `parent` | string | Parent bead ID (null for top-level) |
| `dependencies` | array | Objects with `depends_on_id` and `type` |
| `labels` | array | String labels |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |
| `closed_at` | timestamp | |
| `close_reason` | string | |

### Refresh Strategy

The TUI polls `bd list --all --json --limit 0` at a configurable interval (default: 2 seconds). On detecting changes (comparing bead IDs + `updated_at` timestamps), it updates the tree and, if the currently-selected bead changed, refreshes the detail pane via `bd show <id> --json`.

An alternative approach: use filesystem watching on the `.beads/` directory to trigger refreshes only when the database actually changes, reducing unnecessary polling.

## User Interface

```
+------------------------------------------+----------------------------------------------+
| Beads                            [F] [S] | hep-ws-f6.3                                  |
|                                          |----------------------------------------------|
| [-] hep-ws-f1  feature  closed           | Title:  US-003: Install SignalR client and    |
| [-] hep-ws-f2  feature  closed           |         create GameHub connection hook        |
| [-] hep-ws-f3  epic     closed           | Type:   task         Status: closed           |
|   |-- hep-ws-f3.1  task  closed          | Priority: 1          Owner: timoch@timoch.com |
|   |-- hep-ws-f3.2  task  closed          | Parent: hep-ws-f6                             |
|   +-- hep-ws-f3.3  task  closed          | Created: 2026-03-10  Closed: 2026-03-10      |
| [-] hep-ws-f4  feature  closed           |----------------------------------------------|
| [-] hep-ws-f5  epic     closed           | DESCRIPTION                                  |
|   |-- hep-ws-f5.1  task  closed          | Add the SignalR JavaScript client to the      |
|   |-- hep-ws-f5.2  task  closed          | frontend and create a React hook that manages |
|   +-- hep-ws-f5.3  task  closed          | the GameHub WebSocket connection lifecycle... |
|>[-] hep-ws-f6  epic     closed           |                                               |
|   |-- hep-ws-f6.1  task  closed          | DESIGN                                        |
|   |-- hep-ws-f6.2  task  closed          | - Run in `frontend/`: `npm install            |
|  >|-- hep-ws-f6.3  task  closed          |   @microsoft/signalr`                         |
|   |-- hep-ws-f6.4  task  closed          | - Create `frontend/src/hooks/useGameHub.ts`:  |
|   |-- hep-ws-f6.5  task  closed          |   ...                                         |
|   |-- hep-ws-f6.6  task  closed          |                                               |
|   |-- hep-ws-f6.8  task  closed          | ACCEPTANCE CRITERIA                           |
|   +-- hep-ws-f6.9  task  closed          | - [ ] `@microsoft/signalr` appears in ...     |
|                                          | - [ ] `useGameHub` hook exists and compiles   |
|                                          | ...                                           |
|                                          |                                               |
|                                          | DEPENDENCIES                                  |
|                                          |   blocks: hep-ws-f6.1, hep-ws-f6.2           |
|                                          |   parent: hep-ws-f6                           |
|                                          |                                               |
|                                          | NOTES                                         |
|                                          | See docs/features/F6-end-to-end.md for full   |
|                                          | spec. This is Task 3 of F6.                   |
+------------------------------------------+----------------------------------------------+
| [q] Quit  [/] Search  [f] Filter  [?] Help  | Refreshed 1s ago                           |
+---------------------------------------------------------------------------------------------+
```

### Left Panel: Tree View

- Beads are organized in a tree by parent-child relationships.
- Top-level beads (no parent) are root nodes.
- Children are indented under their parent with tree-drawing characters (`|--`, `+--`).
- Collapsible: press Enter or arrow-right to expand, arrow-left to collapse.
- Each row shows: `<id>  <type>  <status>`
- Type uses short labels: `epic`, `feat`, `task`, `bug`, `chore`, `adr`
- Status uses color + icon:
  - `open` -- white/default `( )`
  - `in_progress` -- yellow `(~)`
  - `blocked` -- red `(!)`
  - `deferred` -- dim/gray `(z)`
  - `closed` -- green `(x)`
- The currently selected bead is highlighted.
- Scrollable: j/k or arrow keys to navigate.

### Right Panel: Detail View

- Shows all non-empty fields of the selected bead.
- Header section: id, title, type, status, priority, owner, parent, dates.
- Body sections (each with a heading): Description, Design, Acceptance Criteria, Notes, Dependencies.
- Markdown-aware rendering: bold, lists, code blocks rendered with terminal formatting.
- Scrollable independently of the tree panel (Tab to switch focus, then j/k or arrows to scroll).

### Status Bar

- Shows keybinding hints.
- Shows time since last refresh.
- Shows active filter if any.

## User Stories

### US-01: View bead tree

**As a** developer reviewing a backlog,
**I want to** see all beads organized as a tree in my terminal,
**so that** I can understand the hierarchy of epics, features, and tasks at a glance.

**Acceptance Criteria:**
- The TUI launches and displays a tree of all beads from the local `.beads` database.
- Top-level beads appear as root nodes; children are nested with indentation.
- Each node shows the bead ID, type (abbreviated), and status with color coding.
- The tree is scrollable with j/k or arrow keys.
- Parent nodes can be collapsed/expanded with Enter or left/right arrows.

### US-02: View bead details

**As a** developer picking up a task,
**I want to** select a bead and see all its fields in a detail pane,
**so that** I can read the full description, design, acceptance criteria, and notes without running separate commands.

**Acceptance Criteria:**
- Selecting a bead in the tree populates the right panel with all non-empty fields.
- The detail pane shows: title, type, status, priority, owner, parent, created/updated/closed dates.
- Body sections (description, design, acceptance_criteria, notes) are displayed with headings.
- Dependency information is shown (both "depends on" and "depended on by").
- Long content is scrollable within the detail pane (Tab to focus, j/k to scroll).
- Markdown formatting (bold, lists, code blocks) is rendered with terminal styling.

### US-03: Real-time refresh

**As a** developer editing beads in another terminal,
**I want** the TUI to automatically reflect changes,
**so that** I can see the effect of my edits without restarting the viewer.

**Acceptance Criteria:**
- The TUI periodically re-fetches the bead list (default interval: 2s, configurable via `--refresh` flag).
- When a bead is added, removed, or updated, the tree updates in place without losing the user's scroll position or selection.
- If the currently-selected bead was updated, the detail pane refreshes.
- The status bar shows time since last successful refresh.
- Alternatively, file-system watching on `.beads/` can trigger on-demand refreshes.

### US-04: Filter and search

**As a** developer with a large backlog,
**I want to** filter the tree by type, status, or text,
**so that** I can focus on relevant beads.

**Acceptance Criteria:**
- Pressing `/` opens a search prompt; typing filters the tree to beads whose ID or title contains the search string.
- Pressing `f` opens a filter menu with options: by type (multi-select), by status (multi-select), by assignee.
- Active filters are shown in the status bar.
- Pressing Escape clears the current filter/search.
- The tree re-roots to show only matching beads and their ancestors (to preserve hierarchy context).

### US-05: Configurable layout

**As a** user with different terminal sizes,
**I want** the layout to adapt to my terminal,
**so that** the TUI is usable on narrow and wide screens.

**Acceptance Criteria:**
- The split between tree and detail panes adjusts on terminal resize.
- On narrow terminals (< 100 cols), the detail pane is hidden; pressing Enter on a bead shows detail in a full-screen overlay.
- The tree panel width can be adjusted with `h`/`l` (or a similar binding) when focused.
- Minimum terminal size: 80x24.

### US-06: Command-line options

**As a** user,
**I want** to configure the viewer from the command line,
**so that** I can point it at different databases and control behavior.

**Acceptance Criteria:**
- Runs in a directory containing a `.beads` directory (auto-discovered by `bd`).
- `--refresh <seconds>` sets the polling interval (default: 2).
- `--filter <query>` applies an initial filter on launch (using `bd query` syntax).
- `--type <type>` filters to a specific bead type on launch.
- `--status <status>` filters to a specific status on launch.
- `--expand-all` starts with all tree nodes expanded.
- `--no-color` disables color output for piping/accessibility.

## Keybindings

| Key | Action |
|---|---|
| `j` / `Down` | Move selection down |
| `k` / `Up` | Move selection up |
| `Enter` / `Right` | Expand tree node / open detail overlay (narrow mode) |
| `Left` | Collapse tree node |
| `Tab` | Switch focus between tree and detail pane |
| `/` | Open search |
| `f` | Open filter menu |
| `Esc` | Clear search/filter, close overlays |
| `r` | Force refresh now |
| `e` | Expand all nodes |
| `c` | Collapse all nodes |
| `g` | Go to top |
| `G` | Go to bottom |
| `q` | Quit |
| `?` | Show help overlay |

## Technology Recommendations

The TUI should be a standalone binary for easy distribution. Recommended options:

| Option | Framework | Pros | Cons |
|---|---|---|---|
| **Go + Bubble Tea** | [bubbletea](https://github.com/charmbracelet/bubbletea) + [lipgloss](https://github.com/charmbracelet/lipgloss) | Same language as `bd` itself; could eventually be integrated as a `bd tui` subcommand; Charm ecosystem is mature and well-documented | - |
| Rust + Ratatui | [ratatui](https://github.com/ratatui/ratatui) | Excellent performance; great ecosystem | Different language from `bd` |
| Python + Textual | [textual](https://github.com/Textualize/textual) | Rapid prototyping; rich widget library | Requires Python runtime; slower |

**Recommendation: Go + Bubble Tea.** It shares the `bd` toolchain, and the long-term path is to ship this as `bd tui` -- a built-in subcommand of the `bd` CLI itself.

## Architecture

```
+-----------------+       JSON/stdout        +------------+
|   bd CLI        | <----------------------- |  .beads/   |
|  (subprocess)   |                          |  (Dolt DB) |
+-----------------+                          +------------+
        |
        | stdout (JSON)
        v
+-----------------+
|  TUI App        |
|  +------------+ |
|  | Data Layer | |  -- parses JSON, builds in-memory bead graph
|  +------------+ |
|  | Tree Model | |  -- parent-child tree with expand/collapse state
|  +------------+ |
|  | View Layer | |  -- renders tree panel + detail panel
|  +------------+ |
|  | Input Loop | |  -- handles keybindings, dispatches updates
|  +------------+ |
+-----------------+
```

### Data Flow

1. On startup, run `bd list --all --json --limit 0` to load all beads.
2. Build an in-memory tree: group beads by `parent` field, sort children by priority then ID.
3. Render the tree panel and await user input.
4. On selection change, run `bd show <id> --json` (or use cached data from the list if sufficient).
5. Every N seconds (or on fs-notify), re-run the list command and diff against the in-memory state.
6. Apply incremental updates to the tree, preserving UI state (selection, scroll, expand/collapse).

## Testing Requirements

The Elm architecture (Model-Update-View) used by Bubble Tea enables testing at three distinct levels. All three levels are required.

### Level 1: Pure Logic Tests (required, high coverage)

Test the `Update` function in isolation -- feed it a model and a message, assert on the returned model. No UI rendering, no subprocesses. This is where the bulk of test coverage lives.

Must cover:
- **Tree navigation:** Moving selection up/down, expanding/collapsing nodes, wrapping at boundaries.
- **Tree construction:** Building the parent-child hierarchy from flat bead list, correct sort order (priority then ID), handling orphaned children gracefully.
- **Filter/search state:** Applying a text filter produces the correct subset of visible beads, ancestors of matches remain visible, clearing filter restores full tree.
- **Selection preservation:** After a refresh that adds/removes beads, the selected bead stays selected (or falls back to nearest neighbor if deleted).
- **Data diffing:** Given an old bead list and a new one, correctly identify added, removed, and updated beads.

### Level 2: View Snapshot Tests (required, moderate coverage)

Call `View()` on a model and assert on the returned string. Use golden-file snapshots (via `teatest`) to catch unintended rendering regressions.

Must cover:
- **Tree rendering:** Correct indentation, tree-drawing characters, type labels, status icons and colors.
- **Detail pane rendering:** Header fields laid out correctly, body sections present with headings, long text wrapped.
- **Empty states:** No beads loaded, no bead selected, bead with missing optional fields.
- **Narrow terminal:** Detail pane hidden, overlay shown instead.

### Level 3: Integration Tests with `teatest` (required, key workflows)

Run the full Bubble Tea program in a virtual terminal. Send keystroke sequences, assert on rendered frames.

Must cover:
- **Startup flow:** Launch with a mock `bd` command, verify tree appears with expected beads.
- **Navigation + detail:** Arrow down to a bead, verify detail pane updates.
- **Expand/collapse:** Expand a parent, verify children appear; collapse, verify they disappear.
- **Search:** Press `/`, type a query, verify tree filters; press Escape, verify tree restores.
- **Refresh:** Simulate a `bd` output change between polls, verify tree updates without losing selection.
- **Resize:** Change terminal dimensions, verify layout adapts.

### Test Infrastructure

- **Mock `bd` CLI:** Tests must not depend on a real `.beads` database. Provide a test helper that stubs `bd` output by injecting a command executor interface.
- **Golden files:** Store expected `View()` output as `.golden` files in `testdata/`. Use `teatest.RequireEqualOutput` or equivalent. Update goldens explicitly with a flag (`-update`).
- **CI:** All three test levels run in CI on every push. No test may require a real terminal (TTY).

## Open Questions

1. **Inline vs subprocess:** Should the TUI call `bd` as a subprocess, or import `bd`'s Go packages directly for faster data access? Subprocess is simpler and more decoupled; direct import is faster but couples to `bd` internals.
2. **Markdown rendering depth:** How much markdown rendering is worth doing in the terminal? Full rendering (with `glamour`) or simplified (just bold/dim/indent for lists)?
3. **Dependency graph view:** Should there be a third panel or overlay showing the dependency DAG (using `bd graph` output)? This could be a future enhancement.
4. **Theme support:** Should the TUI support custom color themes, or just respect terminal colors?
