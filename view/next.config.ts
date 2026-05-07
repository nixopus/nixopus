import type { NextConfig } from 'next';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const nextConfigDir = path.dirname(fileURLToPath(import.meta.url));

// Applied to every response served by Next.js (HTML pages, API routes, static assets).
// Note: HSTS is intentionally omitted — Caddy handles TLS termination and sets
// Strict-Transport-Security at the edge, matching the Go API middleware decision.
const securityHeaders = [
  // Prevents MIME-type sniffing; stops browsers from interpreting files as a
  // different MIME type than declared by the server.
  { key: 'X-Content-Type-Options', value: 'nosniff' },
  // Blocks the page from being embedded in <iframe>/<frame>/<object> on other
  // origins, defending against clickjacking. CSP frame-ancestors also covers
  // this for modern browsers; both co-exist for full browser compatibility.
  { key: 'X-Frame-Options', value: 'DENY' },
  // Controls how much referrer info is included with requests. Sends full URL
  // to same-origin, only origin to cross-origin HTTPS — keeps analytics working
  // while not leaking paths to third parties.
  { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
  // Restricts browser feature APIs. Disabling camera/mic/geo/browsing-topics
  // reduces attack surface; none of these are used by the Nixopus UI.
  {
    key: 'Permissions-Policy',
    value: 'camera=(), microphone=(), geolocation=(), browsing-topics=()'
  },
  // Disable legacy XSS auditor. OWASP recommends '0' because '1; mode=block'
  // can itself be exploited as an XSS vector in some browsers.
  { key: 'X-XSS-Protection', value: '0' },
  // Baseline CSP. unsafe-inline is required for Next.js hydration inline
  // scripts; a nonce-based CSP to eliminate it is tracked separately.
  // connect-src allows ws:/wss: for the terminal WebSocket and https: for
  // the backend API + PostHog (hosts are runtime env vars, so wildcarded here).
  {
    key: 'Content-Security-Policy',
    value: [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline'",
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
      "img-src 'self' data: blob: https:",
      "font-src 'self' https://fonts.gstatic.com",
      "connect-src 'self' ws: wss: https: http:",
      "frame-ancestors 'none'",
      "base-uri 'self'",
      "form-action 'self'"
    ].join('; ')
  }
];

const nextConfig: NextConfig = {
  outputFileTracingRoot: nextConfigDir,
  output: 'standalone',
  // Remove the X-Powered-By: Next.js response header to avoid fingerprinting.
  poweredByHeader: false,
  // Do not ship source maps to the browser in production. Source maps for error
  // monitoring should be uploaded to a service like Sentry server-side instead.
  productionBrowserSourceMaps: false,
  transpilePackages: [],
  basePath: process.env.BASE_PATH || '',
  assetPrefix: process.env.ASSET_PREFIX || undefined,
  env: {
    NEXT_PUBLIC_BASE_PATH: process.env.BASE_PATH || ''
  },
  images: {
    unoptimized: true
  },
  async headers() {
    return [
      {
        // Apply to every route served by Next.js.
        source: '/(.*)',
        headers: securityHeaders
      }
    ];
  }
};

export default nextConfig;
