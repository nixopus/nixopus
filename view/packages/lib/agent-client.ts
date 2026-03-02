import { MastraClient } from '@mastra/client-js';

export const AGENT_ID = 'deploy-agent';

export function createAgentClient(authHeaders: Record<string, string> = {}): MastraClient {
  return new MastraClient({
    baseUrl: '/api/agent/',
    headers: authHeaders
  });
}
