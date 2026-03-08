#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARIES_DIR="$PROJECT_ROOT/src-tauri/binaries"

echo "Installing SearXNG and PyInstaller..."
pip install --quiet searxng pyinstaller

echo "Building SearXNG sidecar binary..."
pyinstaller --onefile \
    --add-data "$PROJECT_ROOT/searxng/settings.yml:." \
    --name searxng_server \
    --distpath "$PROJECT_ROOT/dist" \
    --workpath "$PROJECT_ROOT/build/pyinstaller" \
    --specpath "$PROJECT_ROOT/build" \
    "$PROJECT_ROOT/scripts/searxng_server.py"

# Tauri expects sidecar binaries named with the target triple
target=$(rustc --print host-tuple)
mkdir -p "$BINARIES_DIR"
cp "$PROJECT_ROOT/dist/searxng_server" "$BINARIES_DIR/searxng-$target"

echo "Sidecar binary ready: $BINARIES_DIR/searxng-$target"
