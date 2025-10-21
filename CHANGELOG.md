# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.2] - 2025-10-21

### Added
- Signal sending functionality with -k/--kill flag
- --host option for port filtering and enhanced port matching

### Changed
- Modularized app.go by extracting functions and improving code structure

## [v1.1] - 2025-10-20

### Added
- Strict mode flag for exact string matching
- Enhanced process lookup functionality
- Better tree visualization

## [v1.0] - 2025-10-19

### Added
- Initial release of psjungle
- Process tree visualization with live CPU and memory monitoring
- Support for finding processes by PID, port, name fragment, or regex pattern
- Watch mode for continuously refreshing output
- Cross-platform compatibility with pure Go implementation