#!/usr/bin/env python3
"""Benchmark LLM models for generation TPS via Ollama."""

import json
import subprocess
import sys
import urllib.request
import urllib.error

OLLAMA_URL = "http://localhost:11434"

MODELS = [
    "gemma3:latest",
    "gemma4:latest",
]

PROMPT = (
    "Complete the following Python function. "
    "Only output the missing code, nothing else.\n\n"
    "def fibonacci(n):\n"
    "    if n <= 1:\n"
    "        return n\n"
    "    "
)


def ollama_running() -> bool:
    try:
        urllib.request.urlopen(f"{OLLAMA_URL}/api/tags", timeout=3)
        return True
    except Exception:
        return False


def pull(model: str) -> None:
    print(f"  Pulling {model}...")
    subprocess.run(["ollama", "pull", model], check=True, capture_output=True)


def ensure_model(model: str) -> None:
    req = urllib.request.Request(f"{OLLAMA_URL}/api/tags")
    with urllib.request.urlopen(req, timeout=10) as resp:
        data = json.loads(resp.read())
    names = [m["name"] for m in data.get("models", [])]
    if model not in names and not any(n.startswith(model) for n in names):
        pull(model)


def generate(model: str) -> dict:
    payload = json.dumps({
        "model": model,
        "prompt": PROMPT,
        "stream": False,
        "options": {"num_predict": 128, "temperature": 0},
    }).encode()
    req = urllib.request.Request(
        f"{OLLAMA_URL}/api/generate",
        data=payload,
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        return json.loads(resp.read())


def bench_model(model: str) -> dict | None:
    print(f"\n{'=' * 50}")
    print(f"  Model: {model}")
    print(f"{'=' * 50}")

    try:
        ensure_model(model)
    except Exception as e:
        print(f"  Error ensuring model: {e}")
        return None

    # Warm up (load model into memory)
    print("  Warming up...")
    try:
        generate(model)
    except Exception as e:
        print(f"  Warmup failed: {e}")
        return None

    # Actual benchmark
    print("  Benchmarking...")
    try:
        resp = generate(model)
    except Exception as e:
        print(f"  Benchmark failed: {e}")
        return None

    eval_count = resp.get("eval_count", 0)
    eval_duration = resp.get("eval_duration", 1)
    prompt_eval_count = resp.get("prompt_eval_count", 0)
    prompt_eval_duration = resp.get("prompt_eval_duration", 1)
    total_duration = resp.get("total_duration", 0)
    response_text = resp.get("response", "").strip()

    gen_tps = eval_count / (eval_duration / 1e9) if eval_duration else 0
    prompt_tps = prompt_eval_count / (prompt_eval_duration / 1e9) if prompt_eval_duration else 0
    total_s = total_duration / 1e9

    print(f"  Tokens generated: {eval_count}")
    print(f"  Generation TPS:   {gen_tps:.1f} tok/s")
    print(f"  Prompt tokens:    {prompt_eval_count}")
    print(f"  Prompt TPS:       {prompt_tps:.1f} tok/s")
    print(f"  Total time:       {total_s:.2f}s")
    print(f"  ---")
    print(f"  {response_text}")

    return {
        "model": model,
        "gen_tps": gen_tps,
        "prompt_tps": prompt_tps,
        "eval_count": eval_count,
        "total_s": total_s,
    }


def main():
    models = sys.argv[1:] if len(sys.argv) > 1 else MODELS

    if not ollama_running():
        print("Error: ollama is not running. Start it with: ollama serve")
        sys.exit(1)

    results = []
    for model in models:
        result = bench_model(model)
        if result:
            results.append(result)

    if results:
        print(f"\n{'=' * 50}")
        print("  SUMMARY")
        print(f"{'=' * 50}")
        print(f"  {'Model':<20} {'Gen TPS':>10} {'Prompt TPS':>12} {'Total':>8}")
        print(f"  {'-'*20} {'-'*10} {'-'*12} {'-'*8}")
        for r in results:
            print(
                f"  {r['model']:<20} {r['gen_tps']:>9.1f} {r['prompt_tps']:>11.1f} {r['total_s']:>7.2f}s"
            )


if __name__ == "__main__":
    main()
