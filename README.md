<p align="center">
  <h1 align="center">Zoro</h1>
  <p align="center"><strong>Archived</strong> — tools moved to <a href="https://github.com/urmzd/saige">saige</a>.</p>
</p>

All research tools (web search, file search, read file, knowledge graph CRUD, graph visualization) have been consolidated into [`saige/tools/research/`](https://github.com/urmzd/saige/tree/main/tools/research).

The SearXNG client is now at [`saige/rag/source/searxng/`](https://github.com/urmzd/saige/tree/main/rag/source/searxng).

Use `saige-mcp` to expose these tools over MCP/stdio to any agent (Claude Code, Gemini CLI, etc.):

```bash
go install github.com/urmzd/saige/cmd/saige-mcp@latest
saige-mcp --tools research --searxng-url http://localhost:8080
```

## License

[Apache License 2.0](LICENSE)
