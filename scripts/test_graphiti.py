"""Graphiti REST integration tests.

Usage: python3 scripts/test_graphiti.py [graphiti_url]
Default URL: http://localhost:8000
"""

import json
import sys
import urllib.request
import urllib.error
import time

BASE_URL = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8000"
GROUP_ID = f"test-{int(time.time())}"


def post(path: str, data: dict) -> dict:
    body = json.dumps(data).encode()
    req = urllib.request.Request(
        f"{BASE_URL}{path}",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        print(f"  HTTP {e.code}: {e.read().decode()}")
        raise


def test_add_episode():
    """Test adding an episode to Graphiti."""
    print("Test: Add episode...")
    result = post("/episodes", {
        "name": "Test Episode",
        "episode_body": "Python is a popular programming language created by Guido van Rossum. It was first released in 1991.",
        "source_description": "test",
        "group_id": GROUP_ID,
    })
    assert "uuid" in result or "episode_uuid" in result, f"Expected uuid in response, got: {result}"
    print("  PASS")
    return result


def test_search_facts():
    """Test searching facts after ingestion."""
    print("Test: Search facts...")
    result = post("/search", {
        "query": "Python programming language",
        "group_id": GROUP_ID,
        "num": 10,
    })
    assert isinstance(result, (dict, list)), f"Expected dict or list, got: {type(result)}"
    print(f"  PASS (got {len(result.get('facts', result) if isinstance(result, dict) else result)} results)")
    return result


def test_search_nodes():
    """Test searching nodes."""
    print("Test: Search nodes...")
    result = post("/search/nodes", {
        "query": "Python",
        "num": 10,
    })
    assert isinstance(result, (dict, list)), f"Expected dict or list, got: {type(result)}"
    print(f"  PASS")
    return result


def main():
    print(f"Running Graphiti integration tests against {BASE_URL}\n")

    # Check health
    try:
        req = urllib.request.Request(f"{BASE_URL}/docs")
        with urllib.request.urlopen(req, timeout=10) as resp:
            print(f"Graphiti is up (status {resp.status})\n")
    except Exception as e:
        print(f"ERROR: Graphiti is not reachable at {BASE_URL}: {e}")
        sys.exit(1)

    passed = 0
    failed = 0

    for test_fn in [test_add_episode, test_search_facts, test_search_nodes]:
        try:
            test_fn()
            passed += 1
        except Exception as e:
            print(f"  FAIL: {e}")
            failed += 1

    print(f"\nResults: {passed} passed, {failed} failed")
    sys.exit(1 if failed > 0 else 0)


if __name__ == "__main__":
    main()
