package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
)

func RunSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	query := strings.Join(fs.Args(), " ")
	if query == "" {
		fmt.Fprintln(os.Stderr, "usage: zoro search [-json] QUERY...")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	cfg := config.Load()
	if cfg.SearXNGURL == "" {
		cfg.SearXNGURL = "http://127.0.0.1:8888"
	}

	c, err := app.WireComponents(ctx, cfg, app.WireOpts{NeedSearcher: true})
	if err != nil {
		return err
	}
	defer c.Cleanup()

	results, err := c.Searcher.Search(ctx, query)
	if err != nil {
		return err
	}

	if *jsonOut {
		out, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	} else {
		for i, r := range results {
			fmt.Printf("%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet)
		}
	}
	return nil
}
