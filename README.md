# psjungle

```
                              ****
                            ********
                           **  ******
                            *   ******     ******
   PSJUNGLE  V1.1               ******   *********
                                 ****  *****   ***
                                 ***  ***     **
                           *************       *
                         ******************
                        *****   H*****H*******
                        ***     H-___-H  *********
                         ***    H     H      *******
                          **    H-___-H        *****
                            *   H     H         ****
                                H     H         ***
                                H-___-H         **
                                H     H         *
                                H-___-H
```

I built psjungle because I got fed up with juggling five different commands
every time a process on my laptop started misbehaving. On macOS I would bounce
between `ps`, `pgrep`, `lsof`, `pstree`, Activity Monitor...

`psjungle` scratches that itch. It provides a `pstree`-like view, but adds live
CPU%, human-readable memory (KB/MB/GB), and robust filtering: by PID (`1234`),
TCP/UDP port (`:8080`), name fragment (`node`), or full regex (`/node.*8080`).

No shell-outs, no calls to `lsof`, `pgrep`, or others. Just pure Go via
`gopsutil` for consistent cross-platform behavior.

## Features

- Display focused process trees by PID, TCP/UDP port (`:8080`), case-insensitive name fragment (`node`), or full regex (`/node.*8080`).
- Full command line output (similar to `ps auxww`) with live CPU% and human-readable memory usage (KB/MB/GB).
- Highlights the target process in green.
- Watch mode (`-w` / `--watch`) for continuously refreshing output every *n* seconds.
- Support for multiple PIDs as arguments, intelligently showing separate trees only when needed.
- Pure Go implementation using `gopsutil` for cross-platform compatibility—no `exec.Command` usage.

## Installation

With Go installed:

```bash
go install ./cmd/psjungle
```

Or build a binary:

```bash
go build ./cmd/psjungle
```

## Usage

```bash
psjungle [options] [PID|:port|name|/pattern]...
```

Examples:

```bash
psjungle 1234                # Inspect the tree for PID 1234
psjungle :8080               # Show trees for processes bound to port 8080
psjungle node                # Match processes whose name contains "node" (case-insensitive)
psjungle "/node.*8080"        # Regex match against command line / name
psjungle 1234 5678           # Display process trees for multiple PIDs (intelligently shows separate trees only when needed)
psjungle 1234 5678 9012      # Display process trees for three PIDs
psjungle -w 1234             # Refresh every 2 seconds (default) while showing PID 1234
psjungle -w=5 :3000          # Refresh every 5 seconds for port 3000 listeners
psjungle -w2 1234            # Refresh every 2 seconds while showing PID 1234 (alternative format)
```

Multiple PID Examples:
When providing multiple PIDs, psjungle intelligently shows separate process trees only when needed (when PIDs are not in the same process tree).

For example, to monitor multiple specific processes:
```bash
psjungle 1 1234 4321         # Show trees for root process and two other specific PIDs
```

To monitor all processes matching a specific name:
```bash
psjungle node                # This will automatically show trees for all processes with "node" in their name
```

## Output Format

Each line prints: `PID CPU% Memory CommandLine`—similar to `ps aux`, but with a process tree view.

Memory is displayed in human-readable units (KB/MB/GB). Target processes are highlighted in green.

## Project Layout

- `cmd/psjungle`: CLI entrypoint.
- `internal/psjungle`: Core tree-building and process lookup logic.
- `scripts`: Build and release scripts.

## Testing

```bash
go test ./...
```

Because the code interacts with the local process table, a few tests skip
automatically if the necessary process information is unavailable (for example,
when PID 1 cannot be inspected).

## CI/CD

This project uses GitHub Actions for continuous integration and release management:

- `.github/workflows/build.yml` - Runs tests and builds on push to main branch and pull requests
- `.github/workflows/release.yml` - Creates releases with binary archives when tags are pushed

To create a new release, push a tag with semantic versioning (e.g., `git tag v1.2 && git push origin v1.2`).
