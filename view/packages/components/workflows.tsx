'use client';

import React, { memo, useCallback, forwardRef, useRef, useState } from 'react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { useTheme } from 'next-themes';
import {
  ReactFlow,
  Background,
  Controls,
  Panel as FlowPanel,
  useNodesState,
  useEdgesState,
  addEdge,
  type Connection,
  type Edge,
  type NodeMouseHandler,
  BackgroundVariant,
  type ReactFlowInstance
} from '@xyflow/react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps, NodeTypes } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import {
  Panel,
  PanelGroup,
  PanelResizeHandle,
  type ImperativePanelHandle
} from 'react-resizable-panels';
import { ReactFlowProvider } from '@xyflow/react';
import { Streamdown } from 'streamdown';
import {
  Button,
  Badge,
  Skeleton,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  TooltipProvider
} from '@nixopus/ui';
import {
  Play,
  Square,
  RotateCcw,
  X,
  Loader2,
  Check,
  Pause,
  SkipForward,
  Send,
  User,
  Terminal,
  ChevronDown,
  ChevronRight,
  Plus,
  Workflow,
  AlertCircle,
  Trash2,
  PanelLeftClose,
  PanelLeftOpen,
  Sparkles
} from 'lucide-react';
import * as ContextMenu from '@radix-ui/react-context-menu';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useWorkflows,
  useWorkflowEditor,
  usePanelToggle,
  useNodeDescription,
  useNodePalette,
  useStreamingPlaceholder,
  useChatTimeline,
  useScrollToBottom,
  type PlannerMessage
} from '@/packages/hooks/workflows';
import type {
  WorkflowNode,
  WorkflowEdge,
  WorkflowCanvasHandle,
  WorkflowRunStatus,
  BaseNodeData,
  NodeExecutionStatus,
  ToolNodeData,
  ExecutionMessage,
  NodePaletteItem
} from '@/packages/types/workflow';
import {
  sanitizeWorkflowLabel,
  NODE_CATEGORY_LABELS,
  NODE_CATEGORY_COLORS,
  NODE_STATUS_BORDERS,
  NODE_CATEGORY_DOTS,
  createNodeFromPaletteDrop,
  resetNodesExecutionStatus,
  setPaletteItemDragData,
  stripJsonBlocks,
  looksLikeWorkflowJson,
  splitContentAroundJson,
  extractReadableText,
  getRemainingOutputData,
  CHAT_SUGGESTIONS,
  getExecutionMessageDisplay
} from '@/packages/lib/workflow-utils';
import { DeleteDialog } from '@/components/ui/delete-dialog';

function StatusIndicator({ status }: { status: NodeExecutionStatus }) {
  switch (status) {
    case 'running':
      return <Loader2 className="h-3 w-3 text-primary animate-spin" />;
    case 'success':
      return <Check className="h-3 w-3 text-primary" />;
    case 'failed':
      return <X className="h-3 w-3 text-destructive" />;
    case 'suspended':
      return <Pause className="h-3 w-3 text-primary" />;
    case 'skipped':
      return <SkipForward className="h-3 w-3 text-muted-foreground" />;
    default:
      return null;
  }
}

function BaseNodeComponent({
  children,
  data,
  selected,
  hasInput = true,
  hasOutput = true
}: {
  children?: React.ReactNode;
  data: BaseNodeData;
  selected?: boolean;
  hasInput?: boolean;
  hasOutput?: boolean;
}) {
  const status = data.executionStatus || 'idle';
  return (
    <div
      className={cn(
        'relative rounded-md border bg-card text-card-foreground shadow-sm w-[160px] transition-all',
        NODE_STATUS_BORDERS[status],
        selected && 'ring-1 ring-primary/50 ring-offset-1 ring-offset-background',
        status === 'running' && 'shadow-md'
      )}
    >
      {hasInput && (
        <Handle
          type="target"
          position={Position.Top}
          className="!bg-muted-foreground/60 !w-2 !h-2 !border-2 !border-background !-top-1"
        />
      )}
      <div className="px-3 py-2.5 flex items-center gap-2">
        <p className="text-xs font-medium truncate flex-1">{data.label}</p>
        <StatusIndicator status={status} />
      </div>
      {children && <div className="px-3 pb-2">{children}</div>}
      {hasOutput && (
        <Handle
          type="source"
          position={Position.Bottom}
          className="!bg-muted-foreground/60 !w-2 !h-2 !border-2 !border-background !-bottom-1"
        />
      )}
    </div>
  );
}

const BaseNode = memo(BaseNodeComponent);

function createStandardNode(hasInput = true, hasOutput = true) {
  const StandardNode = memo(function StandardNode({ data, selected }: NodeProps) {
    return (
      <BaseNode
        data={data as BaseNodeData}
        selected={selected}
        hasInput={hasInput}
        hasOutput={hasOutput}
      />
    );
  });
  return StandardNode;
}

const TriggerNode = createStandardNode(false);
const AgentNode = createStandardNode();
const ConditionNode = createStandardNode();
const ApprovalNode = createStandardNode();
const NotificationNode = createStandardNode();
const ForeachNode = createStandardNode();
const LoopNode = createStandardNode();
const MapNode = createStandardNode();
const SleepNode = createStandardNode();

function ToolNodeComponent({ data, selected }: NodeProps) {
  const nodeData = data as unknown as ToolNodeData;
  const category = nodeData.toolCategory ?? 'builtin';
  const isSandboxed = category !== 'builtin';
  return (
    <BaseNode data={nodeData} selected={selected}>
      {isSandboxed && (
        <div className="flex items-center gap-1.5 -mt-0.5">
          <span
            className={cn(
              'w-1.5 h-1.5 rounded-full',
              NODE_CATEGORY_DOTS[category] ?? 'bg-muted-foreground'
            )}
          />
          <span className="text-[10px] text-muted-foreground truncate">
            {nodeData.toolName || nodeData.toolId}
          </span>
        </div>
      )}
    </BaseNode>
  );
}

const ToolNode = memo(ToolNodeComponent);

const workflowNodeTypes: NodeTypes = {
  trigger: TriggerNode,
  agent: AgentNode,
  tool: ToolNode,
  condition: ConditionNode,
  approval: ApprovalNode,
  notification: NotificationNode,
  foreach: ForeachNode,
  loop: LoopNode,
  map: MapNode,
  sleep: SleepNode
};

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start gap-2">
      <span className="text-[11px] font-medium text-muted-foreground whitespace-nowrap">
        {label}
      </span>
      <span className="text-xs text-foreground">{value}</span>
    </div>
  );
}

export function NodeDetailPanel({
  node,
  headers,
  onClose
}: {
  node: { id: string; type: string; data: Record<string, unknown> } | null;
  headers?: Record<string, string>;
  onClose: () => void;
}) {
  const { result, isLoading } = useNodeDescription({ node, headers });
  if (!node) return null;
  const category = result?.execution?.category ?? (node.data.toolCategory as string) ?? node.type;
  return (
    <div className="absolute right-3 top-14 z-20 w-80 bg-card border border-border rounded-lg shadow-lg overflow-hidden animate-in slide-in-from-right-2 duration-200">
      <div className="flex items-center justify-between px-3 py-2 border-b border-border">
        <span className="text-sm font-medium truncate min-w-0">
          {(node.data.label as string) || node.type}
        </span>
        <Button variant="ghost" size="icon" className="h-6 w-6 flex-shrink-0" onClick={onClose}>
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>
      <div className="px-3 py-3 space-y-3 max-h-[400px] overflow-y-auto">
        {isLoading && (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          </div>
        )}
        {!isLoading && result && (
          <>
            <p className="text-sm text-foreground leading-snug">
              {sanitizeWorkflowLabel(result.summary)}
            </p>
            <div className="flex flex-wrap gap-1.5">
              <Badge
                variant="secondary"
                className={cn('text-[10px] px-1.5 py-0', NODE_CATEGORY_COLORS[category])}
              >
                {NODE_CATEGORY_LABELS[category] || category}
              </Badge>
            </div>
            <div className="space-y-2">
              <DetailRow label="Runs in" value={sanitizeWorkflowLabel(result.execution.runsIn)} />
              {result.execution.command && (
                <div className="space-y-1">
                  <span className="text-[11px] font-medium text-muted-foreground">Command</span>
                  <code className="block text-[11px] bg-muted/50 rounded px-2 py-1.5 font-mono break-all">
                    {result.execution.command}
                  </code>
                </div>
              )}
              {result.execution.sandboxConfig && (
                <div className="space-y-1.5">
                  <span className="text-[11px] font-medium text-muted-foreground">Isolation</span>
                  <div className="flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                    <span>
                      {result.execution.sandboxConfig.network
                        ? 'Network allowed'
                        : 'No network access'}
                    </span>
                    <span>Timeout: {result.execution.sandboxConfig.timeout}</span>
                    <span>{result.execution.sandboxConfig.isolation}</span>
                  </div>
                </div>
              )}
              {result.explanation && (
                <div className="space-y-1">
                  <span className="text-[11px] font-medium text-muted-foreground">Details</span>
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    {sanitizeWorkflowLabel(result.explanation)}
                  </p>
                </div>
              )}
            </div>
          </>
        )}
        {!isLoading && !result && (
          <p className="text-xs text-muted-foreground py-4 text-center">
            No execution details available for this node.
          </p>
        )}
      </div>
    </div>
  );
}

function DraggableItem({ item }: { item: NodePaletteItem }) {
  const onDragStart = (event: React.DragEvent) => setPaletteItemDragData(event, item);
  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            draggable
            onDragStart={onDragStart}
            className="flex items-center justify-center px-2 h-8 rounded-md border border-transparent hover:border-border hover:bg-accent cursor-grab active:cursor-grabbing transition-colors"
          >
            <span className="text-[11px] font-medium text-muted-foreground truncate max-w-[80px]">
              {item.label}
            </span>
          </div>
        </TooltipTrigger>
        <TooltipContent side="right" className="text-xs">
          <p className="font-medium">{item.label}</p>
          <p className="text-muted-foreground">{item.description}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export function NodePalette({ className }: { className?: string }) {
  const [expandedCat, setExpandedCat] = useState<string | null>('steps');
  const { items: allItems, categories, isLoadingTools } = useNodePalette();
  return (
    <div className={cn('h-full flex flex-col', className)}>
      {categories.map((cat) => {
        const items = allItems.filter((i) => i.category === cat.key);
        if (cat.key === 'ci-cd' && items.length === 0 && !isLoadingTools) return null;
        const isOpen = expandedCat === cat.key;
        return (
          <div key={cat.key}>
            <button
              onClick={() => setExpandedCat(isOpen ? null : cat.key)}
              className="flex items-center justify-between w-full px-3 py-2 text-[11px] font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              {cat.label}
              <ChevronDown className={cn('h-3 w-3 transition-transform', isOpen && 'rotate-180')} />
            </button>
            {isOpen && (
              <div className="flex flex-wrap gap-1 px-2 pb-2">
                {cat.key === 'ci-cd' && isLoadingTools ? (
                  <div className="flex items-center justify-center w-full py-2">
                    <Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground" />
                  </div>
                ) : (
                  items.map((item, idx) => (
                    <DraggableItem
                      key={`${item.type}-${(item.defaultData as Record<string, unknown>)?.toolId ?? idx}`}
                      item={item}
                    />
                  ))
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

function CanvasToolbarButton({
  tooltip,
  children,
  ...props
}: { tooltip: string } & React.ComponentProps<typeof Button>) {
  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button size="icon" variant="secondary" className="h-8 w-8 shadow-sm" {...props}>
            {children}
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom" className="text-xs">
          {tooltip}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export const WorkflowCanvas = forwardRef<
  WorkflowCanvasHandle,
  {
    initialNodes?: WorkflowNode[];
    initialEdges?: WorkflowEdge[];
    readonly?: boolean;
    toolbar?: {
      runStatus: WorkflowRunStatus;
      onRun: () => void;
      onCancel: () => void;
      onSave: () => void;
      onReset: () => void;
      isDraft?: boolean;
      isSaving?: boolean;
    };
    onNodeClick?: (
      node: { id: string; type: string; data: Record<string, unknown> } | null
    ) => void;
    children?: React.ReactNode;
  }
>(function WorkflowCanvas(
  { initialNodes = [], initialEdges = [], readonly = false, toolbar, onNodeClick, children },
  ref
) {
  const { resolvedTheme } = useTheme();
  const reactFlowInstance = useRef<ReactFlowInstance | null>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes as any);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  React.useImperativeHandle(
    ref,
    () => ({
      getNodes: () => nodes,
      getEdges: () => edges,
      setNodeData: (nodeId: string, data: Record<string, unknown>) => {
        setNodes((prev: any[]) =>
          prev.map((n) => (n.id === nodeId ? { ...n, data: { ...n.data, ...data } } : n))
        );
      },
      resetAllNodes: () => setNodes((prev: any[]) => resetNodesExecutionStatus(prev)),
      replaceGraph: (newNodes: any[], newEdges: Edge[]) => {
        setNodes(newNodes as any);
        setEdges(newEdges);
        setTimeout(() => reactFlowInstance.current?.fitView({ padding: 0.3 }), 50);
      }
    }),
    [nodes, edges, setNodes, setEdges]
  );

  const onConnect = useCallback(
    (params: Connection) => {
      if (readonly) return;
      setEdges((eds: Edge[]) => addEdge({ ...params, animated: true }, eds));
    },
    [setEdges, readonly]
  );

  const onInit = useCallback((instance: ReactFlowInstance) => {
    reactFlowInstance.current = instance;
    instance.fitView({ padding: 0.3 });
  }, []);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      if (readonly || !reactFlowInstance.current) return;
      event.preventDefault();
      const newNode = createNodeFromPaletteDrop(event, (pos) =>
        reactFlowInstance.current!.screenToFlowPosition(pos)
      );
      if (!newNode) return;
      setNodes((nds: any[]) => [...nds, newNode]);
    },
    [setNodes, readonly]
  );

  const handleNodeClick: NodeMouseHandler = useCallback(
    (_event, node) => {
      onNodeClick?.({
        id: node.id,
        type: (node.type as string) || 'tool',
        data: node.data as Record<string, unknown>
      });
    },
    [onNodeClick]
  );

  const handlePaneClick = useCallback(() => onNodeClick?.(null as any), [onNodeClick]);

  return (
    <div className="h-full w-full relative">
      <ReactFlow
        proOptions={{ hideAttribution: true }}
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onInit={onInit}
        onDragOver={onDragOver}
        onDrop={onDrop}
        onNodeClick={handleNodeClick}
        onPaneClick={handlePaneClick}
        nodeTypes={workflowNodeTypes}
        fitView
        colorMode={resolvedTheme === 'dark' ? 'dark' : 'light'}
        deleteKeyCode={readonly ? null : 'Delete'}
        className="bg-background"
        defaultEdgeOptions={{ animated: false, style: { strokeWidth: 1.5 } }}
      >
        <Background
          variant={BackgroundVariant.Dots}
          gap={20}
          size={0.5}
          color="hsl(var(--muted-foreground) / 0.15)"
        />
        <Controls
          showInteractive={false}
          className="!bg-card !border-border !shadow-none !rounded-md"
        />
        {toolbar && (
          <FlowPanel position="top-right" className="m-3 flex items-center gap-1">
            {toolbar.runStatus === 'running' ? (
              <CanvasToolbarButton tooltip="Cancel" onClick={toolbar.onCancel}>
                <Square className="h-3.5 w-3.5 text-destructive" />
              </CanvasToolbarButton>
            ) : (
              <CanvasToolbarButton
                tooltip={toolbar.isDraft ? 'Save workflow first' : 'Run'}
                onClick={toolbar.onRun}
                disabled={!!toolbar.isDraft}
              >
                <Play className="h-3.5 w-3.5" />
              </CanvasToolbarButton>
            )}
            {(toolbar.runStatus === 'success' || toolbar.runStatus === 'failed') && (
              <CanvasToolbarButton tooltip="Reset" onClick={toolbar.onReset}>
                <RotateCcw className="h-3.5 w-3.5" />
              </CanvasToolbarButton>
            )}
          </FlowPanel>
        )}
      </ReactFlow>
      {children}
    </div>
  );
});

function NixopusLogo({ size = 20, className }: { size?: number; className?: string }) {
  const { resolvedTheme } = useTheme();
  const src = resolvedTheme === 'dark' ? '/logo_white.png' : '/logo_black.png';
  return <Image src={src} alt="Nixopus" width={size} height={size} className={className} />;
}

const MessageBubble = memo(function MessageBubble({
  message,
  isStreaming = false,
  isLastMessage = false
}: {
  message: PlannerMessage;
  isStreaming?: boolean;
  isLastMessage?: boolean;
}) {
  const isUser = message.role === 'user';
  const { before, inJsonBlock, after } = splitContentAroundJson(message.content);
  const showJsonIndicator = !isUser && isStreaming && isLastMessage && inJsonBlock;
  const streamingText = useStreamingPlaceholder(showJsonIndicator);
  const displayContent = isUser ? message.content : stripJsonBlocks(message.content);
  const hadWorkflowJson = !isUser && looksLikeWorkflowJson(message.content);
  const hasStrippedOnly = hadWorkflowJson && !displayContent.trim();
  const hasContent =
    displayContent.trim().length > 0 || before.trim().length > 0 || after.trim().length > 0;
  if (!hasContent && !message.graph && !showJsonIndicator && !hasStrippedOnly) return null;
  const showWorkflowIndicator = !!message.graph || hasStrippedOnly;
  const renderAssistantContent = () => {
    const parts: React.ReactNode[] = [];
    if (before.trim())
      parts.push(
        <div
          key="before"
          className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2"
        >
          <Streamdown isAnimating={false}>{before}</Streamdown>
        </div>
      );
    if (showJsonIndicator)
      parts.push(
        <p key="indicator" className="text-primary font-medium mt-1">
          {streamingText}
        </p>
      );
    if (after.trim())
      parts.push(
        <div
          key="after"
          className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2 mt-1"
        >
          <Streamdown isAnimating={false}>{after}</Streamdown>
        </div>
      );
    if (!showJsonIndicator && parts.length === 0 && displayContent.trim())
      parts.push(
        <div
          key="content"
          className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2"
        >
          <Streamdown isAnimating={false}>{displayContent}</Streamdown>
        </div>
      );
    return parts;
  };
  return (
    <div className={cn('flex gap-2', isUser ? 'flex-row-reverse' : 'flex-row')}>
      <div
        className={cn(
          'flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center mt-0.5',
          isUser ? 'bg-primary/10' : 'bg-muted'
        )}
      >
        {isUser ? <User className="h-3 w-3 text-primary" /> : <NixopusLogo size={12} />}
      </div>
      <div
        className={cn(
          'max-w-[85%] rounded-lg px-3 py-2 text-sm leading-relaxed',
          isUser ? 'bg-primary text-primary-foreground' : 'bg-muted/50 text-foreground'
        )}
      >
        {isUser ? (
          <p className="whitespace-pre-wrap break-words">{message.content}</p>
        ) : (
          renderAssistantContent()
        )}
        {showWorkflowIndicator && (
          <div className="mt-2 flex items-center gap-1.5 text-xs text-primary">
            <NixopusLogo size={12} />
            <span>Workflow rendered on canvas</span>
          </div>
        )}
      </div>
    </div>
  );
});

function RawDataOutput({ data }: { data: Record<string, unknown> }) {
  const [expanded, setExpanded] = useState(false);
  const json = JSON.stringify(data, null, 2);
  const isLong = json.length > 120;
  if (isLong) {
    return (
      <>
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground transition-colors"
        >
          {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
          {expanded ? 'Hide details' : 'Show details'}
        </button>
        {expanded && (
          <pre className="mt-1 text-[10px] bg-background/50 rounded px-2 py-1.5 overflow-x-auto max-h-40 font-mono text-muted-foreground">
            {json}
          </pre>
        )}
      </>
    );
  }
  return (
    <pre className="text-[10px] bg-background/50 rounded px-2 py-1.5 overflow-x-auto max-h-40 font-mono text-muted-foreground">
      {json}
    </pre>
  );
}

const ReasoningBubble = memo(function ReasoningBubble({ message }: { message: ExecutionMessage }) {
  return (
    <div className="flex gap-2">
      <div className="flex-shrink-0 w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center mt-0.5">
        <NixopusLogo size={12} />
      </div>
      <div className="flex-1 min-w-0 rounded-lg bg-muted/30 px-3 py-2">
        {message.stepLabel && (
          <div className="flex items-center gap-1.5 mb-1">
            <Loader2 className="h-3 w-3 text-primary animate-spin" />
            <span className="text-[10px] font-medium text-muted-foreground">
              {message.stepLabel}
            </span>
          </div>
        )}
        <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-0.5 text-xs leading-relaxed">
          <Streamdown isAnimating={true}>{message.text || ''}</Streamdown>
        </div>
      </div>
    </div>
  );
});

function ExecutionMessageIcon({
  iconType,
  spin
}: {
  iconType: 'play' | 'loader' | 'loader-muted' | 'check' | 'x' | 'terminal';
  spin?: boolean;
}) {
  const iconClass = 'h-3 w-3';
  switch (iconType) {
    case 'play':
      return <Play className={cn(iconClass, 'text-primary')} />;
    case 'loader':
      return <Loader2 className={cn(iconClass, 'text-primary', spin && 'animate-spin')} />;
    case 'loader-muted':
      return <Loader2 className={cn(iconClass, 'text-muted-foreground', spin && 'animate-spin')} />;
    case 'check':
      return <Check className={cn(iconClass, 'text-emerald-500')} />;
    case 'x':
      return <X className={cn(iconClass, 'text-destructive')} />;
    default:
      return <Terminal className={iconClass} />;
  }
}

const ExecutionMessageBubble = memo(function ExecutionMessageBubble({
  message
}: {
  message: ExecutionMessage;
}) {
  if (message.kind === 'step-reasoning') return <ReasoningBubble message={message} />;
  if (message.kind === 'step-progress' && message.text) {
    return (
      <div className="flex gap-2">
        <div className="flex-shrink-0 w-6 h-6 rounded-full bg-muted/80 flex items-center justify-center mt-0.5">
          <Loader2 className="h-3 w-3 text-muted-foreground animate-spin" />
        </div>
        <div className="flex-1 min-w-0 rounded-lg bg-muted/30 px-3 py-2">
          <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-0.5 prose-headings:my-1 text-xs leading-relaxed">
            <Streamdown isAnimating={false}>{message.text}</Streamdown>
          </div>
        </div>
      </div>
    );
  }
  const display = getExecutionMessageDisplay(message);
  const readableText = message.output ? extractReadableText(message.output) : null;
  const remainingData = message.output ? getRemainingOutputData(message.output) : null;
  return (
    <div className="flex gap-2">
      <div className="flex-shrink-0 w-6 h-6 rounded-full bg-muted/80 flex items-center justify-center mt-0.5">
        <Terminal className="h-3 w-3 text-muted-foreground" />
      </div>
      <div className="flex-1 min-w-0 space-y-1.5">
        <div className="rounded-lg bg-muted/30 border border-border/50 px-3 py-2">
          <div className="flex items-center gap-2">
            <ExecutionMessageIcon iconType={display.iconType} spin={display.spin} />
            <span className="text-xs font-medium truncate">{display.label}</span>
            {(message.kind === 'step-result' ||
              message.kind === 'step-error' ||
              message.kind === 'workflow-complete' ||
              message.kind === 'workflow-error') && (
              <Badge
                variant={display.badgeVariant}
                className="text-[9px] px-1.5 py-0 ml-auto flex-shrink-0"
              >
                {message.kind === 'workflow-complete' || message.kind === 'step-result'
                  ? 'success'
                  : 'failed'}
              </Badge>
            )}
          </div>
          {message.error && (
            <p className="mt-1 text-[11px] text-destructive font-mono">
              {typeof message.error === 'string' ? message.error : JSON.stringify(message.error)}
            </p>
          )}
          {remainingData && (
            <div className="mt-1">
              <RawDataOutput data={remainingData} />
            </div>
          )}
        </div>
        {readableText && (
          <div className="text-xs text-foreground/80 leading-relaxed pl-0.5 prose prose-sm dark:prose-invert max-w-none prose-p:my-0.5 prose-headings:my-1 prose-table:text-xs">
            <Streamdown isAnimating={false}>{readableText}</Streamdown>
          </div>
        )}
      </div>
    </div>
  );
});

function ChatListSkeleton() {
  return (
    <div className="px-3 py-4 space-y-4">
      <div className="flex gap-2">
        <Skeleton className="h-6 w-6 shrink-0 rounded-full" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-[85%]" />
          <Skeleton className="h-4 w-2/3" />
        </div>
      </div>
      <div className="flex gap-2 flex-row-reverse">
        <Skeleton className="h-6 w-6 shrink-0 rounded-full" />
        <Skeleton className="h-10 w-32 rounded-lg" />
      </div>
      <div className="flex gap-2">
        <Skeleton className="h-6 w-6 shrink-0 rounded-full" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-[90%]" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
        </div>
      </div>
    </div>
  );
}

export function WorkflowChat({
  messages,
  executionMessages = [],
  inputValue,
  isStreaming,
  isRunning = false,
  scrollRef,
  onInputChange,
  onSubmit,
  onKeyDown,
  onStop,
  isLoading = false,
  className
}: {
  messages: PlannerMessage[];
  executionMessages?: ExecutionMessage[];
  inputValue: string;
  isStreaming: boolean;
  isRunning?: boolean;
  scrollRef: React.RefObject<HTMLDivElement | null>;
  onInputChange: (value: string) => void;
  onSubmit: (e?: React.FormEvent) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
  onStop: () => void;
  isLoading?: boolean;
  className?: string;
}) {
  const timeline = useChatTimeline(messages, executionMessages);
  useScrollToBottom(scrollRef, [executionMessages]);
  const isEmpty = timeline.length === 0;
  const lastPlannerMsg = messages[messages.length - 1];
  const lastSplit = lastPlannerMsg ? splitContentAroundJson(lastPlannerMsg.content) : null;
  const lastInJsonBlock =
    isStreaming && lastPlannerMsg?.role === 'assistant' && lastSplit?.inJsonBlock === true;
  return (
    <div className={cn('flex flex-col h-full', className)}>
      <div ref={scrollRef} className="flex-1 overflow-y-auto no-scrollbar px-3 py-4 space-y-3">
        {isLoading ? (
          <ChatListSkeleton />
        ) : isEmpty ? (
          <div className="flex flex-col items-center justify-center h-full gap-4 text-center px-2">
            <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
              <NixopusLogo size={20} />
            </div>
            <div>
              <p className="text-sm font-medium">Workflow Planner</p>
              <p className="text-xs text-muted-foreground mt-1">
                Describe what you want to automate
              </p>
            </div>
            <div className="flex flex-col gap-1.5 w-full">
              {CHAT_SUGGESTIONS.map((s) => (
                <button
                  key={s}
                  onClick={() => {
                    onInputChange(s);
                    setTimeout(() => onSubmit(), 0);
                  }}
                  className="text-left text-xs px-3 py-2 rounded-md border border-border hover:bg-accent transition-colors text-muted-foreground hover:text-foreground"
                >
                  {s}
                </button>
              ))}
            </div>
          </div>
        ) : (
          timeline.map((item) => {
            if (item.kind === 'planner') {
              const isLast = isStreaming && item.idx === messages.length - 1;
              return (
                <MessageBubble
                  key={`p-${item.msg.id}`}
                  message={item.msg}
                  isStreaming={isStreaming}
                  isLastMessage={isLast}
                />
              );
            }
            return <ExecutionMessageBubble key={`e-${item.msg.id}`} message={item.msg} />;
          })
        )}
        {isStreaming && !lastInJsonBlock && (
          <div className="flex items-center gap-2 text-xs text-muted-foreground px-8">
            <div className="flex gap-0.5">
              <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground/40 animate-bounce [animation-delay:0ms]" />
              <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground/40 animate-bounce [animation-delay:150ms]" />
              <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground/40 animate-bounce [animation-delay:300ms]" />
            </div>
          </div>
        )}
      </div>
      <form onSubmit={onSubmit} className="p-3">
        <div className="relative">
          <textarea
            value={inputValue}
            onChange={(e) => onInputChange(e.target.value)}
            onKeyDown={onKeyDown}
            placeholder="Describe your workflow..."
            rows={2}
            className="w-full resize-none rounded-lg border border-border bg-background px-3 py-2 pr-10 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
            disabled={isStreaming || isRunning}
          />
          <button
            type={isStreaming ? 'button' : 'submit'}
            onClick={isStreaming ? onStop : undefined}
            className="absolute right-2 bottom-2 p-1.5 rounded-md hover:bg-accent transition-colors text-muted-foreground hover:text-foreground disabled:opacity-50"
            disabled={!isStreaming && !inputValue.trim()}
          >
            {isStreaming ? <Square className="h-3.5 w-3.5" /> : <Send className="h-3.5 w-3.5" />}
          </button>
        </div>
      </form>
    </div>
  );
}

export function WorkflowEditor({
  applicationId,
  workflowId,
  workflowName: initialName,
  initialNodes = [],
  initialEdges = [],
  initialPlanningMessages = [],
  initialExecutionMessages,
  chatThreadId,
  isDraft = false,
  isLoadingMessages = false
}: {
  applicationId: string;
  workflowId: string;
  workflowName: string;
  initialNodes?: WorkflowNode[];
  initialEdges?: WorkflowEdge[];
  initialPlanningMessages?: { role: 'user' | 'assistant'; content: string; graph?: any }[];
  initialExecutionMessages?: ExecutionMessage[];
  chatThreadId?: string | null;
  isDraft?: boolean;
  isLoadingMessages?: boolean;
}) {
  const canvasRef = useRef<WorkflowCanvasHandle | null>(null);
  const sidebarPanelRef = useRef<ImperativePanelHandle | null>(null);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const {
    planner,
    runner,
    isSaving,
    isStillDraft,
    selectedNode,
    setSelectedNode,
    handleSave,
    handleRun,
    handleReset,
    handleNodeClick,
    authHeaders
  } = useWorkflowEditor({
    canvasRef,
    applicationId,
    workflowId,
    workflowName: initialName,
    initialNodes,
    initialEdges,
    initialPlanningMessages,
    initialExecutionMessages,
    chatThreadId,
    isDraft
  });
  const { toggle: toggleSidebar, expand: expandSidebar } = usePanelToggle(sidebarPanelRef);
  return (
    <div className="flex flex-col h-full">
      <PanelGroup direction="horizontal" className="flex-1" autoSaveId="workflow-editor-layout">
        <Panel
          ref={sidebarPanelRef}
          defaultSize={25}
          minSize={15}
          maxSize={50}
          collapsible
          collapsedSize={0}
          onCollapse={() => setSidebarCollapsed(true)}
          onExpand={() => setSidebarCollapsed(false)}
          className="flex flex-col min-w-0"
        >
          <div className="flex items-center justify-end p-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={toggleSidebar}
              title={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            >
              <PanelLeftClose className="h-3.5 w-3.5" />
            </Button>
          </div>
          <div className="flex-1 min-h-0 overflow-hidden">
            <WorkflowChat
              messages={planner.messages}
              executionMessages={runner.executionMessages}
              inputValue={planner.inputValue}
              isStreaming={planner.isStreaming}
              isRunning={runner.isRunning}
              scrollRef={planner.scrollRef}
              onInputChange={planner.setInputValue}
              onSubmit={planner.handleSubmit}
              onKeyDown={planner.handleKeyDown}
              onStop={planner.stopStreaming}
              isLoading={isLoadingMessages || planner.isLoadingHistory}
            />
          </div>
        </Panel>
        <PanelResizeHandle className="w-px flex-shrink-0 bg-border hover:bg-primary/20 data-[resize-handle-active]:bg-primary/30 transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus:outline-none" />
        <Panel defaultSize={75} minSize={30} className="flex flex-col min-w-0 relative">
          {sidebarCollapsed && (
            <Button
              variant="outline"
              size="icon"
              className="absolute left-2 top-14 z-10 h-8 w-8 shadow-md"
              onClick={expandSidebar}
              title="Expand sidebar"
            >
              <PanelLeftOpen className="h-4 w-4" />
            </Button>
          )}
          <ReactFlowProvider>
            <WorkflowCanvas
              ref={canvasRef}
              initialNodes={initialNodes}
              initialEdges={initialEdges}
              readonly={runner.isRunning}
              onNodeClick={handleNodeClick}
              toolbar={{
                runStatus: runner.status,
                onRun: handleRun,
                onCancel: runner.cancelRun,
                onSave: handleSave,
                onReset: handleReset,
                isDraft: isStillDraft,
                isSaving: isSaving
              }}
            >
              {selectedNode && (
                <NodeDetailPanel
                  node={selectedNode}
                  headers={authHeaders()}
                  onClose={() => setSelectedNode(null)}
                />
              )}
            </WorkflowCanvas>
          </ReactFlowProvider>
        </Panel>
      </PanelGroup>
    </div>
  );
}

export function WorkflowsList({ applicationId }: { applicationId: string }) {
  const { t } = useTranslation();
  const router = useRouter();
  const { workflows, isLoading, isDeleting, error, isConfigured, deleteWorkflow } = useWorkflows({
    applicationId
  });
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name?: string } | null>(null);
  if (!isConfigured) {
    return (
      <div className="flex h-full w-full items-center justify-center py-16">
        <div className="text-center max-w-md space-y-4 px-4">
          <div className="flex items-center justify-center size-16 rounded-2xl bg-muted mx-auto">
            <Sparkles className="size-8 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-semibold">AI Agent Not Configured</h3>
          <p className="text-sm text-muted-foreground">
            The AI-powered deployment assistant is not enabled on this instance. To get access,
            reach out to us and we&apos;ll help you get set up.
          </p>
          <a
            href="mailto:support@nixopus.com"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
          >
            Contact support@nixopus.com
          </a>
        </div>
      </div>
    );
  }
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-32 rounded-lg" />
        ))}
      </div>
    );
  }
  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 gap-3">
        <AlertCircle className="h-10 w-10 text-destructive" />
        <p className="text-sm text-destructive">{error}</p>
      </div>
    );
  }
  if (workflows.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 gap-4">
        <Workflow className="h-16 w-16 text-muted-foreground/30" />
        <div className="text-center">
          <p className="text-lg font-medium">{t('workflows.emptyState.title')}</p>
          <p className="text-sm text-muted-foreground mt-1">
            {t('workflows.emptyState.description')}
          </p>
        </div>
        <Button
          onClick={() => router.push(`/apps/application/${applicationId}/workflows/new`)}
          className="gap-2"
        >
          <Plus className="h-4 w-4" />
          {t('workflows.createButton')}
        </Button>
      </div>
    );
  }
  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return;
    await deleteWorkflow(deleteTarget.id);
    setDeleteTarget(null);
  };
  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button
          onClick={() => router.push(`/apps/application/${applicationId}/workflows/new`)}
          className="gap-2"
        >
          <Plus className="h-4 w-4" />
          {t('workflows.createButton')}
        </Button>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {workflows.map((wf) => (
          <ContextMenu.Root key={wf.id}>
            <ContextMenu.Trigger asChild>
              <button
                onClick={() => router.push(`/apps/application/${applicationId}/workflows/${wf.id}`)}
                className="text-left p-4 rounded-lg border border-border bg-card hover:bg-accent transition-colors group w-full"
              >
                <div className="flex items-start gap-3">
                  <Workflow className="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
                  <div className="min-w-0">
                    <p className="font-medium text-sm truncate group-hover:text-primary transition-colors">
                      {wf.name || wf.id}
                    </p>
                    {wf.description && (
                      <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                        {wf.description}
                      </p>
                    )}
                  </div>
                </div>
              </button>
            </ContextMenu.Trigger>
            <ContextMenu.Portal>
              <ContextMenu.Content className="z-50 min-w-[8rem] overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md">
                <ContextMenu.Item
                  className={cn(
                    'relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none',
                    'focus:bg-accent focus:text-accent-foreground',
                    'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
                    'text-destructive focus:bg-destructive focus:text-destructive-foreground'
                  )}
                  onSelect={(e) => {
                    e.preventDefault();
                    setDeleteTarget({ id: wf.id, name: wf.name });
                  }}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  {t('workflows.delete')}
                </ContextMenu.Item>
              </ContextMenu.Content>
            </ContextMenu.Portal>
          </ContextMenu.Root>
        ))}
      </div>
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title={t('workflows.deleteConfirm.title')}
        description={t('workflows.deleteConfirm.description')}
        onConfirm={handleDeleteConfirm}
        isDeleting={isDeleting}
        variant="destructive"
        confirmText={t('workflows.delete')}
        icon={Trash2}
      />
    </div>
  );
}

// --- WorkflowToolbar ---

function ToolbarButton({
  tooltip,
  children,
  ...props
}: { tooltip: string } & React.ComponentProps<typeof Button>) {
  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button size="icon" variant="ghost" className="h-7 w-7" {...props}>
            {children}
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom" className="text-xs">
          {tooltip}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export function WorkflowToolbar({
  runStatus,
  onRun,
  onCancel,
  onSave,
  onAiRecommend,
  onReset,
  isSaving,
  isDraft
}: {
  runStatus: WorkflowRunStatus;
  onRun: () => void;
  onCancel: () => void;
  onSave: () => void;
  onAiRecommend?: () => void;
  onReset: () => void;
  isSaving?: boolean;
  isDraft?: boolean;
}) {
  const isRunning = runStatus === 'running';
  const canRun = !isDraft && !isRunning;
  return (
    <div className="flex items-center justify-end gap-0.5 px-3 py-1.5 border-b border-border bg-card/50">
      {onAiRecommend && (
        <ToolbarButton tooltip="AI Recommend" onClick={onAiRecommend} disabled={isRunning}>
          <Sparkles className="h-3.5 w-3.5" />
        </ToolbarButton>
      )}
      {isRunning ? (
        <ToolbarButton tooltip="Cancel" onClick={onCancel}>
          <Square className="h-3.5 w-3.5 text-destructive" />
        </ToolbarButton>
      ) : (
        <ToolbarButton
          tooltip={isDraft ? 'Save workflow first' : 'Run'}
          onClick={onRun}
          disabled={!canRun}
        >
          <Play className="h-3.5 w-3.5" />
        </ToolbarButton>
      )}
      {(runStatus === 'success' || runStatus === 'failed') && (
        <ToolbarButton tooltip="Reset" onClick={onReset}>
          <RotateCcw className="h-3.5 w-3.5" />
        </ToolbarButton>
      )}
    </div>
  );
}
