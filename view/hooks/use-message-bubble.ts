'use client';

import { useMemo, useState, useCallback } from 'react';
import type { Message, ContentBlock } from '../components/ai/types';

export type GroupedBlock =
  | ContentBlock
  | {
      type: 'grouped-tool';
      toolCall: Extract<ContentBlock, { type: 'tool-call' }>;
      toolResult?: Extract<ContentBlock, { type: 'tool-result' }>;
    };

type ToolCallBlock = Extract<ContentBlock, { type: 'tool-call' }>;
type ToolResultBlock = Extract<ContentBlock, { type: 'tool-result' }>;
type TextBlock = Extract<ContentBlock, { type: 'text' }>;

function isToolCall(block: ContentBlock): block is ToolCallBlock {
  return block.type === 'tool-call';
}

function isToolResult(block: ContentBlock): block is ToolResultBlock {
  return block.type === 'tool-result';
}

function isTextBlock(block: ContentBlock): block is TextBlock {
  return block.type === 'text';
}

function createGroupedToolBlock(
  toolCall: ToolCallBlock
): Extract<GroupedBlock, { type: 'grouped-tool' }> {
  return {
    type: 'grouped-tool',
    toolCall
  };
}

function attachToolResultToGroupedBlock(
  groupedBlock: Extract<GroupedBlock, { type: 'grouped-tool' }>,
  toolResult: ToolResultBlock
): void {
  groupedBlock.toolResult = toolResult;
}

function groupToolCallsWithResults(blocks: ContentBlock[]): GroupedBlock[] {
  const grouped: GroupedBlock[] = [];
  const toolCallIndexMap = new Map<string, number>();

  for (const block of blocks) {
    if (isToolCall(block)) {
      const index = grouped.length;
      toolCallIndexMap.set(block.toolCallId, index);
      grouped.push(createGroupedToolBlock(block));
    } else if (isToolResult(block)) {
      const groupedIndex = toolCallIndexMap.get(block.toolCallId);

      if (groupedIndex !== undefined) {
        const groupedBlock = grouped[groupedIndex] as Extract<
          GroupedBlock,
          { type: 'grouped-tool' }
        >;
        attachToolResultToGroupedBlock(groupedBlock, block);
      } else {
        grouped.push(block);
      }
    } else {
      grouped.push(block);
    }
  }

  return grouped;
}

export function useGroupedBlocks(blocks: ContentBlock[] | undefined): GroupedBlock[] | null {
  return useMemo(() => {
    if (!blocks || blocks.length === 0) {
      return null;
    }

    return groupToolCallsWithResults(blocks);
  }, [blocks]);
}

function extractTextFromBlocks(blocks: ContentBlock[]): string {
  return blocks
    .filter(isTextBlock)
    .map((block) => block.content)
    .join('\n\n');
}

export function extractMessageText(message: Message): string {
  if (message.content) {
    return message.content;
  }

  if (message.blocks && message.blocks.length > 0) {
    return extractTextFromBlocks(message.blocks);
  }

  return '';
}

const COPY_RESET_DELAY_MS = 2000;

async function copyTextToClipboard(text: string): Promise<void> {
  await navigator.clipboard.writeText(text);
}

function resetCopyState(setCopied: (value: boolean) => void): void {
  setTimeout(() => setCopied(false), COPY_RESET_DELAY_MS);
}

export function useCopyToClipboard(text: string) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    if (!text.trim()) {
      return;
    }

    try {
      await copyTextToClipboard(text);
      setCopied(true);
      resetCopyState(setCopied);
    } catch (error) {
      console.error('Failed to copy text:', error);
    }
  }, [text]);

  return { copied, handleCopy };
}

function hasNoContent(message: Message): boolean {
  const hasNoTextContent = !message.content || message.content.trim() === '';
  const hasNoBlocks = !message.blocks || message.blocks.length === 0;
  return hasNoTextContent && hasNoBlocks;
}

function shouldShowThinkingIndicator(
  isUser: boolean,
  isStreaming: boolean,
  message: Message
): boolean {
  if (isUser) return false;
  if (!isStreaming) return false;
  return hasNoContent(message);
}

export function useMessageBubble(message: Message, isStreaming: boolean = false) {
  const isUser = message.role === 'user';

  const showThinkingIndicator = useMemo(() => {
    return shouldShowThinkingIndicator(isUser, isStreaming, message);
  }, [isUser, isStreaming, message]);

  const groupedBlocks = useGroupedBlocks(message.blocks);

  const messageText = useMemo(() => extractMessageText(message), [message]);

  return {
    isUser,
    showThinkingIndicator,
    groupedBlocks,
    messageText
  };
}
