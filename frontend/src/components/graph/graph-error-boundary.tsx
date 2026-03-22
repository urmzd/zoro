"use client";

import { Component, type ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

export class GraphErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex h-full flex-col items-center justify-center gap-2 p-8 text-center">
          <p className="text-sm text-muted-foreground">Failed to load knowledge graph</p>
          <p className="text-xs text-destructive/70 font-mono max-w-md break-all">
            {this.state.error.message}
          </p>
          <button
            type="button"
            onClick={() => this.setState({ error: null })}
            className="mt-2 px-3 py-1.5 text-xs rounded-md bg-muted hover:bg-muted/80 text-foreground"
          >
            Retry
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
