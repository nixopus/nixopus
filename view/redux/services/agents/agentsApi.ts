import { getOctoagentUrl } from '@/redux/conf';
import { authClient } from '@/packages/lib/auth-client';

const AGENT_ID = 'nixopusAgent';
const RESOURCE_ID = 'nixopusAgent';

function getNativeFetch(): typeof fetch {
  if (typeof window === 'undefined') {
    return fetch;
  }
  return fetch;
}

async function getAuthToken(): Promise<string | null> {
  const { data: session } = await authClient.getSession();
  if (!session?.session) {
    return null;
  }
  return '';
}

function createAuthHeaders(token: string | null): HeadersInit {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

function createRequestOptions(
  method: string,
  token: string | null,
  body?: string,
  signal?: AbortSignal
): RequestInit {
  return {
    method,
    headers: createAuthHeaders(token),
    credentials: 'include',
    body,
    signal
  };
}

async function handleApiResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Request failed: ${response.status} - ${errorText}`);
  }
  return response.json();
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === 'AbortError';
}

function handleApiError(error: unknown, operation: string): never {
  if (isAbortError(error)) {
    throw error;
  }
  console.error(`Error ${operation}:`, error);
  throw error;
}

function encodeRuntimeContext(context: Record<string, unknown> = {}): string {
  return btoa(JSON.stringify(context));
}

function buildThreadsUrl(octoagentUrl: string): string {
  const runtimeContext = encodeRuntimeContext({});
  return `${octoagentUrl}/api/memory/network/threads?resourceid=${RESOURCE_ID}&agentId=${AGENT_ID}&runtimeContext=${encodeURIComponent(runtimeContext)}`;
}

function buildThreadMessagesUrl(octoagentUrl: string, threadId: string): string {
  return `${octoagentUrl}/api/memory/network/threads/${threadId}/messages?agentId=${AGENT_ID}`;
}

function extractMessagesFromResponse(data: unknown): ThreadMessage[] {
  if (Array.isArray(data)) {
    return data;
  }

  if (data && typeof data === 'object') {
    const dataObj = data as Record<string, unknown>;
    if (Array.isArray(dataObj.messages)) {
      return dataObj.messages as ThreadMessage[];
    }
    if (Array.isArray(dataObj.data)) {
      return dataObj.data as ThreadMessage[];
    }
  }

  console.warn('Unexpected API response format:', data);
  return [];
}

function convertChatMessageToMastraFormat(message: ChatMessage) {
  return {
    role: message.role,
    content: [{ type: 'text', text: message.content }]
  };
}

function buildMastraMessages(request: ChatRequest) {
  const historyMessages = (request.history || []).map(convertChatMessageToMastraFormat);
  const currentMessage = {
    role: 'user' as const,
    content: [{ type: 'text', text: request.message }]
  };
  return [...historyMessages, currentMessage];
}

function buildStreamPayload(request: ChatRequest, runId: string, threadId: string) {
  return {
    messages: buildMastraMessages(request),
    runId,
    modelSettings: {},
    runtimeContext: {},
    threadId,
    resourceId: RESOURCE_ID
  };
}

function extractDataContent(line: string): string | null {
  if (!line.startsWith('data: ')) {
    return null;
  }
  return line.slice(6).trim();
}

function isDoneSignal(dataContent: string): boolean {
  return dataContent === '[DONE]';
}

function parseTextDelta(data: unknown, callbacks: StreamCallbacks): void {
  if (typeof data !== 'object' || data === null) {
    return;
  }

  const obj = data as Record<string, unknown>;
  const eventType = obj.type as string;

  if (eventType === 'text-delta' && 'payload' in obj) {
    const payload = obj.payload;
    if (
      typeof payload === 'object' &&
      payload !== null &&
      'text' in payload &&
      typeof (payload as Record<string, unknown>).text === 'string'
    ) {
      callbacks.onContent((payload as Record<string, unknown>).text as string);
      return;
    }
  }

  if (eventType === 'message' || eventType === 'message-delta') {
    const payload = obj.payload;
    if (typeof payload === 'object' && payload !== null) {
      const payloadObj = payload as Record<string, unknown>;

      if (typeof payloadObj.text === 'string' && payloadObj.text) {
        callbacks.onContent(payloadObj.text);
        return;
      }

      if (typeof payloadObj.content === 'string' && payloadObj.content) {
        callbacks.onContent(payloadObj.content);
        return;
      }

      if (Array.isArray(payloadObj.content)) {
        const textParts = payloadObj.content
          .filter(
            (item: unknown) =>
              typeof item === 'object' &&
              item !== null &&
              'type' in item &&
              (item as Record<string, unknown>).type === 'text' &&
              'text' in item &&
              typeof (item as Record<string, unknown>).text === 'string'
          )
          .map((item: unknown) => (item as Record<string, unknown>).text as string);

        if (textParts.length > 0) {
          callbacks.onContent(textParts.join(''));
          return;
        }
      }
    }
  }
}

function extractToolCallInfo(toolCallData: unknown): {
  toolName: string | null;
  toolCallId: string | null;
  args: Record<string, any>;
} {
  if (typeof toolCallData !== 'object' || toolCallData === null) {
    return { toolName: null, toolCallId: null, args: {} };
  }

  const obj = toolCallData as Record<string, unknown>;
  const toolName =
    (obj.name as string) || (obj.functionName as string) || (obj.toolName as string) || null;
  const toolCallId =
    (obj.id as string) || (obj.toolCallId as string) || (obj.callId as string) || null;
  const args =
    (obj.arguments as Record<string, any>) ||
    (obj.args as Record<string, any>) ||
    (obj.parameters as Record<string, any>) ||
    {};

  return { toolName, toolCallId, args };
}

function parseToolCall(data: unknown, callbacks: StreamCallbacks): void {
  if (typeof data === 'object' && data !== null && 'type' in data && data.type === 'tool-call') {
    const toolCallData = 'payload' in data ? data.payload : data;
    const { toolName, toolCallId, args } = extractToolCallInfo(toolCallData);

    if (toolName && toolCallId) {
      callbacks.onToolCall?.(toolName, toolCallId, args);
    }
  }
}

function extractToolResultInfo(
  resultData: unknown,
  fullData: unknown
): {
  toolCallId: string | null;
  result: any;
  isError: boolean;
} {
  if (typeof resultData !== 'object' || resultData === null) {
    return { toolCallId: null, result: resultData, isError: false };
  }

  const resultObj = resultData as Record<string, unknown>;
  const fullObj =
    typeof fullData === 'object' && fullData !== null ? (fullData as Record<string, unknown>) : {};

  const toolCallId =
    (resultObj.toolCallId as string) ||
    (resultObj.id as string) ||
    (resultObj.callId as string) ||
    (fullObj.toolCallId as string) ||
    (fullObj.id as string) ||
    null;

  const result =
    resultObj.result !== undefined
      ? resultObj.result
      : resultObj.content !== undefined
        ? resultObj.content
        : resultObj;

  const isError = Boolean(resultObj.isError || resultObj.error);

  return { toolCallId, result, isError };
}

function parseToolResult(data: unknown, callbacks: StreamCallbacks): void {
  if (typeof data === 'object' && data !== null && 'type' in data && data.type === 'tool-result') {
    const resultData = 'payload' in data ? data.payload : data;
    const { toolCallId, result, isError } = extractToolResultInfo(resultData, data);

    if (toolCallId) {
      callbacks.onToolResult?.(toolCallId, result, isError);
    } else {
      console.warn(
        '[AI Stream] Tool result event missing toolCallId. Full data:',
        JSON.stringify(data, null, 2)
      );
    }
  }
}

function parseErrorEvent(data: unknown, callbacks: StreamCallbacks): void {
  if (typeof data !== 'object' || data === null) {
    return;
  }

  const obj = data as Record<string, unknown>;
  if (obj.type === 'error' && 'payload' in obj) {
    const payload = obj.payload;
    if (typeof payload === 'object' && payload !== null) {
      const payloadObj = payload as Record<string, unknown>;

      let errorMessage = 'An error occurred while processing your request.';

      if (payloadObj.error && typeof payloadObj.error === 'object') {
        const errorObj = payloadObj.error as Record<string, unknown>;
        if (typeof errorObj.message === 'string') {
          errorMessage = errorObj.message;
        }
      } else if (typeof payloadObj.message === 'string') {
        errorMessage = payloadObj.message;
      }

      callbacks.onContent(`\n\n**Error:** ${errorMessage}\n\n`);

      callbacks.onError(new Error(errorMessage));
    }
  }
}

function parseFinishEvent(data: unknown, callbacks: StreamCallbacks): boolean {
  if (typeof data === 'object' && data === null) {
    return false;
  }

  const obj = data as Record<string, unknown>;
  const eventType = obj.type as string;

  if (eventType === 'finish' || eventType === 'step-finish') {
    if (eventType === 'finish') {
      callbacks.onDone();
      return true;
    }
  }

  return false;
}

function shouldSkipParseError(dataContent: string): boolean {
  if (dataContent === '[DONE]' || dataContent.length === 0) {
    return true;
  }
  if (!dataContent.startsWith('[') && !dataContent.startsWith('{')) {
    return true;
  }
  return false;
}

function parseStreamData(dataContent: string, callbacks: StreamCallbacks): boolean {
  if (isDoneSignal(dataContent)) {
    callbacks.onDone();
    return true;
  }

  try {
    const data = JSON.parse(dataContent);
    parseErrorEvent(data, callbacks);
    parseTextDelta(data, callbacks);
    parseToolCall(data, callbacks);
    parseToolResult(data, callbacks);

    if (parseFinishEvent(data, callbacks)) {
      return true;
    }
  } catch (err) {
    if (!shouldSkipParseError(dataContent)) {
      console.error('Error parsing SSE data:', err, 'Data:', dataContent.substring(0, 100));
    }
  }

  return false;
}

async function processStreamChunk(
  reader: ReadableStreamDefaultReader<Uint8Array>,
  decoder: TextDecoder,
  buffer: string,
  callbacks: StreamCallbacks
): Promise<{ done: boolean; buffer: string }> {
  try {
    const { done, value } = await reader.read();

    if (done) {
      if (buffer.trim()) {
        const lines = buffer.split('\n');
        for (const line of lines) {
          const dataContent = extractDataContent(line);
          if (dataContent) {
            parseStreamData(dataContent, callbacks);
          }
        }
      }
      callbacks.onDone();
      return { done: true, buffer: '' };
    }

    if (!value || value.length === 0) {
      return { done: false, buffer };
    }

    const decoded = decoder.decode(value, { stream: true });
    buffer += decoded;
    const lines = buffer.split('\n');
    buffer = lines.pop() || '';

    for (const line of lines) {
      const dataContent = extractDataContent(line);
      if (!dataContent) continue;

      const shouldStop = parseStreamData(dataContent, callbacks);
      if (shouldStop) {
        return { done: true, buffer };
      }
    }

    return { done: false, buffer };
  } catch (error) {
    callbacks.onError(error instanceof Error ? error : new Error(String(error)));
    return { done: true, buffer: '' };
  }
}

async function processStream(response: Response, callbacks: StreamCallbacks): Promise<void> {
  if (!response.body) {
    throw new Error('Response body is null');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (true) {
      const { done, buffer: newBuffer } = await processStreamChunk(
        reader,
        decoder,
        buffer,
        callbacks
      );

      if (done) {
        break;
      }

      buffer = newBuffer;
    }
  } catch (error) {
    callbacks.onError(error instanceof Error ? error : new Error(String(error)));
    throw error;
  } finally {
    reader.releaseLock();
  }
}

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface ChatRequest {
  message: string;
  history?: ChatMessage[];
  threadId?: string;
}

export interface ChatResponse {
  content: string;
  done: boolean;
}

export interface StreamCallbacks {
  onContent: (content: string) => void;
  onToolCall?: (toolName: string, toolCallId: string, args: Record<string, any>) => void;
  onToolResult?: (toolCallId: string, result: any, isError?: boolean) => void;
  onDone: () => void;
  onError: (error: Error) => void;
}

export interface Thread {
  id: string;
  resourceId: string;
  title: string;
  metadata: any;
  createdAt: string;
  updatedAt: string;
}

export interface ThreadMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content?: Array<{ type: string; text?: string; [key: string]: any }> | string;
  createdAt?: string;
  [key: string]: any;
}

export async function getThreads(signal?: AbortSignal): Promise<Thread[]> {
  try {
    const token = await getAuthToken();
    const octoagentUrl = await getOctoagentUrl();
    const nativeFetch = getNativeFetch();
    const url = buildThreadsUrl(octoagentUrl);
    const options = createRequestOptions('GET', token, undefined, signal);

    const response = await nativeFetch(url, options);
    return handleApiResponse<Thread[]>(response);
  } catch (error) {
    handleApiError(error, 'fetching threads');
  }
}

export async function getThreadMessages(
  threadId: string,
  signal?: AbortSignal
): Promise<ThreadMessage[]> {
  try {
    const token = await getAuthToken();
    const octoagentUrl = await getOctoagentUrl();
    const nativeFetch = getNativeFetch();
    const url = buildThreadMessagesUrl(octoagentUrl, threadId);
    const options = createRequestOptions('GET', token, undefined, signal);

    const response = await nativeFetch(url, options);
    const data = await handleApiResponse<unknown>(response);
    return extractMessagesFromResponse(data);
  } catch (error) {
    handleApiError(error, 'fetching thread messages');
  }
}

export async function streamAIChat(
  request: ChatRequest,
  callbacks: StreamCallbacks,
  signal?: AbortSignal
): Promise<void> {
  try {
    const token = await getAuthToken();
    const octoagentUrl = await getOctoagentUrl();
    const nativeFetch = getNativeFetch();

    const runId = crypto.randomUUID();
    const threadId = request.threadId || crypto.randomUUID();
    const payload = buildStreamPayload(request, runId, threadId);
    const url = `${octoagentUrl}/api/agents/${AGENT_ID}/stream`;
    const options = createRequestOptions('POST', token, JSON.stringify(payload), signal);

    const response = await nativeFetch(url, options);

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Request failed: ${response.status} - ${errorText}`);
    }

    await processStream(response, callbacks);
  } catch (error) {
    if (isAbortError(error)) {
      return;
    }
    callbacks.onError(error as Error);
  }
}
