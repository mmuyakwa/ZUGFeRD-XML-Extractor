#!/bin/bash

# Build script for ZUGFeRD XML Extractor
echo "Building ZUGFeRD XML Extractor..."

# Create build directory
mkdir -p build

# Build for different platforms
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o build/zugferd-extractor.exe ./cmd/zugferd-extractor

echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o build/zugferd-extractor-linux ./cmd/zugferd-extractor

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o build/zugferd-extractor-macos ./cmd/zugferd-extractor

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o build/zugferd-extractor-macos-arm64 ./cmd/zugferd-extractor

echo "Build complete! Binaries are in the 'build' directory."