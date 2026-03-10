'use client';

import React, { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import { useTheme } from 'next-themes';
import { useSearchParams, useRouter } from 'next/navigation';
import { Streamdown } from 'streamdown';
import {
  Button,
  ScrollArea,
  ScrollBar,
  Textarea,
  Avatar,
  AvatarFallback,
  Separator,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  TooltipProvider,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSub,
  DropdownMenuSubTrigger,
  DropdownMenuSubContent,
  Input
} from '@nixopus/ui';
import {
  Sparkles,
  Send,
  Loader2,
  Plus,
  Trash2,
  User,
  MessageSquare,
  MessageSquareText,
  StopCircle,
  PanelLeftClose,
  PanelLeftOpen,
  X,
  CirclePlus,
  Search,
  Check
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useAgentChat,
  type ChatMessage,
  type PendingToolApproval
} from '@/packages/hooks/ai/use-agent-chat';
import { useChatThreads, type ChatThread } from '@/packages/hooks/ai/use-chat-threads';
import {
  type ChatContext,
  type ContextProviderData,
  useChatContextProviders,
  stripContextFromMessageText
} from '@/packages/hooks/ai/chat-context';
import { useMemorySearch } from '@/packages/hooks/ai/use-memory-search';

function NixopusIcon({ className }: { className?: string }) {
  const { resolvedTheme } = useTheme();
  const src = resolvedTheme === 'dark' ? '/logo_white.png' : '/logo_black.png';
  return <Image src={src} alt="Nixopus" width={16} height={16} className={className} />;
}

const CHAT_STREAMDOWN_COMPONENTS = {
  a: ({ href, children, ...props }: React.ComponentPropsWithoutRef<'a'>) => (
    <a
      {...props}
      href={href}
      target="_blank"
      rel="noreferrer noopener"
      className={cn('text-primary underline underline-offset-2 hover:opacity-90', props.className)}
    >
      {children}
    </a>
  ),
  pre: ({ children, ...props }: React.ComponentPropsWithoutRef<'pre'>) => (
    <pre
      {...props}
      className={cn(
        'my-2 overflow-x-auto rounded-md border border-border/60 bg-background/70 p-3 text-xs',
        props.className
      )}
    >
      {children}
    </pre>
  ),
  code: ({ children, className, ...props }: React.ComponentPropsWithoutRef<'code'>) => {
    const isBlock = Boolean(className?.includes('language-'));
    if (isBlock) {
      return (
        <code {...props} className={className}>
          {children}
        </code>
      );
    }
    return (
      <code
        {...props}
        className={cn(
          'rounded bg-background/70 px-1 py-0.5 font-mono text-[0.85em] text-foreground',
          className
        )}
      >
        {children}
      </code>
    );
  },
  table: ({ children, ...props }: React.ComponentPropsWithoutRef<'table'>) => (
    <div className="my-2 overflow-x-auto">
      <table {...props} className={cn('w-full text-sm border-collapse', props.className)}>
        {children}
      </table>
    </div>
  ),
  th: ({ children, ...props }: React.ComponentPropsWithoutRef<'th'>) => (
    <th
      {...props}
      className={cn(
        'border border-border/60 bg-muted/40 px-2 py-1 text-left text-xs font-medium',
        props.className
      )}
    >
      {children}
    </th>
  ),
  td: ({ children, ...props }: React.ComponentPropsWithoutRef<'td'>) => (
    <td {...props} className={cn('border border-border/50 px-2 py-1 align-top', props.className)}>
      {children}
    </td>
  ),
  blockquote: ({ children, ...props }: React.ComponentPropsWithoutRef<'blockquote'>) => (
    <blockquote
      {...props}
      className={cn(
        'my-2 border-l-2 border-border pl-3 text-muted-foreground italic',
        props.className
      )}
    >
      {children}
    </blockquote>
  )
};

export function ChatPage() {
  const { t } = useTranslation();
  const searchParams = useSearchParams();
  const navRouter = useRouter();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('chat_sidebar_collapsed') === 'true';
    }
    return false;
  });
  const [selectedContexts, setSelectedContexts] = useState<ChatContext[]>([]);
  const [autoRunTools, setAutoRunTools] = useState(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('chat_auto_run_tools') === 'true';
    }
    return false;
  });
  const [pendingDeployPrompt, setPendingDeployPrompt] = useState<string | null>(null);
  const repoParamsHandledRef = useRef(false);

  useEffect(() => {
    localStorage.setItem('chat_sidebar_collapsed', String(sidebarCollapsed));
  }, [sidebarCollapsed]);

  useEffect(() => {
    localStorage.setItem('chat_auto_run_tools', String(autoRunTools));
  }, [autoRunTools]);

  const contextProviders = useChatContextProviders();

  const threads = useChatThreads();
  const chat = useAgentChat({
    threadId: threads.activeThreadId,
    resourceId: threads.resourceId,
    contexts: selectedContexts,
    autoRunTools,
    waitForThread: threads.waitForThread,
    onFirstMessage: (content) => {
      if (threads.activeThreadId) {
        const title = content.length > 50 ? content.slice(0, 50) + '…' : content;
        threads.updateThreadTitle(threads.activeThreadId, title);
      }
    }
  });

  useEffect(() => {
    if (repoParamsHandledRef.current || !threads.isInitialized) return;

    const repoId = searchParams.get('repo_id');
    const repoName = searchParams.get('repo_name');
    const repoFullName = searchParams.get('repo_full_name');

    if (!repoId || !repoName || !repoFullName) return;

    repoParamsHandledRef.current = true;

    const defaultBranch = searchParams.get('repo_default_branch') || 'main';
    const visibility = searchParams.get('repo_visibility') || 'public';
    const cloneUrl = searchParams.get('repo_clone_url') || '';
    const language = searchParams.get('repo_language') || '';
    const description = searchParams.get('repo_description') || '';
    const htmlUrl = searchParams.get('repo_html_url') || '';

    threads.createThread(repoName);

    const meta: Record<string, string> = {
      'GitHub Repo ID': repoId,
      'Default Branch': defaultBranch,
      Visibility: visibility
    };
    if (cloneUrl) meta['Clone URL'] = cloneUrl;
    if (language) meta['Language'] = language;
    if (htmlUrl) meta['GitHub URL'] = htmlUrl;

    setSelectedContexts([
      {
        type: 'Repository',
        id: repoId,
        label: repoFullName,
        meta
      }
    ]);

    const promptLines = [
      `I want to deploy the GitHub repository "${repoFullName}" as a new application.`,
      '',
      `- GitHub Repository ID (numeric): ${repoId} — use this as the "repository" field when calling createProject`,
      `- Repository name: ${repoFullName}`,
      `- Default branch: ${defaultBranch}`,
      `- Visibility: ${visibility}`
    ];
    if (language) promptLines.push(`- Primary language: ${language}`);
    if (description) promptLines.push(`- Description: ${description}`);
    if (cloneUrl) promptLines.push(`- Clone URL: ${cloneUrl}`);
    promptLines.push(
      '',
      'No application exists yet — please use createProject with the GitHub repository ID above to create and deploy it.'
    );

    setPendingDeployPrompt(promptLines.join('\n'));

    navRouter.replace('/chats');
  }, [threads.isInitialized, searchParams]);

  useEffect(() => {
    if (pendingDeployPrompt && threads.activeThreadId) {
      chat.setInputValue(pendingDeployPrompt);
      setPendingDeployPrompt(null);
      setTimeout(() => chat.textareaRef.current?.focus(), 100);
    }
  }, [pendingDeployPrompt, threads.activeThreadId]);

  const handleNewChat = () => {
    threads.createThread(t('ai.threads.untitledChat'));
  };

  if (!chat.isAgentConfigured) {
    return <AgentDisabledState />;
  }

  return (
    <div className="flex h-full w-full overflow-hidden">
      <ThreadSidebar
        threads={threads.threads}
        activeThreadId={threads.activeThreadId}
        resourceId={threads.resourceId}
        isLoading={!threads.isInitialized}
        isCollapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed((prev) => !prev)}
        onSelectThread={threads.setActiveThreadId}
        onNewChat={handleNewChat}
        onDeleteThread={threads.deleteThread}
      />
      <div className="flex flex-1 flex-col min-w-0">
        {threads.activeThreadId ? (
          <>
            {chat.isLoadingHistory ? (
              <MessagesSkeleton />
            ) : (
              <ChatMessages
                messages={chat.messages}
                isStreaming={chat.isStreaming}
                scrollRef={chat.scrollRef}
                onSuggestionClick={chat.handleSuggestionClick}
                pendingToolApproval={chat.pendingToolApproval}
                autoRunTools={autoRunTools}
                onApproveToolCall={chat.handleApproveToolCall}
                onDeclineToolCall={chat.handleDeclineToolCall}
              />
            )}
            <ChatInput
              inputValue={chat.inputValue}
              isStreaming={chat.isStreaming}
              textareaRef={chat.textareaRef}
              selectedContexts={selectedContexts}
              contextProviders={contextProviders}
              autoRunTools={autoRunTools}
              onAutoRunToolsChange={setAutoRunTools}
              onAddContext={(ctx) =>
                setSelectedContexts((prev) => {
                  if (prev.some((c) => c.type === ctx.type && c.id === ctx.id)) return prev;
                  return [...prev, ctx];
                })
              }
              onRemoveContext={(ctx) =>
                setSelectedContexts((prev) =>
                  prev.filter((c) => !(c.type === ctx.type && c.id === ctx.id))
                )
              }
              onSubmit={chat.handleSubmit}
              onKeyDown={chat.handleKeyDown}
              onChange={chat.handleInputChange}
              onStop={chat.stopStreaming}
            />
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center">
            <div className="text-center space-y-4">
              <div className="flex items-center justify-center size-16 rounded-2xl bg-primary/10 mx-auto">
                <Sparkles className="size-8 text-primary" />
              </div>
              <h3 className="text-lg font-semibold">{t('ai.emptyState.title')}</h3>
              <p className="text-sm text-muted-foreground max-w-sm">
                {t('ai.emptyState.description')}
              </p>
              <Button onClick={handleNewChat} className="gap-2">
                <Plus className="size-4" />
                {t('ai.threads.newChat')}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

interface ThreadSidebarProps {
  threads: ChatThread[];
  activeThreadId: string | null;
  resourceId?: string;
  isLoading?: boolean;
  isCollapsed?: boolean;
  onToggleCollapse: () => void;
  onSelectThread: (id: string) => void;
  onNewChat: () => void;
  onDeleteThread: (id: string) => void;
}

function ThreadSidebar({
  threads,
  activeThreadId,
  resourceId,
  isLoading,
  isCollapsed,
  onToggleCollapse,
  onSelectThread,
  onNewChat,
  onDeleteThread
}: ThreadSidebarProps) {
  const { t } = useTranslation();
  const [searchInputValue, setSearchInputValue] = useState('');
  const memorySearch = useMemorySearch(resourceId);

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
        </TooltipProvider>
        <Separator className="my-1" />
        <ScrollArea className="flex-1 w-full">
          <div className="flex flex-col items-center gap-0.5 px-1">
            <TooltipProvider delayDuration={0}>
              {isLoading
                ? [...Array(3)].map((_, i) => <Skeleton key={i} className="size-8 rounded-md" />)
                : threads.map((thread) => (
                    <Tooltip key={thread.id}>
                      <TooltipTrigger asChild>
                        <Button
                          variant={thread.id === activeThreadId ? 'secondary' : 'ghost'}
                          size="icon"
                          className="size-8"
                          onClick={() => onSelectThread(thread.id)}
                        >
                          <MessageSquareText className="size-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent side="right">
                        {thread.title || t('ai.threads.untitledChat')}
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
    <div className="w-64 shrink-0 border-r border-border/50 flex flex-col bg-muted/20">
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
                onClick={onToggleCollapse}
              >
                <PanelLeftClose className="size-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">{t('ai.threads.collapseSidebar')}</TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
      <div className="px-2 py-2">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
          <Input
            value={searchInputValue}
            onChange={(e) => {
              setSearchInputValue(e.target.value);
              if (!e.target.value.trim()) memorySearch.clear();
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                memorySearch.search(searchInputValue);
              }
              if (e.key === 'Escape') {
                setSearchInputValue('');
                memorySearch.clear();
              }
            }}
            placeholder={t('ai.threads.searchChats' as Parameters<typeof t>[0])}
            className="h-8 pl-7 text-xs"
          />
        </div>
      </div>
      <Separator />
      <div className="px-3 py-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {memorySearch.results.length > 0
            ? t('ai.threads.searchResults' as Parameters<typeof t>[0])
            : t('ai.threads.recentChats')}
        </span>
      </div>
      <ScrollArea className="flex-1">
        <div className="px-2 pb-2 space-y-0.5">
          {memorySearch.results.length > 0 ? (
            memorySearch.isSearching ? (
              <ThreadsSkeleton />
            ) : (
              memorySearch.results.map((r) => (
                <button
                  key={r.id}
                  type="button"
                  onClick={() => {
                    if (r.threadId) {
                      onSelectThread(r.threadId);
                      setSearchInputValue('');
                      memorySearch.clear();
                    }
                  }}
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
          ) : threads.length === 0 ? (
            <div className="px-3 py-8 text-center">
              <MessageSquare className="size-8 text-muted-foreground/40 mx-auto mb-2" />
              <p className="text-xs text-muted-foreground">{t('ai.threads.emptyState.title')}</p>
              <p className="text-xs text-muted-foreground/60 mt-1">
                {t('ai.threads.emptyState.description')}
              </p>
            </div>
          ) : (
            threads.map((thread) => (
              <ThreadItem
                key={thread.id}
                thread={thread}
                isActive={thread.id === activeThreadId}
                onSelect={() => onSelectThread(thread.id)}
                onDelete={() => onDeleteThread(thread.id)}
              />
            ))
          )}
        </div>
        <ScrollBar />
      </ScrollArea>
    </div>
  );
}

interface ThreadItemProps {
  thread: ChatThread;
  isActive: boolean;
  onSelect: () => void;
  onDelete: () => void;
}

function ThreadItem({ thread, isActive, onSelect, onDelete }: ThreadItemProps) {
  return (
    <button
      onClick={onSelect}
      className={cn(
        'w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors group text-left',
        isActive
          ? 'bg-primary/10 text-primary font-medium'
          : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
      )}
    >
      <MessageSquare className="size-4 shrink-0" />
      <span className="flex-1 truncate">{thread.title}</span>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <span
              role="button"
              onClick={(e) => {
                e.stopPropagation();
                onDelete();
              }}
              className="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 rounded hover:bg-destructive/10 hover:text-destructive"
            >
              <Trash2 className="size-3.5" />
            </span>
          </TooltipTrigger>
          <TooltipContent side="right">Delete</TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </button>
  );
}

interface ChatMessagesProps {
  messages: ChatMessage[];
  isStreaming: boolean;
  scrollRef: React.RefObject<HTMLDivElement | null>;
  onSuggestionClick: (text: string) => void;
  pendingToolApproval?: PendingToolApproval | null;
  autoRunTools?: boolean;
  onApproveToolCall?: () => void;
  onDeclineToolCall?: () => void;
}

function ChatMessages({
  messages,
  isStreaming,
  scrollRef,
  onSuggestionClick,
  pendingToolApproval,
  autoRunTools,
  onApproveToolCall,
  onDeclineToolCall
}: ChatMessagesProps) {
  const { t } = useTranslation();
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <div ref={containerRef} className="flex-1 overflow-y-auto" {...({ ref: scrollRef } as any)}>
      <div className="max-w-3xl mx-auto px-4 py-6">
        {messages.length === 0 ? (
          <ChatEmptyState onSuggestionClick={onSuggestionClick} />
        ) : (
          <div className="space-y-6">
            {messages.map((message) => (
              <MessageBubble key={message.id} message={message} />
            ))}
            {isStreaming && messages[messages.length - 1]?.role !== 'assistant' && (
              <StreamingIndicator />
            )}
            {pendingToolApproval && !autoRunTools && (
              <ToolApprovalBanner
                pending={pendingToolApproval}
                onApprove={onApproveToolCall}
                onDecline={onDeclineToolCall}
              />
            )}
          </div>
        )}
      </div>
    </div>
  );
}

interface ToolApprovalBannerProps {
  pending: PendingToolApproval;
  onApprove?: () => void;
  onDecline?: () => void;
}

function ToolApprovalBanner({ pending, onApprove, onDecline }: ToolApprovalBannerProps) {
  const { t } = useTranslation();
  const argsPreview =
    typeof pending.args === 'object' && pending.args !== null
      ? JSON.stringify(pending.args).slice(0, 100)
      : String(pending.args);

  return (
    <div className="flex flex-col gap-2 p-4 rounded-xl border border-amber-500/30 bg-amber-500/5">
      <p className="text-sm font-medium">{t('ai.toolApproval.title' as Parameters<typeof t>[0])}</p>
      <p className="text-xs text-muted-foreground">
        <span className="font-medium">{pending.toolName}</span>
        {argsPreview && ` — ${argsPreview}${argsPreview.length >= 100 ? '…' : ''}`}
      </p>
      <div className="flex gap-2 mt-1">
        <Button size="sm" onClick={onApprove} className="gap-1.5">
          <Check className="size-4" />
          {t('ai.toolApproval.approve' as Parameters<typeof t>[0])}
        </Button>
        <Button size="sm" variant="outline" onClick={onDecline} className="gap-1.5">
          <X className="size-4" />
          {t('ai.toolApproval.decline' as Parameters<typeof t>[0])}
        </Button>
      </div>
    </div>
  );
}

interface ChatEmptyStateProps {
  onSuggestionClick: (text: string) => void;
}

function ChatEmptyState({ onSuggestionClick }: ChatEmptyStateProps) {
  const { t } = useTranslation();

  const suggestions = [
    t('ai.suggestions.deploy'),
    t('ai.suggestions.logs'),
    t('ai.suggestions.envVars')
  ];

  return (
    <div className="flex flex-col items-center justify-center py-20 px-4">
      <div className="flex items-center justify-center size-16 rounded-2xl bg-primary/10 mb-6">
        <NixopusIcon className="size-8" />
      </div>
      <h3 className="text-lg font-semibold text-foreground mb-2">{t('ai.emptyState.title')}</h3>
      <p className="text-sm text-muted-foreground text-center max-w-sm mb-2">
        {t('ai.emptyState.description')}
      </p>
      <p className="text-xs text-muted-foreground/60 mb-6">{t('ai.emptyState.tryAskingAbout')}</p>
      <div className="grid grid-cols-1 gap-2 w-full max-w-sm">
        {suggestions.map((suggestion, index) => (
          <button
            key={index}
            type="button"
            onClick={() => onSuggestionClick(suggestion)}
            className="px-4 py-3 text-sm text-left rounded-lg border border-border/50 bg-muted/30 hover:bg-muted/60 hover:border-border transition-colors text-muted-foreground hover:text-foreground"
          >
            {suggestion}
          </button>
        ))}
      </div>
    </div>
  );
}

interface MessageBubbleProps {
  message: ChatMessage;
}

function MessageBubble({ message }: MessageBubbleProps) {
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
          {isUser ? <User className="size-4" /> : <NixopusIcon className="size-4" />}
        </AvatarFallback>
      </Avatar>
      <div className={cn('flex-1 max-w-[85%] flex flex-col', isUser ? 'items-end' : 'items-start')}>
        {isUser && message.contexts && message.contexts.length > 0 && (
          <div className="flex items-center gap-1.5 mb-1 px-1 flex-wrap">
            {message.contexts.map((ctx) => (
              <span
                key={`${ctx.type}-${ctx.id}`}
                className="inline-flex items-center gap-1 text-xs text-muted-foreground"
              >
                <span className="font-medium">{ctx.label}</span>
                {ctx.meta?.Environment && (
                  <span className="text-muted-foreground/60">{ctx.meta.Environment}</span>
                )}
                {ctx.meta?.Status && (
                  <span className="text-muted-foreground/60">{ctx.meta.Status}</span>
                )}
              </span>
            ))}
          </div>
        )}
        <div
          className={cn(
            'rounded-2xl px-4 py-3',
            isUser
              ? 'bg-primary text-primary-foreground rounded-tr-md'
              : 'bg-muted/60 text-foreground rounded-tl-md'
          )}
        >
          {isUser ? (
            <p className="text-sm whitespace-pre-wrap">
              {stripContextFromMessageText(message.content)}
            </p>
          ) : (
            <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-headings:my-2">
              <Streamdown components={CHAT_STREAMDOWN_COMPONENTS} isAnimating={false}>
                {message.content}
              </Streamdown>
            </div>
          )}
        </div>
        <span
          className={cn(
            'text-xs text-muted-foreground mt-1 px-1',
            isUser ? 'text-right' : 'text-left'
          )}
        >
          {formatTime(message.timestamp)}
        </span>
      </div>
    </div>
  );
}

function StreamingIndicator() {
  return (
    <div className="flex gap-3">
      <Avatar className="size-8 shrink-0">
        <AvatarFallback className="bg-muted text-muted-foreground">
          <NixopusIcon className="size-4" />
        </AvatarFallback>
      </Avatar>
      <div className="flex-1">
        <div className="bg-muted/60 rounded-2xl rounded-tl-md px-4 py-3 inline-block">
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-primary/60 animate-pulse" />
            <span
              className="size-2 rounded-full bg-primary/60 animate-pulse"
              style={{ animationDelay: '150ms' }}
            />
            <span
              className="size-2 rounded-full bg-primary/60 animate-pulse"
              style={{ animationDelay: '300ms' }}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

interface ContextSubMenuProps {
  provider: ContextProviderData;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  selectedContexts: ChatContext[];
  onAddContext: (ctx: ChatContext) => void;
  onRemoveContext: (ctx: ChatContext) => void;
}

function ContextSubMenu({
  provider,
  icon: Icon,
  label,
  selectedContexts,
  onAddContext,
  onRemoveContext
}: ContextSubMenuProps) {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');

  const isSelected = (ctx: ChatContext) =>
    selectedContexts.some((c) => c.type === ctx.type && c.id === ctx.id);

  const filtered = search.trim()
    ? provider.items.filter((item) => item.label.toLowerCase().includes(search.toLowerCase()))
    : provider.items;

  return (
    <DropdownMenuSub>
      <DropdownMenuSubTrigger className="flex items-center gap-2">
        <Icon className="size-4" />
        <span>{label}</span>
        {provider.isLoading && <Loader2 className="size-3 animate-spin ml-auto" />}
      </DropdownMenuSubTrigger>
      <DropdownMenuSubContent className="w-64 p-0">
        <div className="p-2 border-b border-border/50">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearch(e.target.value)}
              placeholder={`${t('ai.context.search' as Parameters<typeof t>[0])}...`}
              className="h-8 pl-7 text-xs"
              onClick={(e: React.MouseEvent) => e.stopPropagation()}
              onKeyDown={(e: React.KeyboardEvent) => e.stopPropagation()}
            />
          </div>
        </div>
        <div className="max-h-56 overflow-y-auto py-1">
          {provider.isLoading ? (
            <div className="flex items-center justify-center py-4">
              <Loader2 className="size-4 animate-spin text-muted-foreground" />
            </div>
          ) : filtered.length === 0 ? (
            <div className="px-2 py-4 text-center text-xs text-muted-foreground">
              {t('ai.context.noItems')}
            </div>
          ) : (
            filtered.map((item) => {
              const selected = isSelected(item);
              return (
                <DropdownMenuItem
                  key={item.id}
                  onClick={(e) => {
                    e.preventDefault();
                    if (selected) {
                      onRemoveContext(item);
                    } else {
                      onAddContext(item);
                    }
                  }}
                  className={cn('flex items-center gap-2 mx-1', selected && 'bg-primary/10')}
                >
                  <Check
                    className={cn(
                      'size-3.5 shrink-0',
                      selected ? 'opacity-100 text-primary' : 'opacity-0'
                    )}
                  />
                  <span className="truncate flex-1">{item.label}</span>
                  {item.meta?.Environment && (
                    <span className="text-xs text-muted-foreground shrink-0">
                      {item.meta.Environment}
                    </span>
                  )}
                  {item.meta?.Status && (
                    <span className="text-xs text-muted-foreground shrink-0">
                      {item.meta.Status}
                    </span>
                  )}
                </DropdownMenuItem>
              );
            })
          )}
        </div>
      </DropdownMenuSubContent>
    </DropdownMenuSub>
  );
}

interface ChatInputProps {
  inputValue: string;
  isStreaming: boolean;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  selectedContexts: ChatContext[];
  contextProviders: ContextProviderData[];
  autoRunTools: boolean;
  onAutoRunToolsChange: (value: boolean) => void;
  onAddContext: (ctx: ChatContext) => void;
  onRemoveContext: (ctx: ChatContext) => void;
  onSubmit: (e?: React.FormEvent) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  onStop: () => void;
}

function ChatInput({
  inputValue,
  isStreaming,
  textareaRef,
  selectedContexts,
  contextProviders,
  autoRunTools,
  onAutoRunToolsChange,
  onAddContext,
  onRemoveContext,
  onSubmit,
  onKeyDown,
  onChange,
  onStop
}: ChatInputProps) {
  const { t } = useTranslation();

  const isSelected = (ctx: ChatContext) =>
    selectedContexts.some((c) => c.type === ctx.type && c.id === ctx.id);

  return (
    <div className="shrink-0 border-t border-border/50 bg-background/80 backdrop-blur-sm p-4">
      <div className="max-w-3xl mx-auto">
        <div className="mb-2 flex items-center gap-2 flex-wrap">
          {selectedContexts.map((ctx) => {
            const provider = contextProviders.find((p) => p.config.type === ctx.type);
            const Icon = provider?.config.icon;
            return (
              <span
                key={`${ctx.type}-${ctx.id}`}
                className="inline-flex items-center gap-1.5 pl-2 pr-1 py-0.5 rounded-md text-xs font-medium bg-primary/10 text-primary border border-primary/20"
              >
                {Icon && <Icon className="size-3" />}
                <span className="truncate max-w-[150px]">{ctx.label}</span>
                {ctx.meta?.Environment && (
                  <span className="text-primary/60">{ctx.meta.Environment}</span>
                )}
                {ctx.meta?.Status && <span className="text-primary/60">{ctx.meta.Status}</span>}
                <button
                  type="button"
                  onClick={() => onRemoveContext(ctx)}
                  className="ml-0.5 p-0.5 rounded hover:bg-primary/20 transition-colors"
                >
                  <X className="size-3" />
                </button>
              </span>
            );
          })}
          <div className="flex items-center gap-1 rounded-md border border-border/50 p-0.5">
            <button
              type="button"
              onClick={() => onAutoRunToolsChange(false)}
              className={cn(
                'px-2 py-0.5 rounded text-xs transition-colors',
                !autoRunTools
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {t('ai.toolExecution.askBefore' as Parameters<typeof t>[0])}
            </button>
            <button
              type="button"
              onClick={() => onAutoRunToolsChange(true)}
              className={cn(
                'px-2 py-0.5 rounded text-xs transition-colors',
                autoRunTools
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {t('ai.toolExecution.autoRun' as Parameters<typeof t>[0])}
            </button>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-muted border border-border/50 transition-colors"
              >
                <CirclePlus className="size-3" />
                {t('ai.context.addContext')}
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-48">
              {contextProviders.map((provider) => {
                const Icon = provider.config.icon;
                return (
                  <ContextSubMenu
                    key={provider.config.type}
                    provider={provider}
                    icon={Icon}
                    label={t(provider.config.labelKey as Parameters<typeof t>[0])}
                    selectedContexts={selectedContexts}
                    onAddContext={onAddContext}
                    onRemoveContext={onRemoveContext}
                  />
                );
              })}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <form onSubmit={onSubmit} className="flex gap-3 items-end">
          <Textarea
            ref={textareaRef}
            value={inputValue}
            onChange={onChange}
            onKeyDown={onKeyDown}
            placeholder={t('ai.input.placeholder')}
            className="min-h-[44px] max-h-[120px] resize-none bg-muted/50 border-border/50 focus-visible:ring-primary/30"
            disabled={isStreaming}
            rows={1}
          />
          {isStreaming ? (
            <Button
              type="button"
              size="icon"
              variant="destructive"
              onClick={onStop}
              className="shrink-0 size-11 rounded-lg"
            >
              <StopCircle className="size-4" />
            </Button>
          ) : (
            <Button
              type="submit"
              size="icon"
              disabled={!inputValue.trim()}
              className="shrink-0 size-11 rounded-lg"
            >
              <Send className="size-4" />
            </Button>
          )}
        </form>
        <p className="text-xs text-muted-foreground mt-3 text-center">{t('ai.input.hint')}</p>
      </div>
    </div>
  );
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

function MessagesSkeleton() {
  return (
    <div className="flex-1 overflow-hidden">
      <div className="max-w-3xl mx-auto px-4 py-6 space-y-6">
        <div className="flex gap-3 flex-row-reverse">
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <div className="flex-1 flex flex-col items-end space-y-2">
            <Skeleton className="h-10 w-48 rounded-2xl" />
          </div>
        </div>
        <div className="flex gap-3">
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
            <Skeleton className="h-4 w-2/3" />
          </div>
        </div>
        <div className="flex gap-3 flex-row-reverse">
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <div className="flex-1 flex flex-col items-end space-y-2">
            <Skeleton className="h-10 w-64 rounded-2xl" />
          </div>
        </div>
        <div className="flex gap-3">
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-5/6" />
            <Skeleton className="h-4 w-2/3" />
          </div>
        </div>
      </div>
    </div>
  );
}

function AgentDisabledState() {
  return (
    <div className="flex h-full w-full items-center justify-center">
      <div className="text-center max-w-md space-y-4 px-4">
        <div className="flex items-center justify-center size-16 rounded-2xl bg-muted mx-auto">
          <Sparkles className="size-8 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-semibold">AI Agent Not Configured</h3>
        <p className="text-sm text-muted-foreground">
          The AI-powered deployment assistant is not enabled on this instance. To get access, reach
          out to us and we&apos;ll help you get set up.
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

function formatTime(date: Date): string {
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
