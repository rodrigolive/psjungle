# psjungle

A process tree visualization tool for macOS, similar to `pstree` but with additional features.

## Features

- Display process trees for PIDs, ports, process names, or regex patterns
- Shows complete process lines with all arguments (similar to `ps auxww`)
- Human-readable memory formatting:
  - KB for values under 1000KB
  - MB for values under 1000MB (2 decimal places for <10MB, 1 decimal place for ≥10MB)
  - GB for values 1000MB and above (2 decimal places for <10GB, 1 decimal place for ≥10GB)
- Color highlighting of target processes
- Watch mode for continuous monitoring

## Usage

```
psjungle [options] [PID|:port|name|/pattern]

EXAMPLES:
   psjungle 1234               Display process tree for PID 1234
   psjungle :8080              Display process trees for processes listening on port 8080
   psjungle node               Display process trees for processes with "node" in their name
   psjungle "/node.*8080"       Display process trees for processes matching regex pattern
   psjungle -w 1234            Watch process tree for PID 1234 (refresh every 2 seconds)
   psjungle -w=5 1234          Watch process tree for PID 1234 (refresh every 5 seconds)
```

## Output Format

The output includes:
1. PID (Process ID)
2. CPU% (CPU usage percentage)
3. Memory usage (human-readable format: KB, MB, or GB)
4. Complete command line with arguments

Target processes are highlighted in green.