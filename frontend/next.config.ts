import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {},
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "cdn-city.rexio.pro",
      },
    ],
  },
  // Next.js Route Handlers have a default body size limit (1MB in Next.js 15+).
  // Mobile camera photos can be 10-15MB, so we need to raise this limit.
  // Set to 32MB to accommodate the PRD's 30MB max media file size with overhead.
  experimental: {
    // This configures the body size limit for Route Handlers and Server Actions
    serverActions: {
      bodySizeLimit: '32mb',
    },
  },
};

export default nextConfig;

