'use client';

import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { useAppSelector } from '@/redux/hooks';
import {
  listDynamicWorkflows,
  deleteDynamicWorkflow,
  getDynamicWorkflow,
  listToolRegistry,
  describeNode
} from '@/packages/lib/workflow-storage';
import type { ToolRegistryEntry, ToolCategory, NodePaletteItem } from '@/packages/types/workflow';

interface WorkflowSummary {
  id: string;
  name?: string;
  description?: string;
  source: 'code' | 'dynamic';
}

interface UseWorkflowsOptions {
  applicationId: string;
}

export function useWorkflows({ applicationId }: UseWorkflowsOptions) {
  const [workflows, setWorkflows] = useState<WorkflowSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  const fetchWorkflows = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const headers: Record<string, string> = {};
      if (token) headers['Authorization'] = `Bearer ${token}`;
      if (activeOrg?.id) headers['X-Organization-Id'] = activeOrg.id;

      const dynamicResult = await listDynamicWorkflows({ applicationId, headers });

      const dynamicWorkflows: WorkflowSummary[] = dynamicResult.map((w: any) => ({
        id: w.id,
        name: w.name || w.id,
        description: w.description,
        source: 'dynamic' as const
      }));

      setWorkflows(dynamicWorkflows);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch workflows');
    } finally {
      setIsLoading(false);
    }
  }, [applicationId, token, activeOrg?.id]);

  useEffect(() => {
    fetchWorkflows();
  }, [fetchWorkflows]);

  const deleteWorkflow = useCallback(
    async (workflowId: string) => {
      setIsDeleting(true);
      try {
        const headers: Record<string, string> = {};
        if (token) headers['Authorization'] = `Bearer ${token}`;
        if (activeOrg?.id) headers['X-Organization-Id'] = activeOrg.id;
        await deleteDynamicWorkflow(workflowId, { applicationId, headers });
        await fetchWorkflows();
      } finally {
        setIsDeleting(false);
      }
    },
    [applicationId, token, activeOrg?.id, fetchWorkflows]
  );

  return {
    workflows,
    isLoading,
    isDeleting,
    error,
    refetch: fetchWorkflows,
    deleteWorkflow
  };
}

interface StepDetail {
  id: string;
  description: string;
  inputSchema: string;
  outputSchema: string;
  resumeSchema: string;
  suspendSchema: string;
  stateSchema: string;
  metadata?: Record<string, unknown>;
}

interface AllStepDetail extends StepDetail {
  isWorkflow: boolean;
}

export interface WorkflowDetail {
  name: string;
  description?: string;
  steps: Record<string, StepDetail>;
  allSteps: Record<string, AllStepDetail>;
  stepGraph: any[];
  inputSchema: string;
  outputSchema: string;
  stateSchema: string;
}

export interface PlanningMessageDetail {
  role: 'user' | 'assistant';
  content: string;
  graph?: { name?: string; nodes: any[]; edges: any[] } | null;
  /** Unix timestamp (ms) for chronological ordering with execution messages */
  timestamp?: number;
}

export interface DynamicWorkflowDetail {
  id: string;
  name: string;
  description?: string;
  applicationId?: string;
  chatThreadId?: string;
  nodes: any[];
  edges: any[];
  planningMessages?: PlanningMessageDetail[];
  executionMessages?: import('@/packages/types/workflow').ExecutionMessage[];
}

interface UseWorkflowDetailOptions {
  workflowId: string;
  applicationId: string;
}

export function useWorkflowDetail({ workflowId, applicationId }: UseWorkflowDetailOptions) {
  const [workflow, setWorkflow] = useState<WorkflowDetail | null>(null);
  const [dynamicWorkflow, setDynamicWorkflow] = useState<DynamicWorkflowDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  const fetchDetail = useCallback(async () => {
    if (workflowId === 'new') {
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      const headers: Record<string, string> = {};
      if (token) headers['Authorization'] = `Bearer ${token}`;
      if (activeOrg?.id) headers['X-Organization-Id'] = activeOrg.id;

      const dynamic = await getDynamicWorkflow(workflowId, { applicationId, headers });
      if (dynamic) {
        setDynamicWorkflow(dynamic);
      } else {
        setError('Workflow not found');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch workflow details');
    } finally {
      setIsLoading(false);
    }
  }, [workflowId, applicationId, token, activeOrg?.id]);

  useEffect(() => {
    fetchDetail();
  }, [fetchDetail]);

  return { workflow, dynamicWorkflow, isLoading, error, refetch: fetchDetail };
}

interface UseToolRegistryOptions {
  category?: ToolCategory;
  enabled?: boolean;
}

export function useToolRegistry({ category, enabled = true }: UseToolRegistryOptions = {}) {
  const [tools, setTools] = useState<ToolRegistryEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fetchedRef = useRef(false);

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  const getHeaders = useCallback(() => {
    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    if (activeOrg?.id) headers['X-Organization-Id'] = activeOrg.id;
    return headers;
  }, [token, activeOrg?.id]);

  const fetchTools = useCallback(async () => {
    if (!enabled) return;

    setIsLoading(true);
    setError(null);

    try {
      const result = await listToolRegistry({
        category,
        headers: getHeaders()
      });
      setTools(result);
      fetchedRef.current = true;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tools');
    } finally {
      setIsLoading(false);
    }
  }, [category, enabled, getHeaders]);

  useEffect(() => {
    if (!fetchedRef.current && enabled) {
      fetchTools();
    }
  }, [fetchTools, enabled]);

  const refetch = useCallback(() => {
    fetchedRef.current = false;
    return fetchTools();
  }, [fetchTools]);

  return { tools, isLoading, error, refetch };
}

interface UseNodeDescriptionOptions {
  node: { id: string; type: string; data: Record<string, unknown> } | null;
  headers?: Record<string, string>;
}

export function useNodeDescription({ node, headers }: UseNodeDescriptionOptions) {
  const [result, setResult] = useState<
    import('@/packages/types/workflow').NodeDescribeResult | null
  >(null);
  const [isLoading, setIsLoading] = useState(false);

  const fetchDescription = useCallback(async () => {
    if (!node) return;

    setIsLoading(true);
    setResult(null);

    try {
      const desc = await describeNode(node, { headers });
      setResult(desc);
    } catch {
      setResult(null);
    } finally {
      setIsLoading(false);
    }
  }, [node, headers]);

  useEffect(() => {
    fetchDescription();
  }, [fetchDescription]);

  return { result, isLoading };
}

const PALETTE_ITEMS: NodePaletteItem[] = [
  {
    type: 'trigger',
    label: 'Manual',
    description: 'Start manually',
    category: 'triggers',
    defaultData: {
      label: 'Manual Trigger',
      triggerType: 'manual',
      executionStatus: 'idle',
      config: {}
    }
  },
  {
    type: 'trigger',
    label: 'Git Push',
    description: 'On code push',
    category: 'triggers',
    defaultData: { label: 'Git Push', triggerType: 'git_push', executionStatus: 'idle', config: {} }
  },
  {
    type: 'agent',
    label: 'Deploy',
    description: 'Deploy agent',
    category: 'steps',
    defaultData: {
      label: 'Deploy Agent',
      agentId: 'deploy-agent',
      agentName: 'Deploy Agent',
      executionStatus: 'idle'
    }
  },
  {
    type: 'agent',
    label: 'Analyze',
    description: 'Analysis agent',
    category: 'steps',
    defaultData: {
      label: 'Analysis Agent',
      agentId: 'pipeline-analysis',
      agentName: 'Analysis Agent',
      executionStatus: 'idle'
    }
  },
  {
    type: 'tool',
    label: 'Deploy',
    description: 'Deploy project',
    category: 'steps',
    defaultData: {
      label: 'Deploy Project',
      toolId: 'deployProject',
      toolName: 'Deploy Project',
      executionStatus: 'idle'
    }
  },
  {
    type: 'tool',
    label: 'Scan',
    description: 'Security scan',
    category: 'steps',
    defaultData: {
      label: 'Security Scan',
      toolId: 'securityScan',
      toolName: 'Security Scan',
      executionStatus: 'idle'
    }
  },
  {
    type: 'condition',
    label: 'Branch',
    description: 'Conditional logic',
    category: 'control',
    defaultData: { label: 'Condition', conditionExpression: '', executionStatus: 'idle' }
  },
  {
    type: 'foreach',
    label: 'For Each',
    description: 'Iterate over array',
    category: 'control',
    defaultData: { label: 'For Each', concurrency: 1, executionStatus: 'idle' }
  },
  {
    type: 'loop',
    label: 'Loop',
    description: 'Repeat until condition',
    category: 'control',
    defaultData: {
      label: 'Loop',
      loopType: 'dowhile',
      conditionExpression: 'true',
      executionStatus: 'idle'
    }
  },
  {
    type: 'map',
    label: 'Transform',
    description: 'Map / reshape data',
    category: 'control',
    defaultData: { label: 'Transform', mappingExpression: 'return data', executionStatus: 'idle' }
  },
  {
    type: 'sleep',
    label: 'Delay',
    description: 'Wait / sleep',
    category: 'control',
    defaultData: { label: 'Delay', sleepType: 'duration', duration: 5000, executionStatus: 'idle' }
  },
  {
    type: 'approval',
    label: 'Approve',
    description: 'Human approval',
    category: 'control',
    defaultData: {
      label: 'Approval',
      approvalMessage: 'Approval required',
      executionStatus: 'idle'
    }
  },
  {
    type: 'notification',
    label: 'Slack',
    description: 'Slack message',
    category: 'notify',
    defaultData: {
      label: 'Slack',
      channel: 'slack',
      message: '',
      executionStatus: 'idle',
      config: {}
    }
  },
  {
    type: 'notification',
    label: 'Email',
    description: 'Email alert',
    category: 'notify',
    defaultData: {
      label: 'Email',
      channel: 'email',
      message: '',
      executionStatus: 'idle',
      config: {}
    }
  }
];

const CATEGORIES = [
  { key: 'triggers' as const, label: 'Triggers' },
  { key: 'steps' as const, label: 'Steps' },
  { key: 'ci-cd' as const, label: 'CI / CD' },
  { key: 'control' as const, label: 'Control' },
  { key: 'notify' as const, label: 'Notify' }
];

function registryEntryToPaletteItem(entry: ToolRegistryEntry): NodePaletteItem {
  return {
    type: 'tool',
    label: entry.name.length > 14 ? entry.id : entry.name,
    description: entry.description.slice(0, 60),
    category: 'ci-cd',
    defaultData: {
      label: entry.name,
      toolId: entry.id,
      toolName: entry.name,
      toolCategory: entry.category,
      executionStatus: 'idle'
    }
  };
}

export function useNodePalette() {
  const { tools: registryTools, isLoading: isLoadingTools } = useToolRegistry();

  const cicdItems = useMemo(() => {
    const sandboxTools = registryTools.filter((t) => t.category !== 'builtin');
    return sandboxTools.map(registryEntryToPaletteItem);
  }, [registryTools]);

  const allItems = useMemo(() => [...PALETTE_ITEMS, ...cicdItems], [cicdItems]);

  return {
    items: allItems,
    categories: CATEGORIES,
    isLoadingTools
  };
}
