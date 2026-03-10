import { MastraClient } from '@mastra/client-js';

export const AGENT_ID = 'deploy-agent';

let cachedAgentBaseUrl: string | null = null;
let agentBaseUrlPromise: Promise<string> | null = null;
const AGENT_PROXY_BASE_PATH = '/api/agent';

async function getAgentBaseUrl(): Promise<string> {
  if (cachedAgentBaseUrl !== null) {
    return cachedAgentBaseUrl;
  }

  if (!agentBaseUrlPromise) {
    agentBaseUrlPromise = fetch('/api/config')
      .then(async (res) => {
        if (!res.ok) {
          throw new Error(`Failed to load config (${res.status})`);
        }
        return res.json() as Promise<{ agentUrl?: unknown }>;
      })
      .then((cfg) => {
        const agentUrl = typeof cfg.agentUrl === 'string' ? cfg.agentUrl : '';
        if (agentUrl) {
          cachedAgentBaseUrl = agentUrl;
        }
        return agentUrl;
      })
      .finally(() => {
        agentBaseUrlPromise = null;
      });
  }

  return agentBaseUrlPromise;
}

export function createAgentClient(authHeaders: Record<string, string> = {}): MastraClient {
  return new MastraClient({
    baseUrl: '/api/agent/',
    headers: authHeaders
  });
}

export interface StreamChunk {
  type: string;
  runId?: string;
  payload?: Record<string, unknown>;
}

function parseSseLine(line: string): StreamChunk | null {
  const trimmed = line.trim();
  if (!trimmed || !trimmed.startsWith('data: ')) return null;
  try {
    return JSON.parse(trimmed.slice(6));
  } catch {
    return null;
  }
}

async function* readSseStream(
  body: ReadableStream<Uint8Array>,
  signal?: AbortSignal
): AsyncGenerator<StreamChunk> {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    for (;;) {
      if (signal?.aborted) break;
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';

      for (const line of lines) {
        const chunk = parseSseLine(line);
        if (chunk) yield chunk;
      }
    }
    if (buffer.trim()) {
      const chunk = parseSseLine(buffer);
      if (chunk) yield chunk;
    }
  } finally {
    reader.releaseLock();
  }
}

async function agentFetch(
  path: string,
  body: Record<string, unknown>,
  headers: Record<string, string>,
  signal?: AbortSignal
): Promise<Response> {
  const baseUrl = await getAgentBaseUrl().catch(() => '');
  const url = baseUrl
    ? `${baseUrl.replace(/\/$/, '')}/api${path}`
    : `${AGENT_PROXY_BASE_PATH}/api${path}`;
  const reqHeaders: Record<string, string> = {
    'Content-Type': 'application/json'
  };
  if (headers['Authorization']) {
    reqHeaders['Authorization'] = headers['Authorization'];
  }

  const response = await fetch(url, {
    method: 'POST',
    headers: reqHeaders,
    body: JSON.stringify(body),
    signal
  });

  if (!response.ok) {
    const text = await response.text().catch(() => 'Unknown error');
    throw new Error(`Agent request failed (${response.status}): ${text}`);
  }
  if (!response.body) {
    throw new Error('No response body from agent');
  }

  return response;
}

/**
 * Stream a message to the agent using the configured agent URL from /api/config.
 */
export async function* streamAgent(
  content: string,
  threadId: string,
  resourceId: string,
  headers: Record<string, string>,
  signal?: AbortSignal
): AsyncGenerator<StreamChunk> {
  const response = await agentFetch(
    `/agents/${AGENT_ID}/stream`,
    {
      messages: [{ role: 'user', content }],
      memory: { thread: threadId, resource: resourceId }
    },
    headers,
    signal
  );

  yield* readSseStream(response.body!, signal);
}

export async function* approveAgentToolCall(
  params: { runId: string; toolCallId: string },
  headers: Record<string, string>,
  signal?: AbortSignal
): AsyncGenerator<StreamChunk> {
  const response = await agentFetch(
    `/agents/${AGENT_ID}/approve-tool-call`,
    params,
    headers,
    signal
  );

  yield* readSseStream(response.body!, signal);
}

export async function declineAgentToolCall(
  params: { runId: string; toolCallId: string },
  headers: Record<string, string>
): Promise<void> {
  await agentFetch(`/agents/${AGENT_ID}/decline-tool-call`, params, headers);
}
