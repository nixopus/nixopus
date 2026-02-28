import { MastraClient } from '@mastra/client-js';

const AGENT_URL = process.env.NEXT_PUBLIC_AGENT_URL || '';

export const AGENT_ID = 'deploy-agent';

export function isAgentConfigured(): boolean {
  return Boolean(AGENT_URL);
}

export function createAgentClient(authHeaders: Record<string, string> = {}): MastraClient {
  return new MastraClient({
    baseUrl: '/api/agent/',
    headers: authHeaders
  });
}
