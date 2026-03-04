"use client";

import { create } from "zustand";
import type { Fact, GraphData, NodeDetail } from "@/app/lib/types";

interface KnowledgeStore {
  graphData: GraphData | null;
  selectedNode: NodeDetail | null;
  searchResults: Fact[];
  searchQuery: string;
  isLoading: boolean;

  setGraphData: (data: GraphData) => void;
  setSelectedNode: (node: NodeDetail | null) => void;
  setSearchResults: (results: Fact[]) => void;
  setSearchQuery: (query: string) => void;
  setLoading: (loading: boolean) => void;
  highlightSubgraph: (nodeIds: string[]) => void;
  highlightedNodes: Set<string>;
}

export const useKnowledgeStore = create<KnowledgeStore>((set) => ({
  graphData: null,
  selectedNode: null,
  searchResults: [],
  searchQuery: "",
  isLoading: false,
  highlightedNodes: new Set<string>(),

  setGraphData: (data) => set({ graphData: data }),
  setSelectedNode: (node) => set({ selectedNode: node }),
  setSearchResults: (results) => set({ searchResults: results }),
  setSearchQuery: (query) => set({ searchQuery: query }),
  setLoading: (loading) => set({ isLoading: loading }),
  highlightSubgraph: (nodeIds) => set({ highlightedNodes: new Set(nodeIds) }),
}));
