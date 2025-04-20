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

# 🧪 Run all tests with verbose output
test:
    go test -v ./...

# 🔨 Build for multiple platforms using the build script
build:
    build/multi_platform.sh
