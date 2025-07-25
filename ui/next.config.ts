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
      {
        source: '/users/:path*',
        destination: `${process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'}/users/:path*`,
      },
        {
            source: '/secrets/:path*',
            destination: `${process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'}/secrets/:path*`,
        },
        {
            source: '/roles/:path*',
            destination: `${process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'}/roles/:path*`,
        },
    ]
  },
};

export default nextConfig;
