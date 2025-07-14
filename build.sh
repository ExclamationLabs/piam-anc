#!/bin/bash

# Build script for sql-network-manager
# Builds cross-platform binaries for distribution

set -e

VERSION=${1:-"1.0.0"}
BINARY_NAME="piam-anc"
BUILD_DIR="build"
DIST_DIR="dist"

echo "🏗️  Building $BINARY_NAME version $VERSION"

# Clean previous builds
rm -rf "$BUILD_DIR" "$DIST_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR"

# Build info
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS="-s -w -X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT"

# Build for different platforms
echo "📦 Building binaries..."

platforms=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    output_name="${BINARY_NAME}-${os}-${arch}"
    
    echo "  Building $output_name..."
    
    GOOS=$os GOARCH=$arch go build \
        -ldflags "$LDFLAGS" \
        -o "$BUILD_DIR/$output_name" \
        .
    
    # Create tarball with generic binary name for Homebrew
    echo "  Creating tarball for $output_name..."
    cp "$BUILD_DIR/$output_name" "$BUILD_DIR/piam-anc"
    tar -czf "$DIST_DIR/$output_name.tar.gz" -C "$BUILD_DIR" "piam-anc"
    rm "$BUILD_DIR/piam-anc"
done

# Generate checksums
echo "🔐 Generating checksums..."
cd "$DIST_DIR"
shasum -a 256 *.tar.gz > checksums.txt
cd ..

echo "✅ Build complete!"
echo ""
echo "📁 Artifacts created in $DIST_DIR/:"
ls -la "$DIST_DIR"
echo ""
echo "🔍 Checksums:"
cat "$DIST_DIR/checksums.txt"
echo ""
echo "📝 Next steps:"
echo "1. Upload tar.gz files to GitHub releases"
echo "2. Update Homebrew formula with new version and checksums"
echo "3. Create a GitHub release and upload the tarballs from dist/"