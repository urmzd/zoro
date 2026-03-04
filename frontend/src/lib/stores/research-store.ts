"use client";

import { create } from "zustand";
import type {
  Entity,
  Fact,
  GraphData,
  Relation,
  ResearchState,
  SearchResult,
  SSEEventType,
  TimelineEvent,
} from "@/app/lib/types";

interface ResearchStore extends ResearchState {
  connect: (query: string) => void;
  addTimelineEvent: (type: SSEEventType, message: string) => void;
  setStatus: (status: ResearchState["status"]) => void;
  setSearchResults: (results: SearchResult[]) => void;
  addEntity: (entity: Entity) => void;
  addRelation: (relation: Relation) => void;
  setPriorFacts: (facts: Fact[]) => void;
  setGraphData: (data: GraphData) => void;
  appendSummary: (token: string) => void;
  setError: (error: string) => void;
  reset: () => void;
}

const initialState: ResearchState = {
  status: "idle",
  query: "",
  searchResults: [],
  entities: [],
  relations: [],
  priorFacts: [],
  graphData: null,
  summary: "",
  timeline: [],
  error: null,
};

export const useResearchStore = create<ResearchStore>((set) => ({
  ...initialState,

  connect: (query) => set({ ...initialState, status: "connecting", query }),

  addTimelineEvent: (type, message) =>
    set((s) => ({
      timeline: [
        ...s.timeline,
        { type, message, timestamp: new Date().toISOString() } as TimelineEvent,
      ],
    })),

  setStatus: (status) => set({ status }),

  setSearchResults: (results) => set({ searchResults: results }),

  addEntity: (entity) => set((s) => ({ entities: [...s.entities, entity] })),

  addRelation: (relation) => set((s) => ({ relations: [...s.relations, relation] })),

  setPriorFacts: (facts) => set({ priorFacts: facts }),

  setGraphData: (data) => set({ graphData: data }),

  appendSummary: (token) => set((s) => ({ summary: s.summary + token })),

  setError: (error) => set({ status: "error", error }),

  reset: () => set(initialState),
}));
