# Contributing to psjungle

Thank you for your interest in contributing to psjungle! This document provides guidelines and information to help make the contribution process smooth and effective.

## How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature-name`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin feature/your-feature-name`)
5. Create a Pull Request

## Development Setup

1. Install Go version 1.24.2 or later
2. Clone the repository
3. Run tests: `go test ./...`
4. Build: `go build ./cmd/psjungle`

## Reporting Issues

Please use the GitHub issue tracker to report bugs or suggest features. When reporting a bug, please include:

- Your operating system and version
- The version of psjungle you're using
- Steps to reproduce the issue
- Expected behavior
- Actual behavior

## Code Style

- Follow standard Go formatting (`go fmt`)
- Write clear, concise commit messages
- Add tests for new functionality
- Update documentation as needed

## Pull Request Guidelines

- Keep changes focused on a single feature or bug fix
- Include tests for any new functionality
- Update README.md if adding or changing features
- Ensure all tests pass before submitting

## License

By contributing to psjungle, you agree that your contributions will be licensed under the MIT License.