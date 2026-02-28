import type { Edge, Node } from '@xyflow/react';

export interface WorkflowCanvasHandle {
  getNodes: () => unknown[];
  getEdges: () => Edge[];
  setNodeData: (nodeId: string, data: Partial<unknown>) => void;
  resetAllNodes: () => void;
  replaceGraph: (nodes: unknown[], edges: Edge[]) => void;
}

export type WorkflowNodeType =
  | 'trigger'
  | 'agent'
  | 'tool'
  | 'condition'
  | 'approval'
  | 'notification'
  | 'foreach'
  | 'loop'
  | 'map'
  | 'sleep';

export type NodeExecutionStatus =
  | 'idle'
  | 'running'
  | 'success'
  | 'failed'
  | 'suspended'
  | 'skipped';

export type WorkflowRunStatus =
  | 'idle'
  | 'running'
  | 'success'
  | 'failed'
  | 'suspended'
  | 'cancelled';

export type TriggerType = 'manual' | 'git_push' | 'schedule' | 'webhook';

export type ToolCategory = 'builtin' | 'sandbox-cli' | 'sandbox-container' | 'script';

export interface BaseNodeData {
  label: string;
  description?: string;
  icon?: string;
  executionStatus: NodeExecutionStatus;
  executionDuration?: number;
  executionOutput?: Record<string, unknown>;
  executionError?: string;
  [key: string]: unknown;
}

export interface TriggerNodeData extends BaseNodeData {
  triggerType: TriggerType;
  config: {
    branch?: string;
    schedule?: string;
    webhookUrl?: string;
  };
}

export interface AgentNodeData extends BaseNodeData {
  agentId: string;
  agentName: string;
  systemPrompt?: string;
  tools?: string[];
  model?: string;
}

export interface SandboxOverrides {
  allowNetwork?: boolean;
  timeout?: number;
  image?: string;
}

export interface ToolNodeData extends BaseNodeData {
  toolId: string;
  toolName: string;
  toolCategory?: ToolCategory;
  command?: string;
  image?: string;
  scriptPath?: string;
  timeout?: number;
  retries?: number;
  env?: Record<string, string>;
  sandbox?: SandboxOverrides;
  inputMapping?: Record<string, string>;
}

export interface ConditionNodeData extends BaseNodeData {
  conditionExpression: string;
  trueBranch?: string;
  falseBranch?: string;
}

export interface ApprovalNodeData extends BaseNodeData {
  approvalMessage: string;
  approvers?: string[];
  timeout?: number;
}

export interface NotificationNodeData extends BaseNodeData {
  channel: 'slack' | 'discord' | 'email';
  message: string;
  config: Record<string, string>;
}

export interface ForeachNodeData extends BaseNodeData {
  concurrency?: number;
}

export interface LoopNodeData extends BaseNodeData {
  loopType: 'dowhile' | 'dountil';
  conditionExpression: string;
}

export interface MapNodeData extends BaseNodeData {
  mappingExpression: string;
}

export interface SleepNodeData extends BaseNodeData {
  sleepType: 'duration' | 'until';
  duration?: number;
  untilDate?: string;
}

export type WorkflowNodeData =
  | TriggerNodeData
  | AgentNodeData
  | ToolNodeData
  | ConditionNodeData
  | ApprovalNodeData
  | NotificationNodeData
  | ForeachNodeData
  | LoopNodeData
  | MapNodeData
  | SleepNodeData;

export interface WorkflowDefinition {
  id: string;
  name: string;
  description?: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  createdAt?: string;
  updatedAt?: string;
}

export type WorkflowNode = Node<WorkflowNodeData, WorkflowNodeType>;
export type WorkflowEdge = Edge;

export interface NodePaletteItem {
  type: WorkflowNodeType;
  label: string;
  description: string;
  category: 'triggers' | 'steps' | 'control' | 'notify' | 'ci-cd';
  defaultData: Partial<WorkflowNodeData>;
}

export interface WorkflowRunInfo {
  runId: string;
  workflowId: string;
  status: WorkflowRunStatus;
  startedAt?: string;
  endedAt?: string;
  steps: Record<string, StepRunInfo>;
}

export interface StepRunInfo {
  status: NodeExecutionStatus;
  payload?: Record<string, unknown>;
  output?: Record<string, unknown>;
  error?: string;
  startedAt?: number;
  endedAt?: number;
  suspendPayload?: Record<string, unknown>;
}

export interface NodeDescribeResult {
  nodeId: string;
  summary: string;
  execution: {
    tool: string;
    category: string;
    command?: string;
    runsIn: string;
    sandboxConfig?: {
      network: boolean;
      timeout: string;
      isolation: string;
    };
    estimatedDuration?: string;
    outputs?: string[];
  };
  explanation: string;
}

export interface SandboxRequirements {
  needsNetwork?: boolean;
  needsDocker?: boolean;
  minTimeout?: number;
  setupCommands?: string[];
  image?: string;
}

export interface ToolRegistryEntry {
  id: string;
  name: string;
  category: ToolCategory;
  description: string;
  sandboxRequirements?: SandboxRequirements;
}

export interface WorkflowRunResponse {
  runId: string;
  workflowId: string;
  applicationId: string;
  status: string;
  workflowName?: string;
}

export interface WorkflowRunStatusResponse {
  runId: string;
  workflowId: string;
  status: WorkflowRunStatus;
  value?: Record<string, unknown>;
  activePaths?: string[];
  suspendedPaths?: Record<string, unknown>;
  timestamp?: number;
}

export type WorkflowStreamEventType =
  | 'run-start'
  | 'workflow-start'
  | 'workflow-step-start'
  | 'workflow-step-output'
  | 'workflow-step-progress'
  | 'workflow-step-result'
  | 'workflow-step-finish'
  | 'workflow-step-suspended'
  | 'workflow-finish'
  | 'finish'
  | 'workflow-suspend'
  | 'run-complete'
  | 'run-error'
  | 'deployment-progress'
  | 'deployment-reasoning-chunk'
  | 'build-log';

export interface MastraStreamEvent {
  type: WorkflowStreamEventType;
  runId?: string;
  from?: string;
  payload?: Record<string, unknown>;
}

export type ExecutionMessageKind =
  | 'workflow-start'
  | 'step-start'
  | 'step-result'
  | 'step-error'
  | 'step-reasoning'
  | 'step-progress'
  | 'workflow-complete'
  | 'workflow-error';

export interface ExecutionMessage {
  id: string;
  kind: ExecutionMessageKind;
  stepId?: string;
  stepLabel?: string;
  status?: NodeExecutionStatus | WorkflowRunStatus;
  output?: Record<string, unknown>;
  error?: string;
  text?: string;
  timestamp: number;
}
