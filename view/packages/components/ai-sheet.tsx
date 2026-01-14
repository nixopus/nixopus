'use client';

import React from 'react';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ArrowUpRight, Bot, Grid3x3, Sparkles } from 'lucide-react';
import { useState } from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import { Loader2, Send } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { useAIChat } from '@/packages/hooks/ai/use-ai-chat';
import { cn } from '@/lib/utils';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { User } from 'lucide-react';
import { Streamdown } from 'streamdown';
import { type Thread } from '@/redux/services/agents/agentsApi';
import { MessageSquare, Plus, History } from 'lucide-react';
import { useThreads } from '@/packages/hooks/ai/use-threads';
import {
  TypographyH3,
  TypographyP,
  TypographyMuted,
  TypographySmall
} from '@/components/ui/typography';
import {
  AIContentProps,
  AISheetProps,
  Message,
  MessageBubbleProps,
  ToolCallProps,
  ThreadsSidebarProps,
  EmptyStateProps,
  SUGGESTION_KEYS
} from '../types/ai';
import { formatTime, formatResult, getChevronIcon } from '@/packages/utils/ai';

export function AISheet({ open, onOpenChange }: AISheetProps) {
  const { t } = useTranslation();

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="sm:max-w-xl w-full p-0 flex flex-col gap-0 bg-background/95 backdrop-blur-sm h-full max-h-screen overflow-hidden"
      >
        <SheetHeader className="px-6 py-4 border-b border-border/50 shrink-0">
          <SheetTitle className="flex items-center gap-3 text-lg">
            <div className="flex items-center justify-center size-8 rounded-lg bg-primary/10">
              <Sparkles className="size-4 text-primary" />
            </div>
            <span>{t('ai.title')}</span>
          </SheetTitle>
        </SheetHeader>

        <AIContent open={open} className="flex-1 min-h-0 overflow-hidden" />
      </SheetContent>
    </Sheet>
  );
}

export function AIContent({ open, className, threadId, onThreadChange }: AIContentProps) {
  const { t } = useTranslation();
  const {
    messages,
    inputValue,
    isStreaming,
    isLoadingMessages,
    scrollRef,
    textareaRef,
    handleSubmit,
    handleKeyDown,
    handleSuggestionClick,
    handleInputChange
  } = useAIChat({ open, threadId, onThreadChange });

  const renderContent = () => {
    if (isLoadingMessages) {
      return <LoadingSkeletons />;
    }
    if (messages.length === 0) {
      return <EmptyState onSuggestionClick={handleSuggestionClick} />;
    }
    return (
      <div className="space-y-6">
        {messages.map((message, index) => {
          const isLastMessage = index === messages.length - 1;
          const shouldShowStreaming = isStreaming && isLastMessage && message.role === 'assistant';
          return (
            <MessageBubble
              key={message.id}
              message={message}
              isStreaming={shouldShowStreaming}
              isLastMessage={isLastMessage}
            />
          );
        })}
      </div>
    );
  };

  return (
    <div className={`relative flex flex-col h-full overflow-hidden min-w-0 ${className || ''}`}>
      <ScrollArea className="flex-1 min-h-0 min-w-0" ref={scrollRef}>
        <div className="px-6 py-8 pb-6 min-w-0">{renderContent()}</div>
      </ScrollArea>

      <div className="shrink-0  bg-gradient-to-t from-background via-background to-background/80 px-4 py-4">
        <form onSubmit={handleSubmit} className="max-w-3xl mx-auto">
          <div className="flex items-end gap-3 bg-muted/40 border border-border/60 rounded-2xl px-4 py-3 focus-within:ring-2 focus-within:ring-primary/30 focus-within:border-primary/40 focus-within:bg-muted/60 transition-all duration-200 shadow-sm">
            <Textarea
              ref={textareaRef}
              value={inputValue}
              onChange={handleInputChange}
              onKeyDown={handleKeyDown}
              placeholder={t('ai.input.placeholder')}
              className="flex-1 min-h-[40px] max-h-[160px] resize-none bg-transparent border-0 focus-visible:ring-0 focus-visible:ring-offset-0 overflow-y-auto px-1 py-[10px] text-sm leading-relaxed placeholder:text-muted-foreground/60"
              disabled={isStreaming}
              rows={1}
            />
            <Button
              type="submit"
              size="icon"
              disabled={!inputValue.trim() || isStreaming}
              className="shrink-0 size-9 rounded-xl bg-primary hover:bg-primary/90 text-primary-foreground disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-200 flex items-center justify-center mb-0.5 shadow-sm"
              title={t('ai.input.sendMessage')}
            >
              {isStreaming ? (
                <Loader2 className="size-[18px] animate-spin" />
              ) : (
                <Send className="size-[18px]" />
              )}
            </Button>
          </div>
          <TypographySmall className="text-[11px] text-muted-foreground/50 text-center mt-2.5">
            {t('ai.input.hint')}
          </TypographySmall>
        </form>
      </div>
    </div>
  );
}

export function EmptyState({ onSuggestionClick }: EmptyStateProps) {
  const { t } = useTranslation();

  const suggestions = SUGGESTION_KEYS.map((key) => ({
    key,
    text: t(key)
  })).filter((suggestion) => suggestion.text && suggestion.text.trim() !== '');

  const handleSuggestionClick = (text: string) => {
    if (onSuggestionClick && typeof onSuggestionClick === 'function') {
      onSuggestionClick(text);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center py-20 px-4">
      <div className="relative mb-8">
        <div className="absolute inset-0 bg-primary/20 rounded-3xl blur-xl" />
        <div className="relative flex items-center justify-center size-20 rounded-2xl bg-gradient-to-br from-primary/20 to-primary/5 border border-primary/20">
          <Sparkles className="size-9 text-primary" />
        </div>
      </div>
      <TypographyH3 className="text-xl font-semibold text-foreground mb-2">
        {t('ai.emptyState.title')}
      </TypographyH3>
      <TypographyMuted className="text-center max-w-md leading-relaxed">
        {t('ai.emptyState.description')}
      </TypographyMuted>
      {suggestions.length > 0 && (
        <div className="mt-10 w-full max-w-md">
          <TypographySmall className="font-medium text-muted-foreground/70 uppercase tracking-wider mb-3 text-center">
            {t('ai.emptyState.tryAskingAbout')}
          </TypographySmall>
          <div className="grid grid-cols-1 gap-2">
            {suggestions.map((suggestion) => (
              <Button
                key={suggestion.key}
                type="button"
                variant="ghost"
                onClick={() => handleSuggestionClick(suggestion.text)}
                className="group flex items-center justify-between gap-3 px-4 py-3.5 text-sm text-left rounded-xl border border-border/40 bg-muted/20 hover:bg-muted/50 hover:border-primary/30 hover:shadow-sm transition-all duration-200 text-muted-foreground hover:text-foreground h-auto"
              >
                <span>{suggestion.text}</span>
                <ArrowUpRight className="size-4 opacity-0 -translate-x-1 group-hover:opacity-70 group-hover:translate-x-0 transition-all duration-200" />
              </Button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export function ToolCall({ toolName, result, isError, isComplete }: ToolCallProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasResult = Boolean(result !== undefined && isComplete);

  const handleClick = () => {
    if (hasResult) {
      setIsExpanded(!isExpanded);
    }
  };

  return (
    <div className="my-1">
      <Button
        variant="ghost"
        onClick={handleClick}
        disabled={!hasResult}
        className="flex items-center gap-2 cursor-pointer hover:opacity-80 transition-opacity h-auto p-0"
      >
        <span className="text-muted-foreground/60 text-sm">
          {getChevronIcon(hasResult, isExpanded)}
        </span>
        <div className="flex items-center gap-2 px-2 py-1 rounded-md bg-muted/40">
          <Grid3x3 className="size-3.5 text-amber-500 shrink-0" />
          <span className="text-xs font-mono text-foreground">{toolName}</span>
        </div>
      </Button>

      {isExpanded && hasResult && (
        <div className="ml-6 mt-1 px-2 py-1.5 rounded-md bg-muted/20 border border-border/30">
          {isError ? (
            <div className="text-xs text-destructive/80 font-mono">
              Error: {formatResult(result, true)}
            </div>
          ) : (
            <pre className="text-xs text-muted-foreground font-mono overflow-x-auto whitespace-pre-wrap">
              {formatResult(result, false)}
            </pre>
          )}
        </div>
      )}
    </div>
  );
}

export function MessageBubble({
  message,
  isStreaming = false,
  isLastMessage = false
}: MessageBubbleProps) {
  const isUser = message.role === 'user';

  return (
    <div className={cn('flex gap-3', isUser ? 'flex-row-reverse' : 'flex-row')}>
      <Avatar className="size-8 shrink-0">
        <AvatarFallback
          className={cn(
            'text-xs font-medium',
            isUser ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'
          )}
        >
          {isUser ? <User className="size-4" /> : <Bot className="size-4" />}
        </AvatarFallback>
      </Avatar>
      <div className={cn('flex-1 max-w-[85%] flex flex-col', isUser ? 'items-end' : 'items-start')}>
        <div
          className={cn(
            'rounded-2xl px-4 py-3',
            isUser
              ? 'bg-primary text-primary-foreground rounded-tr-md'
              : 'bg-muted/60 text-foreground rounded-tl-md'
          )}
        >
          {isUser ? (
            <TypographyP className="text-sm whitespace-pre-wrap mb-0">
              {message.content}
            </TypographyP>
          ) : (
            <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2">
              <Streamdown isAnimating={isStreaming}>{message.content}</Streamdown>
            </div>
          )}
        </div>
        <TypographySmall
          className={cn(
            'text-xs text-muted-foreground mt-1 px-1',
            isUser ? 'text-right' : 'text-left'
          )}
        >
          {formatTime(message.timestamp)}
        </TypographySmall>
      </div>
    </div>
  );
}

function LoadingSkeletons() {
  return (
    <div className="space-y-1.5 pt-1">
      {[1, 2, 3, 4, 5].map((i) => (
        <Skeleton key={i} className="h-12 w-full rounded-lg" />
      ))}
    </div>
  );
}

function ErrorState() {
  const { t } = useTranslation();
  return (
    <div className="px-2 py-8 text-center">
      <div className="size-8 rounded-full bg-destructive/10 flex items-center justify-center mx-auto mb-2">
        <MessageSquare className="size-4 text-destructive/70" />
      </div>
      <TypographyMuted className="text-xs">{t('ai.threads.error')}</TypographyMuted>
    </div>
  );
}

function ThreadsEmptyState() {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
      <div className="size-12 rounded-full bg-muted/40 flex items-center justify-center mb-3">
        <MessageSquare className="size-5 text-muted-foreground/60" />
      </div>
      <TypographySmall className="text-xs font-medium text-foreground/80 mb-1">
        {t('ai.threads.emptyState.title')}
      </TypographySmall>
      <TypographySmall className="text-[10px] text-muted-foreground/70">
        {t('ai.threads.emptyState.description')}
      </TypographySmall>
    </div>
  );
}

export function ThreadsSidebar({
  selectedThreadId,
  onThreadSelect,
  onNewThread,
  className
}: ThreadsSidebarProps) {
  const { t } = useTranslation();
  const { threads, isLoading, error, formatDate } = useThreads();

  const renderContent = () => {
    if (isLoading) {
      return <LoadingSkeletons />;
    }
    if (error) {
      return <ErrorState />;
    }
    if (threads.length === 0) {
      return <ThreadsEmptyState />;
    }
    return (
      <div className="space-y-0.5 pt-1">
        {threads.map((thread) => {
          const isSelected = selectedThreadId === thread.id;
          return (
            <Button
              key={thread.id}
              variant="ghost"
              onClick={() => onThreadSelect(thread.id)}
              className={cn(
                'w-full text-left px-2.5 py-2 rounded-lg transition-all duration-150 group h-auto justify-start',
                isSelected
                  ? 'bg-primary/10 border border-primary/20 shadow-sm'
                  : 'border border-transparent'
              )}
            >
              <div className="flex items-start gap-2 w-full">
                <div className="flex-1 min-w-0">
                  <TypographySmall
                    className={cn(
                      'text-xs truncate transition-colors block',
                      isSelected
                        ? 'font-medium text-foreground'
                        : 'text-foreground/80 group-hover:text-foreground'
                    )}
                  >
                    {thread.title || t('ai.threads.untitledChat')}
                  </TypographySmall>
                  <TypographySmall className="text-[10px] text-muted-foreground/70 mt-0.5 block">
                    {formatDate(thread.updatedAt)}
                  </TypographySmall>
                </div>
              </div>
            </Button>
          );
        })}
      </div>
    );
  };

  return (
    <div
      className={cn(
        'flex flex-col h-full border-r border-border/30 bg-gradient-to-b from-muted/20 via-muted/10 to-transparent overflow-hidden w-64',
        className
      )}
    >
      <div className="shrink-0 p-3 space-y-2">
        <Button
          onClick={onNewThread}
          className="w-full justify-center gap-2 h-9 rounded-xl shadow-sm text-sm"
          variant="outline"
        >
          <Plus className="size-3.5" />
          <span className="font-medium">{t('ai.threads.newChat')}</span>
        </Button>
      </div>

      <div className="flex-1 flex flex-col min-h-0 overflow-hidden relative">
        <div className="shrink-0 px-3 py-1.5">
          <div className="flex items-center gap-1.5 text-[10px] font-medium text-muted-foreground uppercase tracking-wider px-1">
            <History className="size-2.5" />
            <span>{t('ai.threads.recentChats')}</span>
          </div>
        </div>

        <div className="flex-1 min-h-0 overflow-hidden">
          <ScrollArea className="h-full">
            <div className="px-2 pb-3">{renderContent()}</div>
          </ScrollArea>
        </div>
      </div>
    </div>
  );
}
