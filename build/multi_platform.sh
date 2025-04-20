#!/usr/bin/env bash
set -e

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
    GOOS=$GOOS GOARCH=$GOARCH go build -o "bin/$output_name" ./cmd/pseudoc
done
