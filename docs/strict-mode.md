# Strict Mode in psjungle

## Introduction

Strict mode is a feature in psjungle that allows you to perform exact string matching instead of regular expression matching. This is particularly useful for finding processes that contain specific strings in their command lines.

## Usage

To use strict mode, add the `-s` or `--strict` flag to your psjungle command:

```bash
psjungle -s "pattern"
```

## Examples

### Finding starman processes

Starman processes often appear as "perl" in the process name but contain "starman" in their command line. Strict mode makes it easy to find these:

```bash
psjungle -s "starman"
```

This will match processes with command lines like:
```
perl /path/to/starman --workers 4 --host 127.0.0.1
```

### Exact string matching

In strict mode, patterns are treated as literal strings rather than regular expressions:

```bash
psjungle -s "server.*prod"    # Matches processes containing the exact string "server.*prod"
psjungle "server.*prod"       # Matches processes matching the regex pattern (e.g., "server-prod", "server_prod", etc.)
```

### Watch mode with strict matching

You can combine strict mode with watch mode:

```bash
psjungle -s -w2 "starman"     # Watch processes containing "starman" every 2 seconds
```

## Key Benefits

1. **Process name vs command line**: Many processes appear with generic names but specific command lines. Strict mode helps you find these.

2. **No regex escaping**: You don't need to escape special regex characters when you want to match them literally.

3. **Predictable matching**: Exact string matching behaves predictably and is easier to understand.

## Technical Details

Strict mode performs case-insensitive substring matching against both the process name and the full command line. If a match is found in either, the process tree is displayed.