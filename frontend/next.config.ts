import type { NextConfig } from "next";

const isDesktop = process.env.NEXT_BUILD_TARGET === "desktop";

const nextConfig: NextConfig = {
  env: { NEXT_PUBLIC_DESKTOP: isDesktop ? "1" : "" },
  ...(isDesktop ? { output: "export", distDir: "out" } : {}),
  ...(!isDesktop
    ? {
        async rewrites() {
          return [
            {
              source: "/api/:path*",
              destination: "http://localhost:8080/api/:path*",
            },
          ];
        },
      }
    : {}),
};

export default nextConfig;
