"use client";

import { IconZoomIn, IconZoomOut, IconFocusCentered } from "@tabler/icons-react";

interface GraphControlsProps {
  onZoomIn: () => void;
  onZoomOut: () => void;
  onReset: () => void;
}

export function GraphControls({ onZoomIn, onZoomOut, onReset }: GraphControlsProps) {
  return (
    <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-10 flex items-center gap-1 rounded-lg border border-border/50 bg-background/80 backdrop-blur-md p-1">
      <button
        onClick={onZoomIn}
        className="rounded-md p-2 hover:bg-muted transition-colors"
        title="Zoom in"
      >
        <IconZoomIn className="h-4 w-4" />
      </button>
      <button
        onClick={onZoomOut}
        className="rounded-md p-2 hover:bg-muted transition-colors"
        title="Zoom out"
      >
        <IconZoomOut className="h-4 w-4" />
      </button>
      <div className="w-px h-4 bg-border" />
      <button
        onClick={onReset}
        className="rounded-md p-2 hover:bg-muted transition-colors"
        title="Fit to view"
      >
        <IconFocusCentered className="h-4 w-4" />
      </button>
    </div>
  );
}
