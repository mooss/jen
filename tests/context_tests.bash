#!/bin/bash
set -euo pipefail

#############
# Functions #
function run() {
  go run ./go/ai/jenai.go commit_message --dry-run "$@"
}

#########
# Setup #
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR/testdir"
echo "file1 content" > "$TMP_DIR/testdir/file1.txt"
echo "file2 content" > "$TMP_DIR/testdir/file2.txt"
echo "nested content" > "$TMP_DIR/testdir/nested.txt"

#########
# Tests #
# Single file inclusion.
echo -e "\nTEST: Single file context"
output=$(run --file "$TMP_DIR/testdir/file1.txt")
echo "$output" | grep -q "====> START OF $TMP_DIR/testdir/file1.txt" || {
    echo "Single file context failed"
    exit 1
}

# Multiple files.
echo -e "\nTEST: Multiple file context"
output=$(run --file "$TMP_DIR/testdir/file1.txt" "$TMP_DIR/testdir/file2.txt")
echo "$output" | grep -q "file1 content" || {
    echo "File1 content missing"
    exit 1
}
echo "$output" | grep -q "file2 content" || {
    echo "File2 content missing"
    exit 1
}

# Directory inclusion.
echo -e "\nTEST: Directory context"
output=$(run --dir "$TMP_DIR/testdir")
echo "$output" | grep -q "nested content" || {
    echo "Directory context missing nested file"
    exit 1
}

# Context placement (above/below).
echo -e "\nTEST: Context placement"
output_above=$(run --context-above --dir "$TMP_DIR/testdir")
output_below=$(run --file "$TMP_DIR/testdir/file1.txt")

if ! echo "$output_above" | head -n1 | grep -q "^# Additional context"; then
    echo "Context above failed"
    exit 1
fi

if ! echo "$output_below" | tail -n8 | grep -q "^# Additional context"; then
    echo "Context below failed"
    exit 1
fi

echo -e "\nAll context tests passed!"
