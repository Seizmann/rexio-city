import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Force webpack instead of Turbopack for Vercel compatibility
  turbopack: {},
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "cdn-city.rexio.pro",
      },
    ],
  },
};

export default nextConfig;
