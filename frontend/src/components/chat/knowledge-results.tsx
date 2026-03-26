import type { Fact } from "@/app/lib/types";

interface KnowledgeResultsProps {
  results: Fact[];
  searched: boolean;
  loading: boolean;
}

export function KnowledgeResults({ results, searched, loading }: KnowledgeResultsProps) {
  if (!searched || loading) return null;

  if (results.length === 0) {
    return (
      <p className="text-center text-sm text-[#a3aac4] py-4">
        No results found. Try a different query or build knowledge through chat first.
      </p>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between px-4">
        <h3 className="font-headline text-2xl font-semibold flex items-center gap-3">
          <span className="material-symbols-outlined text-[#ba9eff]">hub</span>
          Graph Relationships
        </h3>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {results.map((fact) => (
          <div
            key={fact.uuid}
            className="glass-panel p-6 rounded-2xl border border-[#40485d]/15 hover:scale-[1.02] transition-all duration-300"
          >
            <div className="flex flex-col items-center text-center space-y-4">
              {fact.source_node?.name && (
                <div className="px-3 py-1 bg-[#ba9eff]/10 rounded-lg text-[#ba9eff] text-sm font-bold">
                  {fact.source_node.name}
                </div>
              )}
              <div className="flex items-center gap-2 text-[#a3aac4]">
                <span className="material-symbols-outlined text-sm">arrow_downward</span>
                <span className="text-xs font-mono uppercase tracking-widest">
                  {fact.fact?.split(" ").slice(0, 3).join(" ") || "relates to"}
                </span>
              </div>
              {fact.target_node?.name && (
                <div className="px-3 py-1 bg-[#141f38] rounded-lg text-sm">
                  {fact.target_node.name}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      {results.length > 0 && results[0].fact && (
        <div className="glass-panel p-8 rounded-3xl border border-[#40485d]/15 relative overflow-hidden">
          <div className="absolute top-0 right-0 p-4">
            <span className="material-symbols-outlined text-[#a3aac4]/20 text-6xl">
              auto_awesome
            </span>
          </div>
          <div className="max-w-2xl">
            <h4 className="font-headline text-xl font-bold mb-4 flex items-center gap-2">
              <span className="material-symbols-outlined text-[#ff716a]">verified</span>
              Core Verification
            </h4>
            <p className="text-[#a3aac4] leading-relaxed">{results[0].fact}</p>
          </div>
        </div>
      )}
    </div>
  );
}
