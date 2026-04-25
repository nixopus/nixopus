import type { NextConfig } from 'next';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const nextConfigDir = path.dirname(fileURLToPath(import.meta.url));
// Repo root: avoids "multiple lockfiles" warning when nixopus/ and view/ both have lockfiles.
const monorepoRoot = path.join(nextConfigDir, '..');

const nextConfig: NextConfig = {
  outputFileTracingRoot: monorepoRoot,
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
