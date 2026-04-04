package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	kgtypes "github.com/urmzd/saige/knowledge/types"
	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
	"github.com/urmzd/zoro/internal/graph"
)

func RunGraph(args []string) error {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	format := fs.String("format", "text", "output format: text, dot, json")
	limit := fs.Int64("limit", 100, "max number of edges to return")
	nodeID := fs.String("node", "", "show neighborhood of a specific entity UUID")
	depth := fs.Int("depth", 2, "traversal depth when using -node")
	if err := fs.Parse(args); err != nil {
		return err
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
	c, err := app.WireComponents(ctx, cfg, app.WireOpts{NeedGraph: true})
	if err != nil {
		return err
	}
	defer c.Cleanup()

	var data *kgtypes.GraphData

	if *nodeID != "" {
		detail, err := c.Graph.GetNode(ctx, *nodeID, *depth)
		if err != nil {
			return err
		}
		data = nodeDetailToGraphData(detail)
	} else {
		data, err = c.Graph.GetGraph(ctx, *limit)
		if err != nil {
			return err
		}
	}

	switch *format {
	case "dot":
		fmt.Print(graph.ToDOT(data))
	case "json":
		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	default:
		fmt.Print(graph.ToText(data))
	}
	return nil
}

func nodeDetailToGraphData(d *kgtypes.NodeDetail) *kgtypes.GraphData {
	nodes := make([]kgtypes.GraphNode, 0, 1+len(d.Neighbors))
	nodes = append(nodes, d.Node)
	nodes = append(nodes, d.Neighbors...)
	return &kgtypes.GraphData{
		Nodes: nodes,
		Edges: d.Edges,
	}
}
