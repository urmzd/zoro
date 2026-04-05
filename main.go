package main

import (
	"fmt"
	"os"

	"github.com/urmzd/zoro/cmd"
)

var version = "dev"

func main() {
	cmd.Version = version

	subcmd := "serve"
	args := os.Args[1:]
	if len(args) > 0 {
		subcmd = args[0]
		args = args[1:]
	}

	var err error
	switch subcmd {
	case "serve":
		err = cmd.RunServe()
	case "chat":
		err = cmd.RunChat(args)
	case "research":
		err = cmd.RunResearch(args)
	case "search":
		err = cmd.RunSearch(args)
	case "graph":
		err = cmd.RunGraph(args)
	case "knowledge":
		err = cmd.RunKnowledge(args)
	case "version":
		cmd.RunVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: zoro <command> [flags]

Commands:
  serve       Start MCP server on stdio (default)
  chat        Chat with Zoro
  research    Run deep research pipeline
  search      Search the web
  knowledge   Search or store knowledge graph entries
  graph       Visualize the knowledge graph
  version     Print version
  help        Show this help
`)
}
