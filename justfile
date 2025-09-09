default:
    just --list

# 📦 Create a new release with commit-and-tag-version (requires global installation: npm i -g commit-and-tag-version) (args optional, e.g. --release-as 0.1.0)
release *ARGS:
    #!/usr/bin/env sh
    if ! command -v commit-and-tag-version > /dev/null 2>&1; then
        echo "Error: commit-and-tag-version command not found."
        echo "Please install it globally with: npm i -g commit-and-tag-version"
        exit 1
    fi
    commit-and-tag-version {{ARGS}}

# [🤖 CI task] extract content from CHANGELOG.md for use in Gitlab/Github Releases
release-notes:
    #!/usr/bin/env bash
    set -euo pipefail

    changelog_path="{{invocation_directory()}}/CHANGELOG.md"
    release_notes_path="{{invocation_directory()}}/../LATEST_RELEASE_NOTES.md"

    # Create header for release notes
    echo "## What's changed in this release" > "$release_notes_path"

    # Extract content between first and second level 2 heading
    awk '/^## /{
        if (count == 0) {
            count = 1
            next
        } else if (count == 1) {
            exit
        }
    }
    count == 1 { print }' "$changelog_path" >> "$release_notes_path"

    echo "Release notes extracted to $release_notes_path"

# 🧪 Run tests with better output formatting
test:
    @echo "🧪 Running tests..."
    gotestsum --format=pkgname-and-test-fails

# 🏃‍♂️ Run tests with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report generated: coverage.html"

# 🔧 Build for local development
build:
    go build -o bin/pseudoc ./cmd/pseudoc

# 🔧 Build with version information (for releases)
build-release:
    #!/usr/bin/env bash
    set -euo pipefail

    # Get build information
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")

    # Build with ldflags to embed version info
    go build -ldflags "\
        -X 'main.version=${VERSION}' \
        -X 'main.buildTime=${BUILD_TIME}' \
        -X 'main.gitCommit=${GIT_COMMIT}'" \
        -o bin/pseudoc ./cmd/pseudoc

    echo "Built pseudoc with version info:"
    echo "  Version: ${VERSION}"
    echo "  Build Time: ${BUILD_TIME}"
    echo "  Git Commit: ${GIT_COMMIT}"

# 🔨 Build for multiple platforms using the build script
build-multi:
    build/multi_platform.sh

# 🧹 Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

# 🚀 Run the tool locally (for testing)
run *ARGS:
    go run ./cmd/pseudoc {{ARGS}}

# 📝 Show examples of usage
examples:
    @echo "Examples of pseudoc usage:"
    @echo ""
    @echo "# Generate a single PDF:"
    @echo "  pseudoc pdf"
    @echo ""
    @echo "# Generate 5 Word documents with 3 pages each:"
    @echo "  pseudoc docx --count 5 --pages 3"
    @echo ""
    @echo "# Generate Excel file with custom filename:"
    @echo "  pseudoc xlsx --filename my-data --sheets 4"
    @echo ""
    @echo "# Generate 10 random documents:"
    @echo "  pseudoc random --count 10 --output-dir ./test-files"
    @echo ""
    @echo "# Preview what would be generated (dry run):"
    @echo "  pseudoc pdf --count 3 --dry-run"

# 🔍 Check code quality
lint:
    @echo "Running go vet..."
    go vet ./...
    @echo "Running go fmt check..."
    @if [ -n "$$(go fmt ./...)" ]; then \
        echo "Code is not formatted. Please run 'go fmt ./...'"; \
        exit 1; \
    fi
    @echo "✅ Code quality checks passed"

# 🎯 Format code
fmt:
    go fmt ./...
