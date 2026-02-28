import type { DragEvent } from 'react';
import type { ExecutionMessage, WorkflowNodeType } from '@/packages/types/workflow';

export function sanitizeWorkflowLabel(text: string): string {
  return text.replace(/\bMastra\b/gi, 'Nixopus');
}

export function stripJsonBlocks(content: string): string {
  return content.replace(/```json\s*[\s\S]*?```/g, '').trim();
}

export function looksLikeWorkflowJson(content: string): boolean {
  if (!content.trim()) return false;
  if (content.includes('```json')) return true;
  const trimmed = content.trim();
  if (trimmed.startsWith('{') && (content.includes('"nodes"') || content.includes('"edges"'))) {
    return true;
  }
  return false;
}

export function splitContentAroundJson(content: string): {
  before: string;
  inJsonBlock: boolean;
  after: string;
} {
  const jsonStartIdx = content.indexOf('```json');
  if (jsonStartIdx === -1) {
    return { before: content, inJsonBlock: false, after: '' };
  }
  const before = content.slice(0, jsonStartIdx);
  const afterJson = content.slice(jsonStartIdx + 7);
  const closeIdx = afterJson.indexOf('```');
  const hasClosed = closeIdx !== -1;
  const after = hasClosed ? afterJson.slice(closeIdx + 3).trimStart() : '';
  return { before, inJsonBlock: !hasClosed, after };
}

export function extractReadableText(output: Record<string, unknown>): string | null {
  if (typeof output.result === 'string' && output.result.trim()) return output.result.trim();
  if (typeof output.message === 'string' && output.message.trim()) return output.message.trim();
  if (typeof output.text === 'string' && output.text.trim()) return output.text.trim();
  if (typeof output.summary === 'string' && output.summary.trim()) return output.summary.trim();
  return null;
}

export function getRemainingOutputData(
  output: Record<string, unknown>
): Record<string, unknown> | null {
  const textKeys = ['result', 'message', 'text', 'summary'];
  const remaining = Object.keys(output).filter((k) => !textKeys.includes(k));
  if (remaining.length === 0) return null;
  return Object.fromEntries(remaining.map((k) => [k, output[k]]));
}

export function createNodeFromPaletteDrop(
  event: DragEvent,
  screenToFlowPosition: (position: { x: number; y: number }) => { x: number; y: number }
): {
  id: string;
  type: WorkflowNodeType;
  position: { x: number; y: number };
  data: Record<string, unknown>;
} | null {
  const type = event.dataTransfer.getData('application/reactflow-type');
  const dataStr = event.dataTransfer.getData('application/reactflow-data');
  if (!type) return null;

  const position = screenToFlowPosition({
    x: event.clientX,
    y: event.clientY
  });

  const nodeData = dataStr ? JSON.parse(dataStr) : { label: type, executionStatus: 'idle' };

  return {
    id: `${type}-${Date.now()}`,
    type: type as WorkflowNodeType,
    position,
    data: nodeData
  };
}

export const WORKFLOW_STREAMING_MESSAGES = [
  'Sketching a workflow for you…',
  'Designing your pipeline…',
  'Sculpting the diagram…',
  'Mapping out the steps…',
  'Bringing your workflow to life…',
  'Crafting the flow…',
  'Assembling the nodes…',
  'Composing the automation…'
];

export const CHAT_SUGGESTIONS = [
  'Deploy my app with approval gates',
  'CI/CD pipeline: analyze, test, deploy',
  'Rollback workflow with health checks',
  'Multi-environment deployment pipeline'
];

export const NODE_CATEGORY_LABELS: Record<string, string> = {
  builtin: 'API Tool',
  'sandbox-cli': 'CLI Tool',
  'sandbox-container': 'Container',
  script: 'Script',
  agent: 'AI Agent'
};

export const NODE_CATEGORY_COLORS: Record<string, string> = {
  builtin: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
  'sandbox-cli': 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
  'sandbox-container': 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
  script: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
  agent: 'bg-pink-500/10 text-pink-600 dark:text-pink-400'
};

export const NODE_STATUS_BORDERS: Record<string, string> = {
  idle: 'border-border',
  running: 'border-primary/60',
  success: 'border-primary/40',
  failed: 'border-destructive/60',
  suspended: 'border-primary/40',
  skipped: 'border-border'
};

export const NODE_CATEGORY_DOTS: Record<string, string> = {
  builtin: 'bg-blue-500',
  'sandbox-cli': 'bg-amber-500',
  'sandbox-container': 'bg-purple-500',
  script: 'bg-emerald-500'
};

export function resetNodesExecutionStatus<T extends { id: string; data?: Record<string, unknown> }>(
  nodes: T[]
): T[] {
  return nodes.map((n) => ({
    ...n,
    data: {
      ...n.data,
      executionStatus: 'idle',
      executionOutput: undefined,
      executionError: undefined,
      executionDuration: undefined
    }
  }));
}

export function setPaletteItemDragData(
  event: DragEvent,
  item: { type: string; defaultData: object }
): void {
  event.dataTransfer.setData('application/reactflow-type', item.type);
  event.dataTransfer.setData('application/reactflow-data', JSON.stringify(item.defaultData));
  event.dataTransfer.effectAllowed = 'move';
}

export type ExecutionMessageIconType =
  | 'play'
  | 'loader'
  | 'loader-muted'
  | 'check'
  | 'x'
  | 'terminal';

export interface ExecutionMessageDisplay {
  iconType: ExecutionMessageIconType;
  label: string;
  badgeVariant: 'default' | 'secondary' | 'destructive' | 'outline';
  spin?: boolean;
}

export function getExecutionMessageDisplay(message: ExecutionMessage): ExecutionMessageDisplay {
  const stepLabel = message.stepLabel || message.stepId || 'step';
  switch (message.kind) {
    case 'workflow-start':
      return { iconType: 'play', label: 'Workflow started', badgeVariant: 'secondary' };
    case 'step-start':
      return {
        iconType: 'loader',
        label: `Running: ${stepLabel}`,
        badgeVariant: 'secondary',
        spin: true
      };
    case 'step-progress':
      return {
        iconType: 'loader-muted',
        label: stepLabel || 'Progress',
        badgeVariant: 'secondary',
        spin: true
      };
    case 'step-result':
      return { iconType: 'check', label: `${stepLabel} completed`, badgeVariant: 'default' };
    case 'step-error':
      return { iconType: 'x', label: `${stepLabel} failed`, badgeVariant: 'destructive' };
    case 'workflow-complete':
      return { iconType: 'check', label: 'Workflow completed', badgeVariant: 'default' };
    case 'workflow-error':
      return { iconType: 'x', label: 'Workflow failed', badgeVariant: 'destructive' };
    default:
      return { iconType: 'terminal', label: 'Event', badgeVariant: 'secondary' };
  }
}
