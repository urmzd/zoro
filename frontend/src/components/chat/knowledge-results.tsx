interface KnowledgeFact {
  uuid?: string;
  name?: string;
  fact?: string;
  source_node?: { name?: string };
  target_node?: { name?: string };
}

interface KnowledgeResultsProps {
  results: KnowledgeFact[];
  searched: boolean;
  loading: boolean;
}

export function KnowledgeResults({ results, searched, loading }: KnowledgeResultsProps) {
  if (!searched || loading) return null;

  return (
    <div className="space-y-2 w-full">
      {results.length === 0 ? (
        <p className="text-center text-sm text-muted-foreground py-4">
          No results found. Try a different query or build knowledge through chat first.
        </p>
      ) : (
        <div className="space-y-2 max-h-[40vh] overflow-y-auto">
          {results.map((fact, i) => (
            <div
              key={fact.uuid ?? i}
              className="rounded-lg border border-border/50 bg-card/50 px-4 py-3 text-sm"
            >
              {(fact.source_node?.name || fact.target_node?.name) && (
                <div className="flex items-center gap-2 mb-1">
                  {fact.source_node?.name && (
                    <span className="rounded-full bg-indigo-600/20 px-2 py-0.5 text-xs font-medium text-indigo-400">
                      {fact.source_node.name}
                    </span>
                  )}
                  {fact.source_node?.name && fact.target_node?.name && (
                    <span className="text-muted-foreground text-xs">&rarr;</span>
                  )}
                  {fact.target_node?.name && (
                    <span className="rounded-full bg-purple-600/20 px-2 py-0.5 text-xs font-medium text-purple-400">
                      {fact.target_node.name}
                    </span>
                  )}
                </div>
              )}
              <p className="text-foreground/80">{fact.fact}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
