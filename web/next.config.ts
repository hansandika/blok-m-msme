import type { NextConfig } from "next";

const upstream = (process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080").replace(/\/$/, "");

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    // Browser calls /backend/* on the same origin; Next proxies to the Go API.
    return [{ source: "/backend/:path*", destination: `${upstream}/:path*` }];
  },
};

export default nextConfig;
