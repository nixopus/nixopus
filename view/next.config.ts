import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'standalone',
  basePath: process.env.BASE_PATH || '',
  assetPrefix: process.env.ASSET_PREFIX || undefined,
  images: {
    unoptimized: true
  }
};

export default nextConfig;
