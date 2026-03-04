# Default model
autocomplete_model := "qwen3.5:4b"

# Start infrastructure (Neo4j + SearXNG)
infra:
    docker compose up neo4j searxng -d

# Stop infrastructure
infra-down:
    docker compose down

# Run Go API (hot reload with air)
api:
    cd api && air

# Run frontend dev server
web:
    cd frontend && npm run dev

# Run api + frontend with hot reload
run:
    #!/usr/bin/env bash
    trap 'kill 0' EXIT
    cd api && air &
    cd frontend && npm run dev &
    wait

# Run everything (infra + api + web)
dev:
    #!/usr/bin/env bash
    trap 'kill 0; docker compose down' EXIT
    docker compose up neo4j searxng -d
    echo "Waiting for infrastructure..."
    sleep 5
    cd api && air &
    cd frontend && npm run dev &
    wait

# Upgrade ollama (qwen3.5 requires v0.17.1+)
upgrade-ollama:
    brew upgrade ollama || curl -fsSL https://ollama.com/install.sh | sh

# Pull the autocompletion model
pull model=autocomplete_model:
    ollama pull {{ model }}

# Start ollama serve (if not already running)
serve:
    #!/usr/bin/env bash
    if curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "ollama is already running"
    else
        echo "Starting ollama..."
        ollama serve &
        sleep 2
        echo "ollama started"
    fi

# Benchmark TPS for a single model
bench model=autocomplete_model:
    python3 scripts/bench.py {{ model }}

# Benchmark all small Qwen 3.5 models
bench-all:
    python3 scripts/bench.py

# Generate Swagger docs
swagger:
    cd api && swag init -g cmd/zoro/main.go -o docs --parseDependency --parseInternal

# Generate API client from OpenAPI spec
generate:
    oag generate

# Lint and typecheck frontend
check:
    cd frontend && npx biome check . && npx tsc --noEmit

# Run e2e tests
test-e2e:
    python3 scripts/test_e2e.py
