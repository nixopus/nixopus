import { MastraClient } from '@mastra/client-js';

export const AGENT_ID = 'deploy-agent';
export const INCIDENT_AGENT_ID = 'incidentAgent';

let cachedAgentBaseUrl: string | null = null;
let agentBaseUrlPromise: Promise<string> | null = null;
const APP_BASE_PATH = process.env.NEXT_PUBLIC_BASE_PATH || '';

function withBasePath(path: string): string {
  if (!APP_BASE_PATH) return path;
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${APP_BASE_PATH}${normalizedPath}`.replace(/\/{2,}/g, '/');
}

const AGENT_PROXY_BASE_PATH = withBasePath('/api/agent');

async function getAgentBaseUrl(): Promise<string> {
  if (cachedAgentBaseUrl !== null) {
    return cachedAgentBaseUrl;
  }

  if (!agentBaseUrlPromise) {
    agentBaseUrlPromise = fetch(withBasePath('/api/config'))
      .then(async (res) => {
        if (!res.ok) {
          throw new Error(`Failed to load config (${res.status})`);
        }
        return res.json() as Promise<{ agentUrl?: unknown; selfHosted?: boolean }>;
      })
      .then((cfg) => {
        if (cfg.selfHosted) {
          cachedAgentBaseUrl = '';
          return '';
        }
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
    baseUrl: withBasePath('/api/agent/'),
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
  signal?: AbortSignal,
  queryParams?: Record<string, string>
): Promise<Response> {
  const baseUrl = await getAgentBaseUrl().catch(() => '');
  let url = baseUrl
    ? (() => {
        const normalizedBase = baseUrl.replace(/\/$/, '');
        return normalizedBase.endsWith('/api')
          ? `${normalizedBase}${path}`
          : `${normalizedBase}/api${path}`;
      })()
    : `${AGENT_PROXY_BASE_PATH}/api${path}`;

  if (queryParams) {
    const params = new URLSearchParams(queryParams);
    url += `?${params.toString()}`;
  }

  const reqHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'text/event-stream'
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
  signal?: AbortSignal,
  model?: string,
  requireToolApproval?: boolean
): AsyncGenerator<StreamChunk> {
  const query = model ? { model } : undefined;
  const body: Record<string, unknown> = {
    messages: [{ role: 'user', content }],
    memory: { thread: threadId, resource: resourceId }
  };
  if (requireToolApproval) {
    body.requireToolApproval = true;
  }
  const response = await agentFetch(`/agents/${AGENT_ID}/stream`, body, headers, signal, query);

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
