# ── One-command setup ─────────────────────────────────────────────────
# Get from zero to running with: just setup && just dev

download:
    echo "Pulling LLM models (this may take a few minutes)..."
    ollama pull qwen3.5:4b
    ollama pull qwen3.5:0.8b
    ollama pull nomic-embed-text

# Full first-time setup: install deps, pull models, start services
setup: download
    #!/usr/bin/env bash
    set -euo pipefail

    echo "Fetching Go dependencies..."
    go mod tidy

    echo "Starting PostgreSQL + SearXNG via Docker..."
    docker compose up -d

    echo ""
    echo "Done! Run 'just dev' to start the MCP server."

# Run MCP server (starts Docker services + Ollama if needed)
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

    # Run MCP server on stdio
    go run . serve

# Build the binary
build:
    go build -o zoro .

# Install the binary
install:
    go install .

# Stop Docker services
stop:
    docker compose down

# ── Checks ───────────────────────────────────────────────────────────

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run govulncheck
vuln:
    govulncheck ./...

# Tidy modules
tidy:
    go mod tidy

# Run all checks
check:
    go vet ./...

# Full CI gate
ci: check build

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
