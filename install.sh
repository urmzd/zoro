#!/usr/bin/env bash
set -euo pipefail

echo "=== Zoro Setup ==="

# Check prerequisites
for cmd in go ollama docker; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "ERROR: $cmd not found."
        case "$cmd" in
            go)      echo "  Install Go: https://go.dev/dl/" ;;
            ollama)  echo "  Install Ollama: https://ollama.ai" ;;
            docker)  echo "  Install Docker: https://docs.docker.com/get-docker/" ;;
        esac
        exit 1
    fi
done

if ! command -v just &> /dev/null; then
    echo "Installing just..."
    command -v brew &> /dev/null && brew install just || go install github.com/casey/just@latest
fi

just setup
echo ""
echo "Run 'just dev' to start."
