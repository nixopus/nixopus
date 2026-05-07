import { getBaseUrl } from '@/redux/conf';

export const AGENT_ID = 'deploy-agent';

let cachedApiBaseUrl = '';

async function getApiBaseUrl(): Promise<string> {
  if (cachedApiBaseUrl) return cachedApiBaseUrl;
  const url = await getBaseUrl();
  cachedApiBaseUrl = url.replace(/\/+$/, '');
  return cachedApiBaseUrl;
}

export interface StreamChunk {
  type: string;
  data?: unknown;
}

function parseNamedEvent(eventType: string, dataLine: string): StreamChunk | null {
  try {
    return { type: eventType, data: JSON.parse(dataLine) };
  } catch {
    return { type: eventType, data: dataLine };
  }
}

export async function* readSseStream(
  body: ReadableStream<Uint8Array>,
  signal?: AbortSignal
): AsyncGenerator<StreamChunk> {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let currentEvent = '';

  try {
    for (;;) {
      if (signal?.aborted) break;
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';

      for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed === '') {
          currentEvent = '';
          continue;
        }
        if (trimmed.startsWith('event: ')) {
          currentEvent = trimmed.slice(7).trim();
          continue;
        }
        if (trimmed.startsWith('data: ')) {
          const dataContent = trimmed.slice(6);
          const eventType = currentEvent || 'message';
          const chunk = parseNamedEvent(eventType, dataContent);
          if (chunk) yield chunk;
        }
      }
    }
    if (buffer.trim()) {
      const trimmed = buffer.trim();
      if (trimmed.startsWith('data: ')) {
        const dataContent = trimmed.slice(6);
        const eventType = currentEvent || 'message';
        const chunk = parseNamedEvent(eventType, dataContent);
        if (chunk) yield chunk;
      }
    }
  } finally {
    reader.releaseLock();
  }
}

async function agentFetch(
  path: string,
  options: {
    method?: string;
    body?: Record<string, unknown>;
    headers: Record<string, string>;
    signal?: AbortSignal;
  }
): Promise<Response> {
  const baseUrl = await getApiBaseUrl();
  const url = `${baseUrl}/v1/agent${path}`;

  const reqHeaders: Record<string, string> = {
    'Content-Type': 'application/json'
  };
  if (options.headers['Authorization']) {
    reqHeaders['Authorization'] = options.headers['Authorization'];
  }
  if (options.headers['X-Organization-Id']) {
    reqHeaders['X-Organization-Id'] = options.headers['X-Organization-Id'];
  }
  if (options.headers['X-Model-Id']) {
    reqHeaders['X-Model-Id'] = options.headers['X-Model-Id'];
  }

  const init: RequestInit = {
    method: options.method || 'POST',
    headers: reqHeaders,
    signal: options.signal
  };

  if (options.body && init.method !== 'GET') {
    init.body = JSON.stringify(options.body);
  }

  const response = await fetch(url, init);

  if (!response.ok) {
    const text = await response.text().catch(() => 'Unknown error');
    throw new Error(`Agent request failed (${response.status}): ${text}`);
  }

  return response;
}

export async function* streamAgent(
  content: string,
  threadId: string,
  _resourceId: string,
  headers: Record<string, string>,
  signal?: AbortSignal,
  model?: string
): AsyncGenerator<StreamChunk> {
  const body: Record<string, unknown> = {
    input: content,
    thread_id: threadId
  };
  if (model) {
    body.model = model;
  }

  const response = await agentFetch('/chat/stream', { body, headers, signal });

  if (!response.body) {
    throw new Error('No response body from agent');
  }

  yield* readSseStream(response.body, signal);
}

export async function cancelStream(
  threadId: string,
  headers: Record<string, string>
): Promise<{ status: string; message: string }> {
  const response = await agentFetch('/chat/cancel', {
    body: { thread_id: threadId },
    headers
  });
  return response.json();
}

export interface AgentThread {
  id: string;
  title: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export async function createThread(
  headers: Record<string, string>,
  opts?: { id?: string; title?: string }
): Promise<AgentThread> {
  const response = await agentFetch('/threads', {
    body: { id: opts?.id, title: opts?.title },
    headers
  });
  const result = await response.json();
  return result.data;
}

export async function listThreads(headers: Record<string, string>): Promise<AgentThread[]> {
  const response = await agentFetch('/threads', { method: 'GET', headers });
  const result = await response.json();
  return result.data ?? [];
}

export async function getThreadMessages(
  threadId: string,
  headers: Record<string, string>
): Promise<
  Array<{
    id: string;
    thread_id: string;
    role: string;
    content: string;
    tool_calls?: unknown[];
    created_at: string;
    seq: number;
  }>
> {
  const response = await agentFetch(`/threads/${threadId}/messages`, {
    method: 'GET',
    headers
  });
  const result = await response.json();
  return result.data ?? [];
}

export async function updateThread(
  threadId: string,
  title: string,
  headers: Record<string, string>
): Promise<void> {
  await agentFetch(`/threads/${threadId}`, {
    method: 'PATCH',
    body: { title },
    headers
  });
}

export async function deleteThread(
  threadId: string,
  headers: Record<string, string>
): Promise<void> {
  await agentFetch(`/threads/${threadId}`, { method: 'DELETE', headers });
}
