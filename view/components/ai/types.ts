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
