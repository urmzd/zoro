import type { NextConfig } from "next";

const isDesktop = process.env.NEXT_BUILD_TARGET === "desktop";

const nextConfig: NextConfig = {
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
