#!/usr/bin/env bash
set -euo pipefail

echo "=== Zoro Setup ==="

# Check prerequisites
for cmd in rustc cargo node ollama; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "ERROR: $cmd not found."
        case "$cmd" in
            rustc|cargo) echo "  Install Rust: https://rustup.rs" ;;
            node)        echo "  Install Node.js 24+: https://nodejs.org" ;;
            ollama)      echo "  Install Ollama: https://ollama.ai" ;;
        esac
        exit 1
    fi
done

if ! command -v just &> /dev/null; then
    echo "Installing just..."
    command -v brew &> /dev/null && brew install just || cargo install just
fi

just setup
echo ""
echo "Run 'just dev' to start."
