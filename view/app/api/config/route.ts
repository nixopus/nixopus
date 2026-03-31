import { NextResponse } from 'next/server';

function deriveUrls(apiUrl: string) {
  const base = apiUrl.replace(/\/api\/?$/, '');
  const wsScheme = base.startsWith('https') ? 'wss' : 'ws';
  return {
    websocketUrl: `${base.replace(/^https?/, wsScheme)}/ws`,
    webhookUrl: `${base}/api/v1/webhook`
  };
}

export async function GET() {
  const apiUrl = process.env.API_URL || 'http://localhost:8080/api';
  const derived = deriveUrls(apiUrl);

  const response = NextResponse.json({
    baseUrl: apiUrl,
    websocketUrl: process.env.WEBSOCKET_URL || derived.websocketUrl,
    webhookUrl: process.env.WEBHOOK_URL || derived.webhookUrl,
    port: process.env.NEXT_PUBLIC_PORT || '7443',
    passwordLoginEnabled: process.env.PASSWORD_LOGIN_ENABLED !== 'false',
    agentUrl: process.env.AGENT_URL || '',
    githubAppSlug: process.env.GITHUB_APP_SLUG || '',
    selfHosted: process.env.SELF_HOSTED === 'true' || false,
    posthogKey: process.env.POSTHOG_KEY || '',
    posthogHost: process.env.POSTHOG_HOST || ''
  });
  response.headers.set('Cache-Control', 'public, max-age=300, stale-while-revalidate=60');
  return response;
}
