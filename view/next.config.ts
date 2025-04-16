import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'standalone',
  env: {
    PORT: process.env.NEXT_PUBLIC_PORT || '7443'
  }
};

export default nextConfig;
