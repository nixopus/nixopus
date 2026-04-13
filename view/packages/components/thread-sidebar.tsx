'use client';

import React from 'react';
import {
  Button,
  ScrollArea,
  ScrollBar,
  Separator,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  TooltipProvider,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Input
} from '@nixopus/ui';
import {
  Plus,
  Trash2,
  MessageSquare,
  MessageSquareText,
  PanelLeftClose,
  PanelLeftOpen,
  Search,
  Pencil,
  AlertTriangle,
  RotateCw,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Clock,
  Type
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { type ChatThread } from '@/packages/hooks/ai/use-chat-threads';
import { useThreadSidebarSearch } from '@/packages/hooks/ai/use-chat-page';

export interface ThreadSidebarProps {
  threads: ChatThread[];
  activeThreadId: string | null;
  resourceId?: string;
  isLoading?: boolean;
  isCollapsed?: boolean;
  onToggleCollapse: () => void;
  onSelectThread: (id: string) => void;
  onNewChat: () => void;
  onDeleteThread: (id: string) => void;
  onRenameThread: (id: string, title: string) => void;
  onRefresh: () => void;
  isRefreshing?: boolean;
  hideCollapseToggle?: boolean;
}

type ThreadFilter = 'all' | 'chats' | 'incidents';
type ThreadSortKey = 'updatedAt' | 'createdAt' | 'title';
type ThreadSortDir = 'desc' | 'asc';

function useThreadFilterSort(threads: ChatThread[]) {
  const [filter, setFilter] = React.useState<ThreadFilter>('all');
  const [sortKey, setSortKey] = React.useState<ThreadSortKey>('updatedAt');
  const [sortDir, setSortDir] = React.useState<ThreadSortDir>('desc');

  const filtered = React.useMemo(() => {
    let result = threads;
    if (filter === 'incidents') {
      result = result.filter((t) => t.isIncident);
    } else if (filter === 'chats') {
      result = result.filter((t) => !t.isIncident);
    }
    return [...result].sort((a, b) => {
      if (sortKey === 'title') {
        const cmp = a.title.localeCompare(b.title);
        return sortDir === 'asc' ? cmp : -cmp;
      }
      const aTime = a[sortKey].getTime();
      const bTime = b[sortKey].getTime();
      return sortDir === 'desc' ? bTime - aTime : aTime - bTime;
    });
  }, [threads, filter, sortKey, sortDir]);

  const toggleSort = (key: ThreadSortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'desc' ? 'asc' : 'desc'));
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  };

  return { filter, setFilter, sortKey, sortDir, toggleSort, filtered };
}

function Skeleton({ className }: { className?: string }) {
  return <div className={cn('animate-pulse rounded-md bg-muted', className)} />;
}

function ThreadsSkeleton() {
  return (
    <div className="space-y-1 px-1">
      {[...Array(5)].map((_, i) => (
        <div key={i} className="flex items-center gap-2 px-3 py-2">
          <Skeleton className="size-4 shrink-0 rounded" />
          <Skeleton className="h-4 flex-1" />
        </div>
      ))}
    </div>
  );
}

interface ThreadItemProps {
  thread: ChatThread;
  isActive: boolean;
  onSelect: () => void;
  onDelete: () => void;
  onRename: (title: string) => void;
}

function ThreadItem({ thread, isActive, onSelect, onDelete, onRename }: ThreadItemProps) {
  const [isEditing, setIsEditing] = React.useState(false);
  const [editValue, setEditValue] = React.useState(thread.title);
  const inputRef = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    if (isEditing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [isEditing]);

  const handleStartEditing = (e: React.MouseEvent) => {
    e.stopPropagation();
    setEditValue(thread.title);
    setIsEditing(true);
  };

  const handleSave = () => {
    const trimmed = editValue.trim();
    if (trimmed && trimmed !== thread.title) {
      onRename(trimmed);
    }
    setIsEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSave();
    }
    if (e.key === 'Escape') {
      setEditValue(thread.title);
      setIsEditing(false);
    }
  };

  if (isEditing && !thread.isIncident) {
    return (
      <div
        className={cn(
          'relative w-full min-w-0 flex items-center gap-2 px-3 py-1.5 rounded-md text-sm',
          isActive ? 'bg-primary/10' : 'bg-muted/60'
        )}
      >
        <MessageSquare className="size-4 shrink-0 text-muted-foreground" />
        <Input
          ref={inputRef}
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          onBlur={handleSave}
          onKeyDown={handleKeyDown}
          className="h-6 px-1 text-sm border-none bg-transparent focus-visible:ring-1 focus-visible:ring-primary/40"
        />
      </div>
    );
  }

  return (
    <button
      onClick={onSelect}
      onDoubleClick={thread.isIncident ? undefined : handleStartEditing}
      className={cn(
        'relative w-full min-w-0 flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors group text-left',
        isActive
          ? 'bg-primary/10 text-primary font-medium'
          : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
      )}
    >
      {thread.isIncident ? (
        <AlertTriangle className="size-4 shrink-0 text-amber-500" />
      ) : (
        <MessageSquare className="size-4 shrink-0" />
      )}
      <span className="flex-1 min-w-0 truncate text-left">{thread.title}</span>
      {thread.isIncident && (
        <span className="shrink-0 px-1.5 py-0.5 rounded text-[10px] font-medium leading-none bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/20">
          Incident
        </span>
      )}
      {!thread.isIncident && (
        <TooltipProvider delayDuration={0}>
          <div className="absolute right-1 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5 bg-muted/80 backdrop-blur-sm rounded">
            <Tooltip>
              <TooltipTrigger asChild>
                <span
                  role="button"
                  onClick={handleStartEditing}
                  className="p-1 rounded hover:bg-muted hover:text-foreground"
                >
                  <Pencil className="size-3.5" />
                </span>
              </TooltipTrigger>
              <TooltipContent side="right">Rename</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <span
                  role="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    onDelete();
                  }}
                  className="p-1 rounded hover:bg-destructive/10 hover:text-destructive"
                >
                  <Trash2 className="size-3.5" />
                </span>
              </TooltipTrigger>
              <TooltipContent side="right">Delete</TooltipContent>
            </Tooltip>
          </div>
        </TooltipProvider>
      )}
    </button>
  );
}

export function ThreadSidebar({
  threads,
  activeThreadId,
  resourceId,
  isLoading,
  isCollapsed,
  onToggleCollapse,
  onSelectThread,
  onNewChat,
  onDeleteThread,
  onRenameThread,
  onRefresh,
  isRefreshing,
  hideCollapseToggle
}: ThreadSidebarProps) {
  const { t } = useTranslation();
  const sidebarSearch = useThreadSidebarSearch(resourceId);
  const fs = useThreadFilterSort(threads);
  const hasIncidents = threads.some((t) => t.isIncident);

  if (isCollapsed) {
    return (
      <div className="w-12 shrink-0 border-r border-border/50 flex flex-col items-center bg-muted/20 py-2 gap-1">
        <TooltipProvider delayDuration={0}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon" className="size-8" onClick={onToggleCollapse}>
                <PanelLeftOpen className="size-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">{t('ai.threads.expandSidebar')}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon" className="size-8" onClick={onNewChat}>
                <Plus className="size-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">{t('ai.threads.newChat')}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="size-8"
                onClick={onRefresh}
                disabled={isRefreshing}
              >
                <RotateCw className={cn('size-4', isRefreshing && 'animate-spin')} />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">Refresh</TooltipContent>
          </Tooltip>
        </TooltipProvider>
        {hasIncidents && (
          <>
            <Separator className="my-1" />
            <TooltipProvider delayDuration={0}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant={fs.filter === 'incidents' ? 'secondary' : 'ghost'}
                    size="icon"
                    className="size-8"
                    onClick={() => {
                      const next: Record<ThreadFilter, ThreadFilter> = {
                        all: 'chats',
                        chats: 'incidents',
                        incidents: 'all'
                      };
                      fs.setFilter(next[fs.filter]);
                    }}
                  >
                    {fs.filter === 'incidents' ? (
                      <AlertTriangle className="size-4 text-amber-500" />
                    ) : fs.filter === 'chats' ? (
                      <MessageSquareText className="size-4" />
                    ) : (
                      <ArrowUpDown className="size-4" />
                    )}
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="right">
                  {fs.filter === 'all'
                    ? 'All threads'
                    : fs.filter === 'chats'
                      ? 'Chats only'
                      : 'Incidents only'}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </>
        )}
        <Separator className="my-1" />
        <ScrollArea className="flex-1 w-full">
          <div className="flex flex-col items-center gap-0.5 px-1">
            <TooltipProvider delayDuration={0}>
              {isLoading
                ? [...Array(3)].map((_, i) => <Skeleton key={i} className="size-8 rounded-md" />)
                : fs.filtered.map((thread) => (
                    <Tooltip key={thread.id}>
                      <TooltipTrigger asChild>
                        <Button
                          variant={thread.id === activeThreadId ? 'secondary' : 'ghost'}
                          size="icon"
                          className="size-8"
                          onClick={() => onSelectThread(thread.id)}
                        >
                          {thread.isIncident ? (
                            <AlertTriangle className="size-4 text-amber-500" />
                          ) : (
                            <MessageSquareText className="size-4" />
                          )}
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent side="right">
                        {thread.isIncident
                          ? `[Incident] ${thread.title}`
                          : thread.title || t('ai.threads.untitledChat')}
                      </TooltipContent>
                    </Tooltip>
                  ))}
            </TooltipProvider>
          </div>
          <ScrollBar />
        </ScrollArea>
      </div>
    );
  }

  return (
    <div
      className={cn(
        'shrink-0 flex flex-col',
        isCollapsed ? 'w-12 border-r border-border/50 bg-muted/20' : 'w-full'
      )}
    >
      <div className="p-3 flex items-center gap-2">
        <Button onClick={onNewChat} variant="outline" className="flex-1 gap-2 justify-start">
          <Plus className="size-4" />
          {t('ai.threads.newChat')}
        </Button>
        <TooltipProvider delayDuration={0}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="size-9 shrink-0"
                onClick={onRefresh}
                disabled={isRefreshing}
              >
                <RotateCw className={cn('size-4', isRefreshing && 'animate-spin')} />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">Refresh</TooltipContent>
          </Tooltip>
          {!hideCollapseToggle && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-9 shrink-0"
                  onClick={onToggleCollapse}
                >
                  <PanelLeftClose className="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="right">{t('ai.threads.collapseSidebar')}</TooltipContent>
            </Tooltip>
          )}
        </TooltipProvider>
      </div>
      <div className="px-2 py-2">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
          <Input
            value={sidebarSearch.searchInputValue}
            onChange={(e) => sidebarSearch.handleSearchInputChange(e.target.value)}
            onKeyDown={(e) => sidebarSearch.handleSearchKeyDown(e.key)}
            placeholder={t('ai.threads.searchChats' as Parameters<typeof t>[0])}
            className="h-8 pl-7 text-xs"
          />
        </div>
      </div>
      <Separator />
      {sidebarSearch.memorySearchResults.length > 0 ? (
        <div className="px-3 py-2">
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
            {t('ai.threads.searchResults' as Parameters<typeof t>[0])}
          </span>
        </div>
      ) : (
        <div className="px-2 py-1.5 flex items-center gap-1">
          <div className="flex-1 flex items-center gap-0.5 rounded-md bg-muted/50 p-0.5">
            <button
              type="button"
              onClick={() => fs.setFilter('all')}
              className={cn(
                'flex-1 text-xs py-1 rounded-sm transition-colors',
                fs.filter === 'all'
                  ? 'bg-background shadow-sm font-medium'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              All
            </button>
            {hasIncidents && (
              <>
                <button
                  type="button"
                  onClick={() => fs.setFilter('chats')}
                  className={cn(
                    'flex-1 text-xs py-1 rounded-sm transition-colors',
                    fs.filter === 'chats'
                      ? 'bg-background shadow-sm font-medium'
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  Chats
                </button>
                <button
                  type="button"
                  onClick={() => fs.setFilter('incidents')}
                  className={cn(
                    'flex-1 text-xs py-1 rounded-sm transition-colors',
                    fs.filter === 'incidents'
                      ? 'bg-background shadow-sm font-medium'
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  Incidents
                </button>
              </>
            )}
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="size-7 shrink-0">
                <ArrowUpDown className="size-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-44">
              {[
                { key: 'updatedAt' as const, label: 'Last updated', icon: Clock },
                { key: 'createdAt' as const, label: 'Date created', icon: Clock },
                { key: 'title' as const, label: 'Title', icon: Type }
              ].map(({ key, label, icon: Icon }) => (
                <DropdownMenuItem key={key} onClick={() => fs.toggleSort(key)}>
                  <Icon className="size-3.5 mr-2" />
                  {label}
                  {fs.sortKey === key &&
                    (fs.sortDir === 'desc' ? (
                      <ArrowDown className="size-3 ml-auto" />
                    ) : (
                      <ArrowUp className="size-3 ml-auto" />
                    ))}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}
      <ScrollArea className="flex-1 [&_[data-radix-scroll-area-viewport]>div]:!block [&_[data-radix-scroll-area-viewport]>div]:!min-w-0">
        <div className="px-2 pb-2 space-y-0.5">
          {sidebarSearch.memorySearchResults.length > 0 ? (
            sidebarSearch.isSearching ? (
              <ThreadsSkeleton />
            ) : (
              sidebarSearch.memorySearchResults.map((r) => (
                <button
                  key={r.id}
                  type="button"
                  onClick={() => sidebarSearch.handleSelectSearchResult(r.threadId, onSelectThread)}
                  className="w-full flex flex-col gap-0.5 px-3 py-2 rounded-md text-left text-sm hover:bg-muted/60 transition-colors"
                >
                  <span className="text-xs text-muted-foreground truncate">
                    {r.threadTitle || t('ai.threads.untitledChat')}
                  </span>
                  <span className="text-xs truncate">{r.content}</span>
                </button>
              ))
            )
          ) : isLoading ? (
            <ThreadsSkeleton />
          ) : fs.filtered.length === 0 ? (
            <div className="px-3 py-8 text-center">
              <MessageSquare className="size-8 text-muted-foreground/40 mx-auto mb-2" />
              {fs.filter !== 'all' ? (
                <p className="text-xs text-muted-foreground">
                  {fs.filter === 'incidents' ? 'No incidents' : 'No chats'}
                </p>
              ) : (
                <>
                  <p className="text-xs text-muted-foreground">
                    {t('ai.threads.emptyState.title')}
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-1">
                    {t('ai.threads.emptyState.description')}
                  </p>
                </>
              )}
            </div>
          ) : (
            fs.filtered.map((thread) => (
              <ThreadItem
                key={thread.id}
                thread={thread}
                isActive={thread.id === activeThreadId}
                onSelect={() => onSelectThread(thread.id)}
                onDelete={() => onDeleteThread(thread.id)}
                onRename={(title) => onRenameThread(thread.id, title)}
              />
            ))
          )}
        </div>
        <ScrollBar />
      </ScrollArea>
    </div>
  );
}
