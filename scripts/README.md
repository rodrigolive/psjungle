# Scripts

This directory contains scripts used for building and releasing psjungle.

## build-release.sh

This script builds psjungle for multiple platforms (macOS, Linux, and Windows) and creates compressed archives (tar.gz, tgz, and zip) for distribution.

Usage:
```bash
./scripts/build-release.sh [version]
```

Example:
```bash
./scripts/build-release.sh v1.2
```

The script creates archives for:
- macOS (darwin) amd64 and arm64
- Linux amd64 and arm64
- Windows amd64 and arm64

Archive formats:
- .tar.gz (standard tar.gz compression)
- .tgz (symbolic links to tar.gz files)
- .zip (for all platforms)