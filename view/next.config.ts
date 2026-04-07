import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'standalone',
  transpilePackages: [],
  basePath: process.env.BASE_PATH || '',
  assetPrefix: process.env.ASSET_PREFIX || undefined,
  env: {
    NEXT_PUBLIC_BASE_PATH: process.env.BASE_PATH || ''
  },
  images: {
    unoptimized: true
  }
};

export default nextConfig;
