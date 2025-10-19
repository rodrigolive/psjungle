#!/bin/bash

# Script to build psjungle for multiple platforms and create archives

VERSION=${1:-"latest"}

echo "Building psjungle version: $VERSION"

# Create dist directory
mkdir -p dist

# Build for macOS (amd64 and arm64)
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o dist/psjungle-darwin-amd64 ./cmd/psjungle
GOOS=darwin GOARCH=arm64 go build -o dist/psjungle-darwin-arm64 ./cmd/psjungle

# Build for Linux (amd64 and arm64)
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o dist/psjungle-linux-amd64 ./cmd/psjungle
GOOS=linux GOARCH=arm64 go build -o dist/psjungle-linux-arm64 ./cmd/psjungle

# Create archives
echo "Creating archives..."

cd dist

# Create tar.gz archives
tar -czvf psjungle-$VERSION-darwin-amd64.tar.gz psjungle-darwin-amd64
tar -czvf psjungle-$VERSION-darwin-arm64.tar.gz psjungle-darwin-arm64
tar -czvf psjungle-$VERSION-linux-amd64.tar.gz psjungle-linux-amd64
tar -czvf psjungle-$VERSION-linux-arm64.tar.gz psjungle-linux-arm64

# Create zip archives for macOS (Windows users typically expect zip)
zip psjungle-$VERSION-darwin-amd64.zip psjungle-darwin-amd64
zip psjungle-$VERSION-darwin-arm64.zip psjungle-darwin-arm64

# Remove binaries
rm psjungle-darwin-amd64 psjungle-darwin-arm64 psjungle-linux-amd64 psjungle-linux-arm64

echo "Build complete! Archives created in dist/ directory:"
ls -la