package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	kgtypes "github.com/urmzd/saige/knowledge/types"
	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
)

func RunKnowledge(args []string) error {
	if len(args) == 0 {
		printKnowledgeUsage()
		os.Exit(1)
	}

	subcmd := args[0]
	rest := args[1:]

	switch subcmd {
	case "search":
		return runKnowledgeSearch(rest)
	case "store":
		return runKnowledgeStore(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown knowledge subcommand: %s\n\n", subcmd)
		printKnowledgeUsage()
		os.Exit(1)
	}
	return nil
}

func runKnowledgeSearch(args []string) error {
	fs := flag.NewFlagSet("knowledge search", flag.ExitOnError)
	limit := fs.Int("limit", 10, "max number of facts to return")
	groupID := fs.String("group", "", "filter by group ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	query := strings.Join(fs.Args(), " ")
	if query == "" {
		fmt.Fprintln(os.Stderr, "usage: zoro knowledge search [-limit N] [-group ID] QUERY...")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()
	c, err := app.WireComponents(ctx, cfg, app.WireOpts{NeedKnowledgeRW: true})
	if err != nil {
		return err
	}
	defer c.Cleanup()

	opts := []kgtypes.SearchOption{kgtypes.WithLimit(*limit)}
	if *groupID != "" {
		opts = append(opts, kgtypes.WithGroupID(*groupID))
	}

	resp, err := c.Graph.SearchFacts(ctx, query, opts...)
	if err != nil {
		return err
	}

	if len(resp.Facts) == 0 {
		fmt.Println("No relevant knowledge found.")
		return nil
	}

	for _, f := range resp.Facts {
		fmt.Printf("- %s -> %s: %s\n", f.SourceNode.Name, f.TargetNode.Name, f.FactText)
	}
	return nil
}

func runKnowledgeStore(args []string) error {
	fs := flag.NewFlagSet("knowledge store", flag.ExitOnError)
	source := fs.String("source", "cli", "source description for the knowledge")
	groupID := fs.String("group", "", "group ID to associate with")
	if err := fs.Parse(args); err != nil {
		return err
	}

	text := strings.Join(fs.Args(), " ")
	if text == "" {
		fmt.Fprintln(os.Stderr, "usage: zoro knowledge store [-source NAME] [-group ID] TEXT...")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()
	c, err := app.WireComponents(ctx, cfg, app.WireOpts{NeedKnowledgeRW: true})
	if err != nil {
		return err
	}
	defer c.Cleanup()

	input := &kgtypes.EpisodeInput{
		Name:    *source,
		Body:    text,
		Source:  *source,
		GroupID: *groupID,
	}

	resp, err := c.Graph.IngestEpisode(ctx, input)
	if err != nil {
		return err
	}

	fmt.Printf("Stored %d entities and %d relations.\n", len(resp.EntityNodes), len(resp.EpisodicEdges))
	return nil
}

func printKnowledgeUsage() {
	fmt.Fprintf(os.Stderr, `Usage: zoro knowledge <command> [flags]

Commands:
  search    Search the knowledge graph for facts
  store     Store text into the knowledge graph
`)
}
