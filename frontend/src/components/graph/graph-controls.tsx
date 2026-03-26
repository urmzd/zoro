"use client";

interface GraphControlsProps {
  onZoomIn: () => void;
  onZoomOut: () => void;
  onReset: () => void;
}

export function GraphControls({ onZoomIn, onZoomOut, onReset }: GraphControlsProps) {
  return (
    <div className="absolute bottom-6 left-6 flex flex-col gap-2 z-10">
      <button
        type="button"
        onClick={onZoomIn}
        className="w-10 h-10 flex items-center justify-center bg-[#050a18] border border-[#40485d]/20 rounded-lg hover:bg-[#1f2b49] transition-colors"
        title="Zoom in"
      >
        <span className="material-symbols-outlined">add</span>
      </button>
      <button
        type="button"
        onClick={onZoomOut}
        className="w-10 h-10 flex items-center justify-center bg-[#050a18] border border-[#40485d]/20 rounded-lg hover:bg-[#1f2b49] transition-colors"
        title="Zoom out"
      >
        <span className="material-symbols-outlined">remove</span>
      </button>
      <button
        type="button"
        onClick={onReset}
        className="w-10 h-10 flex items-center justify-center bg-[#050a18] border border-[#40485d]/20 rounded-lg hover:bg-[#1f2b49] transition-colors"
        title="Fit to view"
      >
        <span className="material-symbols-outlined">recenter</span>
      </button>
    </div>
  );
}
