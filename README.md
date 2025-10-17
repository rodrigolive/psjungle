# psjungle

A pure-Go process tree visualization tool for macOS and Linux. It works like `pstree`, but includes richer filtering options, human-readable metrics, and an interactive watch mode without shelling out to `lsof`, `pgrep`, or other external binaries.

## Features

- Display focused process trees by PID, TCP/UDP port (`:8080`), case-insensitive name fragment (`node`), or full regex (`/node.*8080`).
- Full command line output (similar to `ps auxww`) with live CPU% and human-readable memory usage (KB/MB/GB).
- Highlights the target process in green.
- Watch mode (`-w` / `--watch`) for continuously refreshing output every *n* seconds.
- Pure Go implementation using `gopsutil` for cross-platform compatibility—no `exec.Command` usage.

## Building & Running

```bash
# Run from source
go run ./cmd/psjungle --help

# Build a binary
go build ./cmd/psjungle
```

Use the resulting binary just like any other CLI tool:

```
psjungle [options] [PID|:port|name|/pattern]
```

### Examples

```
psjungle 1234                # Inspect the tree for PID 1234
psjungle :8080               # Show trees for processes bound to port 8080
psjungle node                # Match processes whose name contains "node" (case-insensitive)
psjungle "/node.*8080"        # Regex match against command line / name
psjungle -w 1234             # Refresh every 2 seconds (default) while showing PID 1234
psjungle -w=5 :3000          # Refresh every 5 seconds for port 3000 listeners
```

When experimenting with watch mode in scripts or tests, prepend the command with `timeout` to stop it automatically:

```bash
timeout 10s psjungle -w 2 1234
```

## Output Format

Each line prints: `PID CPU% Memory CommandLine`. Memory is displayed in the smallest unit that keeps the value above three digits (KB, MB, GB). Target processes are highlighted in green, and tree glyphs (`├──`, `└──`, `│`) visualize parent/child relationships.

## Project Layout

- `cmd/psjungle`: CLI entrypoint.
- `internal/psjungle`: Core tree-building, rendering, and lookup logic.
- `internal/psjungle_test`: Black-box tests that exercise the exported API.

## Testing

Run the full test suite with:

```bash
go test ./...
```

Because the code interacts with the local process table, a few tests skip automatically if the necessary process information is unavailable (for example, when PID 1 cannot be inspected).*** End Patch
