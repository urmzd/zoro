package searcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/urmzd/zoro/internal/models"
)

const searxngURL = "http://127.0.0.1:8888"

type searxngResponse struct {
	Results []searxngResult `json:"results"`
}

type searxngResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type Searcher struct {
	http *http.Client
}

func New() *Searcher {
	return &Searcher{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *Searcher) Search(ctx context.Context, query string) ([]models.SearchResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searxngURL+"/search", nil)
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}

	q := req.URL.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("engines", "google,bing")
	req.URL.RawQuery = q.Encode()

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("searxng search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("searxng returned %d", resp.StatusCode)
	}

	var body searxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("parse searxng response: %w", err)
	}

	var results []models.SearchResult
	seenURLs := make(map[string]bool)

	for _, r := range body.Results {
		if r.URL == "" || seenURLs[r.URL] {
			continue
		}
		seenURLs[r.URL] = true

		results = append(results, models.SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
		})

		if len(results) >= 8 {
			break
		}
	}

	log.Printf("searxng returned %d results for %q", len(results), query)
	return results, nil
}
