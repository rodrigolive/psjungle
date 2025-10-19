# psjungle Documentation

## Overview

`psjungle` is a process tree visualization tool that combines the functionality of `ps`, `pgrep`, `lsof`, and `pstree` into a single command with live CPU and memory monitoring.

## Basic Usage

```bash
psjungle [options] [PID|:port|pattern]...
```

## Input Types

psjungle accepts several types of input:

1. **PID**: A numeric process ID (e.g., `1234`)
2. **Port**: A colon followed by a port number (e.g., `:8080`)
3. **Pattern**: A string used for matching process names or command lines

## Matching Modes

### Regex Mode (Default)

By default, psjungle treats input patterns as regular expressions:

```bash
psjungle node                # Match processes whose name/command line contains "node"
psjungle "node.*8080"        # Match processes with both "node" and "8080" in name/command line
psjungle "[Ss]erver"         # Match processes containing "Server" or "server"
```

### Strict Mode (-s/--strict)

When the `-s` or `--strict` flag is used, psjungle performs exact string matching instead of regex:

```bash
psjungle -s "starman"        # Match processes that contain the exact string "starman"
psjungle -s "node.*8080"     # Match processes that contain the exact string "node.*8080" (NOT a regex)
psjungle -s -w2 starman      # Watch mode with strict matching for processes containing "starman"
```

This is particularly useful for processes that appear differently in process names vs. command lines, such as starman processes which show as "perl" in the process name but contain "starman" in their command line.

## Examples

### By PID

```bash
psjungle 1234                # Show process tree rooted at PID 1234
```

### By Port

```bash
psjungle :8080               # Show process trees for processes listening on port 8080
```

### By Name/Pattern (Regex Mode)

```bash
psjungle node                # Find all processes matching "node" (regex)
psjungle "node.*8080"        # Find processes matching both "node" and "8080" (regex)
psjungle worker              # Find all processes matching "worker" (regex)
```

### By Exact String (Strict Mode)

```bash
psjungle -s "starman"        # Find processes containing exact string "starman"
psjungle -s "node.*8080"     # Find processes containing exact string "node.*8080" (not regex)
psjungle -s "timeout"        # Find processes containing exact string "timeout"
```

### Multiple Arguments

When providing multiple arguments, all arguments are treated as PIDs and psjungle intelligently shows separate trees only when needed:

```bash
psjungle 1234 5678           # Display process trees for multiple PIDs
psjungle 1 1234 4321         # Show trees for root process and two other PIDs
```

### Watch Mode

Use the `-w` flag to continuously refresh the output:

```bash
psjungle -w 1234             # Watch PID 1234 (refresh every 2 seconds by default)
psjungle -w=5 :8080           # Watch port 8080 (refresh every 5 seconds)
psjungle -w2 node            # Watch processes matching "node" (refresh every 2 seconds)
psjungle -s -w2 starman      # Watch processes containing exact string "starman" (refresh every 2 seconds)
```

## Key Differences

### Regex vs Strict Mode

- **Regex Mode**: `psjungle "star.an"` will match "starman" because `.` matches any character
- **Strict Mode**: `psjungle -s "star.an"` will NOT match "starman" because it looks for the literal string "star.an"

### Pattern Matching Behavior

Both regex and strict modes search in both the process name and full command line:

```bash
psjungle starman             # Finds starman processes (regex mode)
psjungle -s starman          # Finds starman processes (strict mode)
```

These will both match a process that appears in `ps aux` as:
```
user  47605  0.0  0.1 perl /path/to/starman ...
```

## Output Format

Each line prints: `PID CPU% Memory CommandLine`

- PID: Process ID
- CPU%: Current CPU percentage usage
- Memory: Current resident memory usage in human-readable format (KB/MB/GB)
- CommandLine: Full command line of the process

Target processes are highlighted in green.

## Command Line Options

- `-w`, `--watch`: Watch mode with refresh interval
- `-f`, `--flat`: Flat mode (removes tree indentation)
- `-s`, `--strict`: Strict mode (exact string matching instead of regex)
- `-h`, `--help`: Show help text

## Special Features

1. **Intelligent Tree Display**: When showing multiple PID trees, only displays separate trees when processes are not in the same hierarchy.

2. **Memory Formatting**: Displays memory usage in human-readable units (KB/MB/GB).

3. **Process Context**: Always shows process trees rooted from PID 1 to provide context.

4. **Cross-Platform**: Pure Go implementation using gopsutil for consistent behavior across platforms.