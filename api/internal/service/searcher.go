package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/urmzd/zoro/api/internal/model"
)

type Searcher struct {
	baseURL string
	client  *http.Client
}

func NewSearcher(baseURL string) *Searcher {
	return &Searcher{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type searxngResponse struct {
	Results []searxngResult `json:"results"`
}

type searxngResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func (s *Searcher) Search(ctx context.Context, query string) ([]model.SearchResult, error) {
	u := fmt.Sprintf("%s/search?q=%s&format=json&engines=google,bing,duckduckgo",
		s.baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("web search: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("web search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("web search: searxng returned status %d", resp.StatusCode)
	}

	var sr searxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("web search: %w", err)
	}

	seen := make(map[string]bool)
	var results []model.SearchResult
	for _, r := range sr.Results {
		if seen[r.URL] || r.URL == "" {
			continue
		}
		seen[r.URL] = true
		results = append(results, model.SearchResult{
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
