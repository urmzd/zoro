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

# Build the production app
build:
    docker compose up -d
    go build -o zoro .
    cd frontend && npm run build

# Stop Docker services
stop:
    docker compose down

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
