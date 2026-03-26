"use client";

import { motion } from "framer-motion";
import type { NodeDetail } from "@/app/lib/types";

interface NodeDetailPanelProps {
  detail: NodeDetail;
  onClose: () => void;
}

const TYPE_COLORS: Record<string, string> = {
  entity: "bg-[#ba9eff]/10 text-[#ba9eff]",
  person: "bg-[#699cff]/10 text-[#699cff]",
  organization: "bg-[#a27cff]/10 text-[#a27cff]",
  concept: "bg-emerald-500/10 text-emerald-400",
  location: "bg-[#ff716a]/10 text-[#ff716a]",
};

export function NodeDetailPanel({ detail, onClose }: NodeDetailPanelProps) {
  const typeColor = TYPE_COLORS[detail.node.type] || TYPE_COLORS.entity;

  return (
    <motion.div
      initial={{ x: "100%" }}
      animate={{ x: 0 }}
      exit={{ x: "100%" }}
      transition={{ type: "spring", damping: 25, stiffness: 300 }}
      className="absolute right-6 top-6 bottom-6 z-20 w-80 bg-[#141f38]/90 backdrop-blur-xl border border-[#40485d]/30 rounded-2xl shadow-2xl p-6 flex flex-col overflow-hidden"
    >
      <div className="flex justify-between items-start mb-6">
        <div className="flex flex-col gap-1">
          <span className={`text-[10px] font-bold uppercase tracking-tighter px-2 py-0.5 rounded self-start ${typeColor}`}>
            {detail.node.type}
          </span>
          <h4 className="text-xl font-headline font-bold">{detail.node.name}</h4>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="p-1 hover:bg-white/5 rounded-full"
        >
          <span className="material-symbols-outlined text-sm">close</span>
        </button>
      </div>

      <div className="flex-1 space-y-6 overflow-y-auto pr-2">
        {detail.node.summary && (
          <div>
            <label className="text-[10px] text-[#a3aac4] font-bold uppercase tracking-widest block mb-2">
              Summary
            </label>
            <p className="text-sm text-[#dee5ff]/80 leading-relaxed">
              {detail.node.summary}
            </p>
          </div>
        )}

        {detail.edges.length > 0 && (
          <div>
            <label className="text-[10px] text-[#a3aac4] font-bold uppercase tracking-widest block mb-3">
              Relations ({detail.edges.length})
            </label>
            <div className="space-y-2">
              {detail.edges.map((edge) => (
                <div
                  key={edge.id}
                  className="p-3 bg-[#0f1930] rounded-lg border border-[#40485d]/10"
                >
                  <span className="text-[#ba9eff] text-xs font-bold">{edge.type}</span>
                  {edge.fact && (
                    <p className="text-xs text-[#a3aac4] mt-1">{edge.fact}</p>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {detail.neighbors.length > 0 && (
          <div>
            <label className="text-[10px] text-[#a3aac4] font-bold uppercase tracking-widest block mb-3">
              Strong Neighbors
            </label>
            <div className="space-y-2">
              {detail.neighbors.map((n) => {
                const nColor = TYPE_COLORS[n.type] || TYPE_COLORS.entity;
                const dotColor = nColor.includes("ba9eff") ? "bg-[#ba9eff]" :
                  nColor.includes("699cff") ? "bg-[#699cff]" :
                  nColor.includes("ff716a") ? "bg-[#ff716a]" : "bg-[#ba9eff]";
                return (
                  <div
                    key={n.id}
                    className="flex items-center gap-3 p-2 bg-[#0f1930] rounded-lg border border-[#40485d]/10 cursor-pointer hover:bg-[#1f2b49] transition-colors"
                  >
                    <div className={`w-2 h-2 rounded-full ${dotColor}`} />
                    <span className="text-xs font-medium">{n.name}</span>
                    <span className="text-[10px] text-[#a3aac4] ml-auto">{n.type}</span>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </motion.div>
  );
}
