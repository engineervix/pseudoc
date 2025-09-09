#!/usr/bin/env bash
set -e

echo "Building pseudoc for multiple platforms..."
echo "=========================================="

# Get build information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")

echo "Version: $VERSION"
echo "Build Time: $BUILD_TIME"
echo "Git Commit: $GIT_COMMIT"
echo ""

# Create bin directory if it doesn't exist
mkdir -p bin

# Build for multiple platforms
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"

for platform in $PLATFORMS; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    output_name="pseudoc-$GOOS-$GOARCH"
    if [ $GOOS = "windows" ]; then
        output_name+=".exe"
    fi

    echo "Building for $GOOS/$GOARCH..."

    # Build with ldflags to embed version information
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X 'main.version=$VERSION' -X 'main.buildTime=$BUILD_TIME' -X 'main.gitCommit=$GIT_COMMIT'" \
        -o "bin/$output_name" ./cmd/pseudoc

    # Show file size for reference
    if command -v ls >/dev/null 2>&1; then
        size=$(ls -lh "bin/$output_name" | awk '{print $5}')
        echo "  ✓ bin/$output_name ($size)"
    else
        echo "  ✓ bin/$output_name"
    fi
done

echo ""
echo "✅ Multi-platform build complete!"
echo "Built binaries:"
ls -la bin/
