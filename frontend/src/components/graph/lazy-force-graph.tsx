"use client";

import { type ComponentType, Suspense, useEffect, useState } from "react";

// biome-ignore lint/suspicious/noExplicitAny: wraps dynamic third-party component
let ForceGraph2DLazy: ComponentType<any> | null = null;

// biome-ignore lint/suspicious/noExplicitAny: wraps dynamic third-party component
export function LazyForceGraph2D(props: any) {
  // biome-ignore lint/suspicious/noExplicitAny: wraps dynamic third-party component
  const [Component, setComponent] = useState<ComponentType<any> | null>(null);

  useEffect(() => {
    if (ForceGraph2DLazy) {
      setComponent(() => ForceGraph2DLazy);
      return;
    }
    import("react-force-graph-2d")
      .then((mod) => {
        ForceGraph2DLazy = mod.default || mod;
        setComponent(() => ForceGraph2DLazy);
      })
      .catch((err) => {
        console.error("Failed to load ForceGraph2D:", err);
      });
  }, []);

  if (!Component) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Loading graph...
      </div>
    );
  }

  return (
    <Suspense
      fallback={
        <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
          Loading graph...
        </div>
      }
    >
      <Component {...props} />
    </Suspense>
  );
}
