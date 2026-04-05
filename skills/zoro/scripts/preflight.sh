#!/usr/bin/env bash
set -euo pipefail

errors=0

# Check zoro binary
if ! command -v zoro &>/dev/null; then
  echo "FAIL: zoro binary not found on PATH"
  echo "  Fix: go install github.com/urmzd/zoro@latest  OR  just build && export PATH=\$PWD:\$PATH"
  errors=$((errors + 1))
fi

# Check Docker
if ! docker info &>/dev/null; then
  echo "FAIL: Docker is not running"
  echo "  Fix: start Docker Desktop or the Docker daemon"
  errors=$((errors + 1))
else
  # Check PostgreSQL container
  if ! docker compose ps --status running 2>/dev/null | grep -q postgres; then
    echo "FAIL: PostgreSQL container is not running"
    echo "  Fix: docker compose up -d postgres"
    errors=$((errors + 1))
  fi

  # Check SearXNG container (optional — zoro can manage its own subprocess)
  if ! docker compose ps --status running 2>/dev/null | grep -q searxng; then
    echo "WARN: SearXNG container is not running (zoro will try to start its own)"
  fi
fi

# Check Ollama
if ! curl -sf http://localhost:11434/api/tags &>/dev/null; then
  echo "FAIL: Ollama is not responding at localhost:11434"
  echo "  Fix: ollama serve"
  errors=$((errors + 1))
fi

if [ "$errors" -gt 0 ]; then
  echo ""
  echo "$errors preflight check(s) failed. Run 'just setup' to fix."
  exit 1
fi

echo "All preflight checks passed."
