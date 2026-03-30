package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
)

func runChat(args []string) error {
	fs := flag.NewFlagSet("chat", flag.ExitOnError)
	sessionID := fs.String("s", "", "session ID to continue")
	if err := fs.Parse(args); err != nil {
		return err
	}

	message := strings.Join(fs.Args(), " ")
	if message == "" {
		fmt.Fprintln(os.Stderr, "usage: zoro chat [-s SESSION] MESSAGE...")
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

	c, err := app.WireComponents(ctx, cfg, app.WireOpts{NeedAgent: true})
	if err != nil {
		return err
	}
	defer c.Cleanup()

	response, returnedID, err := c.Agent.InvokeSync(ctx, *sessionID, message)
	if err != nil {
		return err
	}

	fmt.Println(response)
	fmt.Fprintf(os.Stderr, "\nsession: %s\n", returnedID)
	return nil
}
