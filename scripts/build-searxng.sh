#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARIES_DIR="$PROJECT_ROOT/src-tauri/binaries"
VENV_DIR="$PROJECT_ROOT/.venv"

# Create venv if it doesn't exist
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv "$VENV_DIR"
fi

source "$VENV_DIR/bin/activate"

echo "Installing PyInstaller..."
pip install --quiet pyinstaller msgspec

echo "Installing SearXNG from source..."
pip install --quiet --no-build-isolation "git+https://github.com/searxng/searxng.git@master"

# Find the searx package location for --collect-all
SEARX_DIR=$(python -c "import searx; import os; print(os.path.dirname(searx.__file__))")
echo "SearXNG found at: $SEARX_DIR"

echo "Building SearXNG sidecar binary..."
pyinstaller --onefile \
    --add-data "$PROJECT_ROOT/searxng/settings.yml:." \
    --collect-all searx \
    --hidden-import searx.webapp \
    --hidden-import searx.search \
    --hidden-import searx.engines \
    --hidden-import searx.plugins \
    --hidden-import searx.botdetection \
    --hidden-import flask_babel \
    --hidden-import babel \
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
