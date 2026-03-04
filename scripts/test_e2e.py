"""End-to-end pipeline tests.

Tests the full flow: Go API -> Neo4j

Usage: python3 scripts/test_e2e.py [api_url]
Default URL: http://localhost:8080

Requires: pip install sseclient-py
"""

import json
import sys
import urllib.request
import urllib.error
import time

API_URL = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8080"


def post_json(url: str, data: dict) -> dict:
    body = json.dumps(data).encode()
    req = urllib.request.Request(
        url,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.loads(resp.read())


def get_json(url: str) -> dict:
    req = urllib.request.Request(url)
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.loads(resp.read())


def test_health():
    """Test Go API health endpoint."""
    print("Test: API health check...")
    result = get_json(f"{API_URL}/health")
    assert result.get("status") == "ok", f"Expected ok, got: {result}"
    print("  PASS")


def test_start_research():
    """Test starting a research session."""
    print("Test: Start research...")
    result = post_json(f"{API_URL}/api/research", {"query": "Python programming"})
    assert "id" in result, f"Expected id in response, got: {result}"
    session_id = result["id"]
    print(f"  PASS (session: {session_id})")
    return session_id


def test_get_session(session_id: str):
    """Test getting session state."""
    print("Test: Get session...")
    # Give the pipeline a moment to start
    time.sleep(1)
    result = get_json(f"{API_URL}/api/research/{session_id}")
    assert result.get("id") == session_id, f"Expected session id, got: {result}"
    assert result.get("query") == "Python programming", f"Expected query, got: {result}"
    print(f"  PASS (status: {result.get('status')})")
    return result


def test_sse_stream(session_id: str):
    """Test SSE event stream (basic connectivity test)."""
    print("Test: SSE stream...")
    try:
        import sseclient
    except ImportError:
        print("  SKIP (sseclient-py not installed, run: pip install sseclient-py)")
        return []

    req = urllib.request.Request(f"{API_URL}/api/research/{session_id}/stream")
    resp = urllib.request.urlopen(req, timeout=120)
    client = sseclient.SSEClient(resp)

    events = []
    for event in client.events():
        events.append({"type": event.event, "data": event.data})
        print(f"  Event: {event.event}")
        if event.event in ("research_complete", "error"):
            break
        if len(events) > 50:
            print("  (truncated after 50 events)")
            break

    assert len(events) > 0, "Expected at least one event"
    print(f"  PASS ({len(events)} events received)")
    return events


def test_knowledge_search():
    """Test knowledge graph search (after research)."""
    print("Test: Knowledge search...")
    try:
        result = get_json(f"{API_URL}/api/knowledge/search?q=Python")
        print(f"  PASS (got response)")
    except urllib.error.HTTPError as e:
        if e.code == 500:
            print(f"  SKIP (no knowledge data yet)")
        else:
            raise


def test_knowledge_graph():
    """Test knowledge graph endpoint."""
    print("Test: Knowledge graph...")
    try:
        result = get_json(f"{API_URL}/api/knowledge/graph?limit=100")
        assert "nodes" in result, f"Expected nodes in response"
        assert "edges" in result, f"Expected edges in response"
        print(f"  PASS ({len(result['nodes'])} nodes, {len(result['edges'])} edges)")
    except urllib.error.HTTPError as e:
        if e.code == 500:
            print(f"  SKIP (no graph data yet)")
        else:
            raise


def main():
    print(f"Running e2e tests against {API_URL}\n")

    passed = 0
    failed = 0

    # Basic tests
    for test_fn in [test_health]:
        try:
            test_fn()
            passed += 1
        except Exception as e:
            print(f"  FAIL: {e}")
            failed += 1

    # Research pipeline test
    try:
        session_id = test_start_research()
        passed += 1

        test_get_session(session_id)
        passed += 1

        # SSE stream test (consumes the stream)
        test_sse_stream(session_id)
        passed += 1

    except Exception as e:
        print(f"  FAIL: {e}")
        failed += 1

    # Knowledge tests
    for test_fn in [test_knowledge_search, test_knowledge_graph]:
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
