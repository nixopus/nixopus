import { Thread } from '@/redux/services/agents/agentsApi';

export type ContentBlock =
  | { type: 'text'; content: string }
  | { type: 'tool-call'; toolName: string; toolCallId: string; arguments: Record<string, any> }
  | { type: 'tool-result'; toolCallId: string; result: any; isError?: boolean };

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  blocks?: ContentBlock[];
  timestamp: Date;
}

export interface AISheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export interface AIContentProps {
  open: boolean;
  className?: string;
  threadId?: string | null;
  onThreadChange?: (threadId: string) => void;
}

export interface EmptyStateProps {
  onSuggestionClick: (text: string) => void;
}

export const SUGGESTION_KEYS = [
  'ai.suggestions.deploy',
  'ai.suggestions.logs',
  'ai.suggestions.envVars'
] as const;

export interface SuggestionChipProps {
  text: string;
  onClick: () => void;
}

export interface ToolCallProps {
  toolName: string;
  toolCallId: string;
  arguments?: Record<string, any>;
  result?: any;
  isError?: boolean;
  isComplete?: boolean;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

export interface MessageBubbleProps {
  message: Message;
  isStreaming?: boolean;
  isLastMessage?: boolean;
}

export interface ThreadsSidebarProps {
  selectedThreadId: string | null;
  onThreadSelect: (threadId: string | null) => void;
  onNewThread: () => void;
  className?: string;
}

export interface ThreadItemProps {
  thread: Thread;
  isSelected: boolean;
  onSelect: (threadId: string) => void;
  formatDate: (dateString: string) => string;
}
