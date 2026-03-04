package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/urmzd/zoro/api/internal/model"
)

type Searcher struct{}

func NewSearcher() *Searcher {
	return &Searcher{}
}

func (s *Searcher) Search(ctx context.Context, query string) ([]model.SearchResult, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	taskCtx, taskCancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer taskCancel()

	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&num=8", query)

	var results []model.SearchResult
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`#search`, chromedp.ByID),
		chromedp.Evaluate(`
			(() => {
				const results = [];
				document.querySelectorAll('div.g').forEach(el => {
					const titleEl = el.querySelector('h3');
					const linkEl = el.querySelector('a');
					const snippetEl = el.querySelector('[data-sncf], .VwiC3b, .IsZvec');
					if (titleEl && linkEl) {
						results.push({
							title: titleEl.innerText || '',
							url: linkEl.href || '',
							snippet: snippetEl ? snippetEl.innerText : ''
						});
					}
				});
				return results.slice(0, 8);
			})()
		`, &results),
	)

	if err != nil {
		log.Printf("chromedp search error: %v", err)
		return nil, fmt.Errorf("web search: %w", err)
	}

	return results, nil
}
