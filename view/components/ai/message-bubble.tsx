'use client';

import React from 'react';
import { Streamdown } from 'streamdown';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';
import { User, Bot, Copy, Check } from 'lucide-react';
import { formatTime } from './format-time';
import { ToolCall } from './tool-call';
import type { ContentBlock } from './types';
import { Button } from '../ui/button';
import {
  useMessageBubble,
  useCopyToClipboard,
  type GroupedBlock
} from '../../hooks/use-message-bubble';

interface MessageBubbleProps {
  message: import('./types').Message;
  isStreaming?: boolean;
  isLastMessage?: boolean;
}

function ThinkingIndicator() {
  return (
    <div className="flex items-center gap-1.5 py-1 min-h-[20px]">
      <span className="size-2 rounded-full bg-muted-foreground/60 animate-pulse" />
      <span
        className="size-2 rounded-full bg-muted-foreground/60 animate-pulse"
        style={{ animationDelay: '150ms' }}
      />
      <span
        className="size-2 rounded-full bg-muted-foreground/60 animate-pulse"
        style={{ animationDelay: '300ms' }}
      />
    </div>
  );
}

interface MessageAvatarProps {
  isUser: boolean;
}

function MessageAvatar({ isUser }: MessageAvatarProps) {
  return (
    <Avatar className="size-8 shrink-0 shadow-sm">
      <AvatarFallback
        className={cn(
          'text-xs font-medium',
          isUser
            ? 'bg-primary text-primary-foreground'
            : 'bg-gradient-to-br from-muted to-muted/50 text-muted-foreground border border-border/30'
        )}
      >
        {isUser ? <User className="size-4" /> : <Bot className="size-4" />}
      </AvatarFallback>
    </Avatar>
  );
}

interface MessageTextContentProps {
  content: string;
  isStreaming?: boolean;
}

function MessageTextContent({ content, isStreaming = false }: MessageTextContentProps) {
  return (
    <div className="overflow-x-auto -mx-4 px-4">
      <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2 [&_table]:w-full [&_table]:min-w-full">
        <Streamdown isAnimating={isStreaming}>{content}</Streamdown>
      </div>
    </div>
  );
}

interface MessageTimestampProps {
  timestamp: Date;
  isUser: boolean;
}

function MessageTimestamp({ timestamp, isUser }: MessageTimestampProps) {
  return (
    <span
      className={cn('text-xs text-muted-foreground mt-1 px-1', isUser ? 'text-right' : 'text-left')}
    >
      {formatTime(timestamp)}
    </span>
  );
}

interface MessageBlocksProps {
  groupedBlocks: GroupedBlock[];
  isStreaming?: boolean;
}

function MessageBlocks({ groupedBlocks, isStreaming = false }: MessageBlocksProps) {
  return (
    <div className="space-y-1.5">
      {groupedBlocks.map((block, index) => {
        if ('type' in block && block.type === 'grouped-tool') {
          const toolCall = block.toolCall;
          const toolResult = block.toolResult;
          const nextBlock = groupedBlocks[index + 1];
          const hasTextAfter = nextBlock && 'type' in nextBlock && nextBlock.type === 'text';
          return (
            <div
              key={`tool-${toolCall.toolCallId}-${index}`}
              className={hasTextAfter ? 'mb-1.5' : ''}
            >
              <ToolCall
                toolName={toolCall.toolName}
                toolCallId={toolCall.toolCallId}
                arguments={toolCall.arguments}
                result={toolResult?.result}
                isError={toolResult?.isError}
                isComplete={!!toolResult}
              />
            </div>
          );
        }

        if ('type' in block && block.type === 'text') {
          const isLastTextBlock =
            index === groupedBlocks.length - 1 ||
            !groupedBlocks.slice(index + 1).some((b) => 'type' in b && b.type === 'text');
          return (
            <div key={`text-${index}`} className="overflow-x-auto -mx-4 px-4">
              <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2 [&_table]:w-full [&_table]:min-w-full">
                <Streamdown isAnimating={isStreaming && isLastTextBlock}>
                  {block.content}
                </Streamdown>
              </div>
            </div>
          );
        }

        return null;
      })}
    </div>
  );
}

interface FallbackBlocksProps {
  blocks: ContentBlock[];
  isStreaming?: boolean;
}

function FallbackBlocks({ blocks, isStreaming = false }: FallbackBlocksProps) {
  return (
    <div className="space-y-2">
      {blocks.map((block, index) => {
        if (block.type === 'tool-call') {
          const toolResult = blocks.find(
            (b) => b.type === 'tool-result' && b.toolCallId === block.toolCallId
          ) as Extract<ContentBlock, { type: 'tool-result' }> | undefined;
          return (
            <ToolCall
              key={`tool-${block.toolCallId}-${index}`}
              toolName={block.toolName}
              toolCallId={block.toolCallId}
              arguments={block.arguments}
              result={toolResult?.result}
              isError={toolResult?.isError}
              isComplete={!!toolResult}
            />
          );
        }

        if (block.type === 'text') {
          return (
            <div key={`text-${index}`} className="overflow-x-auto -mx-4 px-4">
              <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2 [&_table]:w-full [&_table]:min-w-full">
                <Streamdown isAnimating={isStreaming}>{block.content}</Streamdown>
              </div>
            </div>
          );
        }

        return null;
      })}
    </div>
  );
}

interface MessageContentProps {
  message: import('./types').Message;
  isStreaming?: boolean;
  showThinkingIndicator: boolean;
  groupedBlocks: GroupedBlock[] | null;
}

function MessageContent({
  message,
  isStreaming = false,
  showThinkingIndicator,
  groupedBlocks
}: MessageContentProps) {
  if (message.role === 'user') {
    return <p className="text-sm whitespace-pre-wrap">{message.content}</p>;
  }

  if (showThinkingIndicator) {
    return <ThinkingIndicator />;
  }

  if (groupedBlocks && groupedBlocks.length > 0) {
    return <MessageBlocks groupedBlocks={groupedBlocks} isStreaming={isStreaming} />;
  }

  if (message.blocks && message.blocks.length > 0) {
    return <FallbackBlocks blocks={message.blocks} isStreaming={isStreaming} />;
  }

  return <MessageTextContent content={message.content} isStreaming={isStreaming} />;
}

function CopyButton({ text, isUser }: { text: string; isUser: boolean }) {
  const { copied, handleCopy } = useCopyToClipboard(text);

  if (!text.trim()) {
    return null;
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      className={cn('size-7 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity mt-1')}
      onClick={handleCopy}
      title="Copy message"
    >
      {copied ? (
        <Check className="size-3.5 text-green-600 dark:text-green-400" />
      ) : (
        <Copy className="size-3.5 text-muted-foreground" />
      )}
    </Button>
  );
}

export function MessageBubble({
  message,
  isStreaming = false,
  isLastMessage = false
}: MessageBubbleProps) {
  const { isUser, showThinkingIndicator, groupedBlocks, messageText } = useMessageBubble(
    message,
    isStreaming
  );

  return (
    <div className={cn('flex gap-3 min-w-0 group', isUser ? 'flex-row-reverse' : 'flex-row')}>
      <MessageAvatar isUser={isUser} />
      <div
        className={cn(
          'flex-1 max-w-[85%] min-w-0 flex flex-col',
          isUser ? 'items-end' : 'items-start'
        )}
      >
        <div
          className={cn(
            'rounded-2xl px-4 py-3 min-h-[44px] max-w-full',
            isUser
              ? 'bg-primary/10 text-foreground rounded-tr-md border border-primary/20'
              : 'text-foreground rounded-tl-md'
          )}
        >
          <MessageContent
            message={message}
            isStreaming={isStreaming}
            showThinkingIndicator={showThinkingIndicator}
            groupedBlocks={groupedBlocks}
          />
        </div>
        <div
          className={cn('flex items-center gap-2 mt-1', isUser ? 'flex-row-reverse' : 'flex-row')}
        >
          <MessageTimestamp timestamp={message.timestamp} isUser={isUser} />
          {isLastMessage && <CopyButton text={messageText} isUser={isUser} />}
        </div>
      </div>
    </div>
  );
}
