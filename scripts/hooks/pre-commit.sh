#!/bin/bash

# Format staged Go files with `gofumpt`
function format_modified_go_files() {
    # Get the list of staged Go files.
    files=$(git diff --cached --name-only --diff-filter=ACMR | grep '\.go$')

    # Check if gofumpt is installed.
    if ! command -v gofumpt &> /dev/null; then
        echo "Error: gofumpt is not installed. Please install it by running 'go install mvdan.cc/gofumpt@latest'" >&2
        exit 1
    fi

    # Check if there are Go files to format.
    if [ -n "$files" ]; then
        echo "Running gofumpt on staged Go files:"
        for file in $files; do
            gofumpt -l -w "$file"
            git add "$file"
        done
    fi
}

# Lint modified go files
function lint_modified_go_files() {
    echo "Running golangci-lint on staged Go files:"
    golangci-lint run --fix --new --fast -c .golangci.yml
}

format_modified_go_files
lint_modified_go_files