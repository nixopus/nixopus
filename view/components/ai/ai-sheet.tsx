'use client';

import React from 'react';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useTranslation } from '@/hooks/use-translation';
import { Sparkles, Send, Loader2 } from 'lucide-react';
import { EmptyState } from './empty-state';
import { MessageBubble } from './message-bubble';
import { StreamingIndicator } from './streaming-indicator';
import { useAIChat } from './use-ai-chat';

interface AISheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AISheet({ open, onOpenChange }: AISheetProps) {
  const { t } = useTranslation();
  const {
    messages,
    inputValue,
    isStreaming,
    scrollRef,
    textareaRef,
    handleSubmit,
    handleKeyDown,
    handleSuggestionClick,
    handleInputChange
  } = useAIChat({ open });

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="sm:max-w-xl w-full p-0 flex flex-col gap-0 bg-background/95 backdrop-blur-sm"
      >
        <SheetHeader className="px-6 py-4 border-b border-border/50 shrink-0">
          <SheetTitle className="flex items-center gap-3 text-lg">
            <div className="flex items-center justify-center size-8 rounded-lg bg-primary/10">
              <Sparkles className="size-4 text-primary" />
            </div>
            <span>{t('ai.title')}</span>
          </SheetTitle>
        </SheetHeader>

        <ScrollArea className="flex-1 min-h-0" ref={scrollRef}>
          <div className="px-4 py-6">
            {messages.length === 0 ? (
              <EmptyState onSuggestionClick={handleSuggestionClick} />
            ) : (
              <div className="space-y-6">
                {messages.map((message) => (
                  <MessageBubble key={message.id} message={message} />
                ))}
                {isStreaming && <StreamingIndicator />}
              </div>
            )}
          </div>
        </ScrollArea>

        <div className="shrink-0 border-t border-border/50 bg-background/80 backdrop-blur-sm p-4">
          <form onSubmit={handleSubmit} className="flex gap-3 items-end">
            <Textarea
              ref={textareaRef}
              value={inputValue}
              onChange={handleInputChange}
              onKeyDown={handleKeyDown}
              placeholder={t('ai.input.placeholder')}
              className="min-h-[44px] max-h-[120px] resize-none bg-muted/50 border-border/50 focus-visible:ring-primary/30"
              disabled={isStreaming}
              rows={1}
            />
            <Button
              type="submit"
              size="icon"
              disabled={!inputValue.trim() || isStreaming}
              className="shrink-0 size-11 rounded-lg"
            >
              {isStreaming ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Send className="size-4" />
              )}
            </Button>
          </form>
          <p className="text-xs text-muted-foreground mt-3 text-center">{t('ai.input.hint')}</p>
        </div>
      </SheetContent>
    </Sheet>
  );
}
