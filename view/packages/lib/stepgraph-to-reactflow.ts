import type { WorkflowNode, WorkflowEdge, WorkflowNodeType } from '@/packages/types/workflow';

interface SerializedStep {
  id: string;
  description?: string;
  metadata?: Record<string, unknown>;
  component?: string;
  serializedStepFlow?: StepGraphEntry[];
  canSuspend?: boolean;
}

interface StepEntry {
  type: 'step';
  step: SerializedStep;
}

interface SleepEntry {
  type: 'sleep' | 'sleepUntil';
  id: string;
  duration?: number;
}

interface ParallelEntry {
  type: 'parallel';
  steps: StepEntry[];
}

interface ConditionalEntry {
  type: 'conditional';
  steps: StepEntry[];
  serializedConditions: { id: string; fn: string }[];
}

interface LoopEntry {
  type: 'loop';
  step: SerializedStep;
  serializedCondition: { id: string; fn: string };
  loopType: 'dowhile' | 'dountil';
}

interface ForeachEntry {
  type: 'foreach';
  step: SerializedStep;
  opts: { concurrency: number };
}

type StepGraphEntry =
  | StepEntry
  | SleepEntry
  | ParallelEntry
  | ConditionalEntry
  | LoopEntry
  | ForeachEntry;

interface StepInfo {
  id: string;
  description: string;
  metadata?: Record<string, unknown>;
}

const NODE_WIDTH = 160;
const NODE_GAP_Y = 100;
const NODE_GAP_X = 200;

function resolveNodeType(step: SerializedStep): WorkflowNodeType {
  if (step.canSuspend) return 'approval';
  if (step.component) return 'agent';
  return 'tool';
}

function makeNodeData(
  step: SerializedStep,
  stepsInfo: Record<string, StepInfo>,
  nodeType: WorkflowNodeType
) {
  const info = stepsInfo[step.id];
  const label = info?.description || step.description || step.id;

  const base = {
    label,
    executionStatus: 'idle' as const,
    description: step.id
  };

  switch (nodeType) {
    case 'agent':
      return { ...base, agentId: step.id, agentName: label };
    case 'approval':
      return { ...base, approvalMessage: 'Approval required' };
    case 'condition':
      return { ...base, conditionExpression: '' };
    default:
      return { ...base, toolId: step.id, toolName: label };
  }
}

interface LayoutState {
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  y: number;
  prevNodeIds: string[];
}

function addStepNode(
  state: LayoutState,
  step: SerializedStep,
  stepsInfo: Record<string, StepInfo>,
  x: number
): string {
  const nodeType = resolveNodeType(step);
  const nodeId = `step-${step.id}`;

  state.nodes.push({
    id: nodeId,
    type: nodeType,
    position: { x, y: state.y },
    data: makeNodeData(step, stepsInfo, nodeType) as any
  });

  return nodeId;
}

function connectNodes(state: LayoutState, sourceIds: string[], targetIds: string[]) {
  for (const src of sourceIds) {
    for (const tgt of targetIds) {
      state.edges.push({
        id: `e-${src}-${tgt}`,
        source: src,
        target: tgt
      });
    }
  }
}

function processEntry(
  entry: StepGraphEntry,
  state: LayoutState,
  stepsInfo: Record<string, StepInfo>,
  centerX: number
) {
  switch (entry.type) {
    case 'step': {
      const nodeId = addStepNode(state, entry.step, stepsInfo, centerX);
      connectNodes(state, state.prevNodeIds, [nodeId]);
      state.prevNodeIds = [nodeId];
      state.y += NODE_GAP_Y;
      break;
    }

    case 'sleep':
    case 'sleepUntil': {
      const nodeId = `sleep-${entry.id}`;
      const isSleepUntil = entry.type === 'sleepUntil';
      state.nodes.push({
        id: nodeId,
        type: 'sleep' as WorkflowNodeType,
        position: { x: centerX, y: state.y },
        data: {
          label: isSleepUntil ? 'Sleep Until' : `Delay ${entry.duration ?? ''}ms`,
          sleepType: isSleepUntil ? 'until' : 'duration',
          duration: entry.duration,
          executionStatus: 'idle' as const
        } as any
      });
      connectNodes(state, state.prevNodeIds, [nodeId]);
      state.prevNodeIds = [nodeId];
      state.y += NODE_GAP_Y;
      break;
    }

    case 'parallel': {
      const totalWidth = (entry.steps.length - 1) * NODE_GAP_X;
      const startX = centerX - totalWidth / 2;
      const parallelIds: string[] = [];
      const parallelY = state.y;

      for (let i = 0; i < entry.steps.length; i++) {
        const x = startX + i * NODE_GAP_X;
        const s = entry.steps[i]!;
        const nodeId = addStepNode(
          { ...state, y: parallelY, nodes: state.nodes, edges: state.edges, prevNodeIds: [] },
          s.step,
          stepsInfo,
          x
        );
        state.nodes[state.nodes.length - 1]!.position.y = parallelY;
        connectNodes(state, state.prevNodeIds, [nodeId]);
        parallelIds.push(nodeId);
      }

      state.prevNodeIds = parallelIds;
      state.y = parallelY + NODE_GAP_Y;
      break;
    }

    case 'conditional': {
      const condId = `cond-${entry.serializedConditions[0]?.id || Date.now()}`;
      state.nodes.push({
        id: condId,
        type: 'condition' as WorkflowNodeType,
        position: { x: centerX, y: state.y },
        data: {
          label: 'Condition',
          conditionExpression: '',
          executionStatus: 'idle' as const
        } as any
      });
      connectNodes(state, state.prevNodeIds, [condId]);
      state.y += NODE_GAP_Y;

      const totalWidth = (entry.steps.length - 1) * NODE_GAP_X;
      const startX = centerX - totalWidth / 2;
      const branchIds: string[] = [];
      const branchY = state.y;

      for (let i = 0; i < entry.steps.length; i++) {
        const x = startX + i * NODE_GAP_X;
        const s = entry.steps[i]!;
        const nodeId = addStepNode(
          { ...state, y: branchY, nodes: state.nodes, edges: state.edges, prevNodeIds: [] },
          s.step,
          stepsInfo,
          x
        );
        state.nodes[state.nodes.length - 1]!.position.y = branchY;
        connectNodes(state, [condId], [nodeId]);
        branchIds.push(nodeId);
      }

      state.prevNodeIds = branchIds;
      state.y = branchY + NODE_GAP_Y;
      break;
    }

    case 'loop': {
      const loopId = `loop-${entry.step.id}`;
      state.nodes.push({
        id: loopId,
        type: 'loop' as WorkflowNodeType,
        position: { x: centerX, y: state.y },
        data: {
          label: `${entry.loopType === 'dowhile' ? 'While' : 'Until'}: ${entry.step.description || entry.step.id}`,
          loopType: entry.loopType,
          conditionExpression: entry.serializedCondition?.fn || '',
          executionStatus: 'idle' as const
        } as any
      });
      connectNodes(state, state.prevNodeIds, [loopId]);
      state.y += NODE_GAP_Y;

      const bodyId = addStepNode(state, entry.step, stepsInfo, centerX);
      connectNodes(state, [loopId], [bodyId]);
      state.prevNodeIds = [bodyId];
      state.y += NODE_GAP_Y;
      break;
    }

    case 'foreach': {
      const foreachId = `foreach-${entry.step.id}`;
      state.nodes.push({
        id: foreachId,
        type: 'foreach' as WorkflowNodeType,
        position: { x: centerX, y: state.y },
        data: {
          label: `For Each (×${entry.opts?.concurrency || 1})`,
          concurrency: entry.opts?.concurrency || 1,
          executionStatus: 'idle' as const
        } as any
      });
      connectNodes(state, state.prevNodeIds, [foreachId]);
      state.y += NODE_GAP_Y;

      const bodyId = addStepNode(state, entry.step, stepsInfo, centerX);
      connectNodes(state, [foreachId], [bodyId]);
      state.prevNodeIds = [bodyId];
      state.y += NODE_GAP_Y;
      break;
    }
  }
}

export function stepGraphToReactFlow(
  stepGraph: any[],
  stepsInfo: Record<string, StepInfo>
): { nodes: WorkflowNode[]; edges: WorkflowEdge[] } {
  if (!stepGraph || stepGraph.length === 0) {
    return { nodes: [], edges: [] };
  }

  const centerX = 300;

  const triggerNode: WorkflowNode = {
    id: 'trigger-start',
    type: 'trigger',
    position: { x: centerX, y: 0 },
    data: {
      label: 'Start',
      triggerType: 'manual',
      executionStatus: 'idle',
      config: {}
    } as any
  };

  const state: LayoutState = {
    nodes: [triggerNode],
    edges: [],
    y: NODE_GAP_Y,
    prevNodeIds: ['trigger-start']
  };

  for (const entry of stepGraph as StepGraphEntry[]) {
    processEntry(entry, state, stepsInfo, centerX);
  }

  return { nodes: state.nodes, edges: state.edges };
}
