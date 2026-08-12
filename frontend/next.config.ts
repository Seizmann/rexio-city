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
};

export default nextConfig;

