package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	cmd := "serve"
	args := os.Args[1:]
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	var err error
	switch cmd {
	case "serve":
		err = runServe()
	case "chat":
		err = runChat(args)
	case "research":
		err = runResearch(args)
	case "search":
		err = runSearch(args)
	case "version":
		runVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
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
  version     Print version
  help        Show this help
`)
}
