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

# Create package directories for each platform
echo "Creating package directories..."
mkdir -p dist/psjungle-$VERSION-darwin-amd64 dist/psjungle-$VERSION-darwin-arm64
mkdir -p dist/psjungle-$VERSION-linux-amd64 dist/psjungle-$VERSION-linux-arm64

# Copy binaries to package directories
cp dist/psjungle-darwin-amd64 dist/psjungle-$VERSION-darwin-amd64/
cp dist/psjungle-darwin-arm64 dist/psjungle-$VERSION-darwin-arm64/
cp dist/psjungle-linux-amd64 dist/psjungle-$VERSION-linux-amd64/
cp dist/psjungle-linux-arm64 dist/psjungle-$VERSION-linux-arm64/

# Copy documentation files to each package directory
cp README.md dist/psjungle-$VERSION-darwin-amd64/
cp README.md dist/psjungle-$VERSION-darwin-arm64/
cp README.md dist/psjungle-$VERSION-linux-amd64/
cp README.md dist/psjungle-$VERSION-linux-arm64/

cp LICENSE dist/psjungle-$VERSION-darwin-amd64/
cp LICENSE dist/psjungle-$VERSION-darwin-arm64/
cp LICENSE dist/psjungle-$VERSION-linux-amd64/
cp LICENSE dist/psjungle-$VERSION-linux-arm64/

cp CHANGELOG.md dist/psjungle-$VERSION-darwin-amd64/
cp CHANGELOG.md dist/psjungle-$VERSION-darwin-arm64/
cp CHANGELOG.md dist/psjungle-$VERSION-linux-amd64/
cp CHANGELOG.md dist/psjungle-$VERSION-linux-arm64/

# Create archives
echo "Creating archives..."

cd dist

# Create tar.gz archives with full package directories
tar -czvf psjungle-$VERSION-darwin-amd64.tar.gz psjungle-$VERSION-darwin-amd64
tar -czvf psjungle-$VERSION-darwin-arm64.tar.gz psjungle-$VERSION-darwin-arm64
tar -czvf psjungle-$VERSION-linux-amd64.tar.gz psjungle-$VERSION-linux-amd64
tar -czvf psjungle-$VERSION-linux-arm64.tar.gz psjungle-$VERSION-linux-arm64

# Create .tgz symbolic links (tgz is just another extension for tar.gz)
# Remove existing links first to avoid errors
rm -f psjungle-$VERSION-darwin-amd64.tgz psjungle-$VERSION-darwin-arm64.tgz
rm -f psjungle-$VERSION-linux-amd64.tgz psjungle-$VERSION-linux-arm64.tgz
ln -s psjungle-$VERSION-darwin-amd64.tar.gz psjungle-$VERSION-darwin-amd64.tgz
ln -s psjungle-$VERSION-darwin-arm64.tar.gz psjungle-$VERSION-darwin-arm64.tgz
ln -s psjungle-$VERSION-linux-amd64.tar.gz psjungle-$VERSION-linux-amd64.tgz
ln -s psjungle-$VERSION-linux-arm64.tar.gz psjungle-$VERSION-linux-arm64.tgz

# Create zip archives for all platforms with full package directories
echo "Creating zip archives..."
zip -r psjungle-$VERSION-darwin-amd64.zip psjungle-$VERSION-darwin-amd64
zip -r psjungle-$VERSION-darwin-arm64.zip psjungle-$VERSION-darwin-arm64
zip -r psjungle-$VERSION-linux-amd64.zip psjungle-$VERSION-linux-amd64
zip -r psjungle-$VERSION-linux-arm64.zip psjungle-$VERSION-linux-arm64

# Clean up package directories and binaries
rm -rf psjungle-$VERSION-darwin-amd64 psjungle-$VERSION-darwin-arm64
rm -rf psjungle-$VERSION-linux-amd64 psjungle-$VERSION-linux-arm64
rm psjungle-darwin-amd64 psjungle-darwin-arm64 psjungle-linux-amd64 psjungle-linux-arm64

echo "Build complete! Archives created in dist/ directory:"
ls -la