# ── One-command setup ─────────────────────────────────────────────────
# Get from zero to running with: just setup && just dev

# Full first-time setup: install deps, pull models, build sidecar
setup:
    #!/usr/bin/env bash
    set -euo pipefail

    echo "Installing frontend dependencies..."
    cd frontend && npm install && cd ..

    echo "Fetching Rust dependencies..."
    cd src-tauri && cargo fetch && cd ..

    echo "Building SearXNG sidecar..."
    bash scripts/build-searxng.sh

    echo "Pulling LLM models (this may take a few minutes)..."
    ollama pull qwen3.5:4b
    ollama pull qwen3.5:0.8b
    ollama pull nomic-embed-text

    echo ""
    echo "Done! Run 'just dev' to start."

# Build the SearXNG sidecar binary via PyInstaller
build-searxng:
    bash scripts/build-searxng.sh

# Run in development mode
dev:
    #!/usr/bin/env bash
    set -euo pipefail

    # Start Ollama if needed
    if ! curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "Starting Ollama..."
        ollama serve &
        sleep 2
    fi

    cargo tauri dev

# Build the production desktop app
build:
    cargo tauri build

# ── Checks ───────────────────────────────────────────────────────────

# Run all checks
check: check-rust check-frontend

# Lint and test Rust
check-rust:
    cd src-tauri && cargo clippy -- -D warnings && cargo test

# Lint and typecheck frontend
check-frontend:
    cd frontend && npx biome check . && npx tsc --noEmit

# ── Utilities ────────────────────────────────────────────────────────

# Pull an Ollama model
pull model:
    ollama pull {{ model }}

# Upgrade Ollama
upgrade-ollama:
    brew upgrade ollama || curl -fsSL https://ollama.com/install.sh | sh

# Benchmark model performance
bench model="qwen3.5:4b":
    python3 scripts/bench.py {{ model }}
