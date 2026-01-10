'use client';

import React from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { type Thread } from '@/redux/services/agents/agentsApi';
import { MessageSquare, Plus, History } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useThreads } from '../../hooks/use-threads';
import { useTranslation } from '@/hooks/use-translation';

interface ThreadsSidebarProps {
  selectedThreadId: string | null;
  onThreadSelect: (threadId: string | null) => void;
  onNewThread: () => void;
  className?: string;
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
      <p className="text-xs text-muted-foreground">{t('ai.threads.error')}</p>
    </div>
  );
}

function EmptyState() {
  const { t } = useTranslation();
  return (
    <div className="px-2 py-8 text-center">
      <div className="size-8 rounded-full bg-muted/60 flex items-center justify-center mx-auto mb-2">
        <MessageSquare className="size-4 text-muted-foreground/70" />
      </div>
      <p className="text-xs text-muted-foreground">{t('ai.threads.emptyState.title')}</p>
      <p className="text-[10px] text-muted-foreground/60 mt-1">
        {t('ai.threads.emptyState.description')}
      </p>
    </div>
  );
}

interface ThreadItemProps {
  thread: Thread;
  isSelected: boolean;
  onSelect: (threadId: string) => void;
  formatDate: (dateString: string) => string;
}

function ThreadItem({ thread, isSelected, onSelect, formatDate }: ThreadItemProps) {
  const { t } = useTranslation();
  return (
    <button
      onClick={() => onSelect(thread.id)}
      className={cn(
        'w-full text-left px-2.5 py-2 rounded-lg transition-all duration-150 group',
        'hover:bg-muted/60 focus:bg-muted/60 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/30',
        isSelected
          ? 'bg-primary/10 border border-primary/20 shadow-sm'
          : 'border border-transparent'
      )}
    >
      <div className="flex items-start gap-2">
        <div className="flex-1 min-w-0">
          <div
            className={cn(
              'text-xs truncate transition-colors',
              isSelected
                ? 'font-medium text-foreground'
                : 'text-foreground/80 group-hover:text-foreground'
            )}
          >
            {thread.title || t('ai.threads.untitledChat')}
          </div>
          <div className="text-[10px] text-muted-foreground/70 mt-0.5">
            {formatDate(thread.updatedAt)}
          </div>
        </div>
      </div>
    </button>
  );
}

function ThreadsList({
  threads,
  selectedThreadId,
  onThreadSelect,
  formatDate
}: {
  threads: Thread[];
  selectedThreadId: string | null;
  onThreadSelect: (threadId: string | null) => void;
  formatDate: (dateString: string) => string;
}) {
  return (
    <div className="space-y-0.5 pt-1">
      {threads.map((thread) => (
        <ThreadItem
          key={thread.id}
          thread={thread}
          isSelected={selectedThreadId === thread.id}
          onSelect={onThreadSelect}
          formatDate={formatDate}
        />
      ))}
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
      return <EmptyState />;
    }
    return (
      <ThreadsList
        threads={threads}
        selectedThreadId={selectedThreadId}
        onThreadSelect={onThreadSelect}
        formatDate={formatDate}
      />
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
