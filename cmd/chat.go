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

func RunChat(args []string) error {
	fs := flag.NewFlagSet("chat", flag.ExitOnError)
	sessionID := fs.String("s", "", "session ID to continue")
	jsonOut := fs.Bool("json", false, "output as JSON (includes session_id)")
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

	if *jsonOut {
		out, _ := json.Marshal(map[string]string{
			"response":   response,
			"session_id": returnedID,
		})
		fmt.Println(string(out))
	} else {
		fmt.Println(response)
		fmt.Fprintf(os.Stderr, "\nsession: %s\n", returnedID)
	}
	return nil
}
