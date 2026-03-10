"use client";

import { IconSettings } from "@tabler/icons-react";

export function SidebarSettings() {
  return (
    <div className="px-2 py-1 pb-2">
      <button
        type="button"
        className="flex items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground hover:text-foreground hover:bg-muted/50 w-full transition-colors"
        onClick={() => {
          // Future: open settings modal
        }}
      >
        <IconSettings className="h-4 w-4" />
        Settings
      </button>
    </div>
  );
}
