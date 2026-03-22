# ── One-command setup ─────────────────────────────────────────────────
# Get from zero to running with: just setup && just dev

# Full first-time setup: install deps, pull models, start services
setup:
    #!/usr/bin/env bash
    set -euo pipefail

    echo "Installing frontend dependencies..."
    cd frontend && npm install && cd ..

    echo "Fetching Go dependencies..."
    go mod tidy

    echo "Starting SurrealDB + SearXNG via Docker..."
    docker compose up -d

    echo "Pulling LLM models (this may take a few minutes)..."
    ollama pull qwen3.5:4b
    ollama pull qwen3.5:0.8b
    ollama pull nomic-embed-text

    echo ""
    echo "Done! Run 'just dev' to start."

# Run in development mode (Go server + Next.js dev server)
dev:
    #!/usr/bin/env bash
    set -euo pipefail

    # Start Ollama if needed
    if ! curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "Starting Ollama..."
        ollama serve &
        sleep 2
    fi

    # Ensure Docker services are running
    docker compose up -d

    # Start Go backend
    go run . &
    GO_PID=$!

    # Start Next.js frontend
    cd frontend && npm run dev &
    NEXT_PID=$!

    # Wait for either to exit
    trap "kill $GO_PID $NEXT_PID 2>/dev/null" EXIT
    wait

# Build the production web server
build:
    docker compose up -d
    go build -o zoro .
    cd frontend && npm run build

# Stop Docker services
stop:
    docker compose down

# ── Desktop (Wails) ──────────────────────────────────────────────────

# Build desktop app (SurrealDB + SearXNG are set up on first launch)
build-desktop:
    #!/usr/bin/env bash
    set -euo pipefail

    echo "Building frontend static export..."
    cd frontend && NEXT_BUILD_TARGET=desktop npm run build && cd ..

    echo "Copying frontend assets for embed..."
    rm -rf cmd/desktop/assets
    cp -r frontend/out cmd/desktop/assets

    echo "Building desktop binary..."
    CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags desktop,production -o zoro-desktop ./cmd/desktop

    echo ""
    echo "Done! Run ./zoro-desktop to launch."
    echo "  First launch downloads SurrealDB and installs SearXNG into a local venv."
    echo "  Requires: Ollama, Python 3.10+"

# Run desktop app in dev mode (no Docker needed)
dev-desktop:
    #!/usr/bin/env bash
    set -euo pipefail

    # Start Ollama if needed
    if ! curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "Starting Ollama..."
        ollama serve &
        sleep 2
    fi

    # Build frontend for embedding
    cd frontend && NEXT_BUILD_TARGET=desktop npm run build && cd ..
    rm -rf cmd/desktop/assets
    cp -r frontend/out cmd/desktop/assets

    # Run desktop app (manages SurrealDB + SearXNG as subprocesses)
    CGO_LDFLAGS="-framework UniformTypeIdentifiers" go run -tags desktop,dev ./cmd/desktop

# ── Checks ───────────────────────────────────────────────────────────

# Run all checks
check: check-go check-frontend

# Lint and vet Go
check-go:
    go vet ./...

# Lint and typecheck frontend
check-frontend:
    cd frontend && npx biome check . && npx tsc --noEmit

# ── Code generation ─────────────────────────────────────────────────

# Generate frontend API client from OpenAPI spec
generate:
    oag generate

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
