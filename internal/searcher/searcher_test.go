package searcher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "test query" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("format") != "json" {
			t.Errorf("expected format=json, got %s", r.URL.Query().Get("format"))
		}

		resp := searxngResponse{
			Results: []searxngResult{
				{Title: "Result 1", URL: "https://example.com/1", Content: "Snippet 1"},
				{Title: "Result 2", URL: "https://example.com/2", Content: "Snippet 2"},
				{Title: "Result 3", URL: "https://example.com/3", Content: "Snippet 3"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	s := New(srv.URL)
	results, err := s.Search(context.Background(), "test query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Title != "Result 1" {
		t.Errorf("results[0].Title = %q, want %q", results[0].Title, "Result 1")
	}
	if results[1].URL != "https://example.com/2" {
		t.Errorf("results[1].URL = %q", results[1].URL)
	}
}

func TestSearch_DeduplicatesURLs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := searxngResponse{
			Results: []searxngResult{
				{Title: "First", URL: "https://example.com/same", Content: "A"},
				{Title: "Duplicate", URL: "https://example.com/same", Content: "B"},
				{Title: "Different", URL: "https://example.com/other", Content: "C"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	s := New(srv.URL)
	results, err := s.Search(context.Background(), "dedup test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results after dedup, got %d", len(results))
	}
}

func TestSearch_SkipsEmptyURLs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := searxngResponse{
			Results: []searxngResult{
				{Title: "No URL", URL: "", Content: "A"},
				{Title: "Has URL", URL: "https://example.com", Content: "B"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	s := New(srv.URL)
	results, err := s.Search(context.Background(), "empty url test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearch_LimitsTo8Results(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var results []searxngResult
		for i := 0; i < 15; i++ {
			results = append(results, searxngResult{
				Title:   "Result",
				URL:     "https://example.com/" + string(rune('a'+i)),
				Content: "Content",
			})
		}
		json.NewEncoder(w).Encode(searxngResponse{Results: results})
	}))
	defer srv.Close()

	s := New(srv.URL)
	results, err := s.Search(context.Background(), "many results")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 8 {
		t.Fatalf("expected 8 results (limit), got %d", len(results))
	}
}

func TestSearch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := New(srv.URL)
	_, err := s.Search(context.Background(), "error test")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestSearch_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	s := New(srv.URL)
	_, err := s.Search(context.Background(), "bad json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSearch_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(searxngResponse{})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := New(srv.URL)
	_, err := s.Search(ctx, "cancelled")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
