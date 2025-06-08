import type {NextConfig} from "next";

const nextConfig: NextConfig = {
  /* config options here */
  productionBrowserSourceMaps: true,
  async rewrites() {
    return [
      {
        source: '/key_management/:path*',
        destination: `${process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'}/key_management/:path*`,
      },
    ]
  },
};

export default nextConfig;
