import type {
  NodeDescribeResult,
  ToolRegistryEntry,
  ToolCategory,
  WorkflowRunResponse,
  WorkflowRunStatusResponse,
  ExecutionMessage
} from '@/packages/types/workflow';

const BASE_PATH = '/api/agent/dynamic-workflows';
const REGISTRY_PATH = '/api/agent/tool-registry';

export interface PlanningMessage {
  role: 'user' | 'assistant';
  content: string;
  graph?: { name?: string; nodes: any[]; edges: any[] } | null;
  /** Unix timestamp (ms) for chronological ordering with execution messages */
  timestamp?: number;
}

interface SaveWorkflowParams {
  id: string;
  name: string;
  description?: string;
  applicationId: string;
  nodes: any[];
  edges: any[];
  chatThreadId?: string;
  planningMessages?: PlanningMessage[];
  executionMessages?: ExecutionMessage[];
  headers?: Record<string, string>;
}

interface SaveWorkflowResult {
  id: string;
  name: string;
  status: string;
}

export async function saveWorkflow(params: SaveWorkflowParams): Promise<SaveWorkflowResult> {
  const { headers = {}, ...body } = params;

  const isUpdate = body.id && body.id !== 'new';
  const url = isUpdate ? `${BASE_PATH}/${body.id}` : BASE_PATH;
  const method = isUpdate ? 'PUT' : 'POST';

  if (!isUpdate) {
    body.id =
      body.name
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-|-$/g, '') || `workflow-${Date.now()}`;
  }

  const res = await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json', ...headers },
    body: JSON.stringify(body)
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(err.error || `Failed to save workflow (${res.status})`);
  }

  return res.json();
}

export async function listDynamicWorkflows(options: {
  applicationId: string;
  headers?: Record<string, string>;
}): Promise<any[]> {
  const { applicationId, headers = {} } = options;
  const url = `${BASE_PATH}?applicationId=${encodeURIComponent(applicationId)}`;

  const res = await fetch(url, { headers });
  if (!res.ok) return [];
  const data = await res.json();
  const workflows = data.workflows || [];
  // Frontend filter fallback if API doesn't yet support applicationId
  return workflows.filter((w: any) => !w.applicationId || w.applicationId === applicationId);
}

export async function getDynamicWorkflow(
  id: string,
  options: {
    applicationId: string;
    headers?: Record<string, string>;
  }
): Promise<any | null> {
  const { applicationId, headers = {} } = options;
  const url = `${BASE_PATH}/${id}?applicationId=${encodeURIComponent(applicationId)}`;

  const res = await fetch(url, { headers });
  if (!res.ok) return null;
  const workflow = await res.json();
  // Validate workflow belongs to this app
  if (workflow?.applicationId && workflow.applicationId !== applicationId) {
    return null;
  }
  return workflow;
}

export async function deleteDynamicWorkflow(
  id: string,
  options: {
    applicationId: string;
    headers?: Record<string, string>;
  }
): Promise<void> {
  const { applicationId, headers = {} } = options;
  const url = `${BASE_PATH}/${id}?applicationId=${encodeURIComponent(applicationId)}`;

  await fetch(url, { method: 'DELETE', headers });
}

export async function describeNode(
  node: { id: string; type: string; data: Record<string, unknown> },
  options: {
    sandboxProvider?: string;
    headers?: Record<string, string>;
  } = {}
): Promise<NodeDescribeResult | null> {
  const { sandboxProvider, headers = {} } = options;

  const res = await fetch(`${BASE_PATH}/describe-node`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...headers },
    body: JSON.stringify({ node, sandboxProvider })
  });

  if (!res.ok) return null;
  return res.json();
}

export async function runWorkflow(
  id: string,
  options: {
    applicationId: string;
    inputData?: Record<string, unknown>;
    sandboxConfig?: {
      provider?: string;
      allowNetwork?: boolean;
      timeout?: number;
    };
    headers?: Record<string, string>;
  }
): Promise<WorkflowRunResponse> {
  const { applicationId, inputData, sandboxConfig, headers = {} } = options;

  const res = await fetch(`${BASE_PATH}/${id}/run`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...headers },
    body: JSON.stringify({ applicationId, inputData, sandboxConfig })
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(err.error || `Failed to run workflow (${res.status})`);
  }

  return res.json();
}

export async function getRunStatus(
  id: string,
  runId: string,
  options: {
    applicationId: string;
    headers?: Record<string, string>;
  }
): Promise<WorkflowRunStatusResponse> {
  const { applicationId, headers = {} } = options;
  const url = `${BASE_PATH}/${id}/runs/${runId}?applicationId=${encodeURIComponent(applicationId)}`;

  const res = await fetch(url, { headers });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(err.error || `Failed to get run status (${res.status})`);
  }

  return res.json();
}

export async function streamWorkflowRun(
  id: string,
  options: {
    applicationId: string;
    inputData?: Record<string, unknown>;
    sandboxConfig?: {
      provider?: string;
      allowNetwork?: boolean;
      timeout?: number;
    };
    headers?: Record<string, string>;
    signal?: AbortSignal;
  }
): Promise<Response> {
  const { applicationId, inputData, sandboxConfig, headers = {}, signal } = options;

  const res = await fetch(`${BASE_PATH}/${id}/run/stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...headers },
    body: JSON.stringify({ applicationId, inputData, sandboxConfig }),
    signal
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(err.error || `Failed to start workflow stream (${res.status})`);
  }

  return res;
}

export async function listToolRegistry(
  options: {
    category?: ToolCategory;
    headers?: Record<string, string>;
  } = {}
): Promise<ToolRegistryEntry[]> {
  const { category, headers = {} } = options;
  const params = category ? `?category=${encodeURIComponent(category)}` : '';
  const url = `${REGISTRY_PATH}${params}`;

  const res = await fetch(url, { headers });
  if (!res.ok) return [];

  const data = await res.json();
  return data.tools ?? [];
}
