'use client';

import React from 'react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { Send, Loader2 } from 'lucide-react';
import { MessageBubble } from './message-bubble';
import { useAIChat } from '../../hooks/use-ai-chat';
import type { Message } from './types';
import { Sparkles } from 'lucide-react';
import { ArrowUpRight } from 'lucide-react';

interface AIContentProps {
  open: boolean;
  className?: string;
  threadId?: string | null;
  onThreadChange?: (threadId: string) => void;
}

function LoadingSkeletons() {
  return (
    <div className="space-y-6">
      {[1, 2, 3].map((i) => {
        const isEven = i % 2 === 0;
        return (
          <div key={i} className={`flex ${isEven ? 'justify-end' : 'justify-start'}`}>
            <Skeleton className={`h-20 ${isEven ? 'w-2/3' : 'w-3/4'} rounded-2xl`} />
          </div>
        );
      })}
    </div>
  );
}

function MessagesList({ messages, isStreaming }: { messages: Message[]; isStreaming: boolean }) {
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
}

function SendButtonIcon({ isStreaming }: { isStreaming: boolean }) {
  if (isStreaming) {
    return <Loader2 className="size-[18px] animate-spin" />;
  }
  return <Send className="size-[18px]" />;
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
    return <MessagesList messages={messages} isStreaming={isStreaming} />;
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
              <SendButtonIcon isStreaming={isStreaming} />
            </Button>
          </div>
          <p className="text-[11px] text-muted-foreground/50 text-center mt-2.5">
            {t('ai.input.hint')}
          </p>
        </form>
      </div>
    </div>
  );
}

interface EmptyStateProps {
  onSuggestionClick: (text: string) => void;
}

const SUGGESTION_KEYS = [
  'ai.suggestions.deploy',
  'ai.suggestions.logs',
  'ai.suggestions.envVars'
] as const;

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
      <h3 className="text-xl font-semibold text-foreground mb-2">{t('ai.emptyState.title')}</h3>
      <p className="text-sm text-muted-foreground text-center max-w-md leading-relaxed">
        {t('ai.emptyState.description')}
      </p>
      {suggestions.length > 0 && (
        <div className="mt-10 w-full max-w-md">
          <p className="text-xs font-medium text-muted-foreground/70 uppercase tracking-wider mb-3 text-center">
            {t('ai.emptyState.tryAskingAbout')}
          </p>
          <div className="grid grid-cols-1 gap-2">
            {suggestions.map((suggestion) => (
              <SuggestionChip
                key={suggestion.key}
                text={suggestion.text}
                onClick={() => handleSuggestionClick(suggestion.text)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

interface SuggestionChipProps {
  text: string;
  onClick: () => void;
}

function SuggestionChip({ text, onClick }: SuggestionChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="group flex items-center justify-between gap-3 px-4 py-3.5 text-sm text-left rounded-xl border border-border/40 bg-muted/20 hover:bg-muted/50 hover:border-primary/30 hover:shadow-sm transition-all duration-200 text-muted-foreground hover:text-foreground"
    >
      <span>{text}</span>
      <ArrowUpRight className="size-4 opacity-0 -translate-x-1 group-hover:opacity-70 group-hover:translate-x-0 transition-all duration-200" />
    </button>
  );
}
