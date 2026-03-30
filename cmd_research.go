package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
)

func runResearch(args []string) error {
	query := strings.Join(args, " ")
	if query == "" {
		fmt.Fprintln(os.Stderr, "usage: zoro research QUERY...")
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

	c, err := app.WireComponents(ctx, cfg, app.WireOpts{NeedOrchestrator: true})
	if err != nil {
		return err
	}
	defer c.Cleanup()

	summary, err := c.Orchestrator.RunSync(ctx, query)
	if err != nil {
		return err
	}

	fmt.Println(summary)
	return nil
}
