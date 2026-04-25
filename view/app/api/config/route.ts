import { NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';

/** Bracket access avoids build-time inlining of empty process.env in standalone output. */
function runtimeEnv(name: 'AGENT_MODEL' | 'AGENT_LIGHT_MODEL' | 'LLM_PROVIDER'): string {
  return process.env[name] ?? '';
}

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

  const isSelfHosted = process.env.SELF_HOSTED === 'true' || false;

  const response = NextResponse.json({
    baseUrl: apiUrl,
    websocketUrl: process.env.WEBSOCKET_URL || derived.websocketUrl,
    webhookUrl: process.env.WEBHOOK_URL || derived.webhookUrl,
    port: process.env.NEXT_PUBLIC_PORT || '7443',
    passwordLoginEnabled: process.env.PASSWORD_LOGIN_ENABLED !== 'false',
    agentUrl: process.env.AGENT_URL || '',
    githubAppSlug: process.env.GITHUB_APP_SLUG || '',
    selfHosted: isSelfHosted,
    posthogKey: process.env.POSTHOG_KEY || '',
    posthogHost: process.env.POSTHOG_HOST || '',
    turnstileSiteKey: process.env.TURNSTILE_SITE_KEY || '',
    ...(isSelfHosted && {
      agentModel: runtimeEnv('AGENT_MODEL'),
      agentLightModel: runtimeEnv('AGENT_LIGHT_MODEL'),
      llmProvider: runtimeEnv('LLM_PROVIDER') || 'openrouter'
    })
  });
  response.headers.set('Cache-Control', 'public, max-age=300, stale-while-revalidate=60');
  return response;
}
