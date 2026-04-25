'use client';

import React from 'react';
import Image from 'next/image';
import { useTheme } from 'next-themes';
import { Streamdown } from 'streamdown';
import {
  STREAMDOWN_PLUGINS,
  STREAMDOWN_CONTROLS,
  STREAMDOWN_ANIMATED
} from '@/packages/lib/streamdown-config';
import {
  Button,
  Textarea,
  Avatar,
  AvatarFallback,
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
  Send,
  Loader2,
  User,
  StopCircle,
  X,
  CirclePlus,
  Search,
  Check,
  ChevronRight,
  ChevronDown,
  Copy,
  Zap
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  type ChatMessage,
  type MessagePart,
  type PendingToolApproval,
  type OmStatus,
  type AgentQuestion,
  type AgentQuestionField
} from '@/packages/hooks/ai/use-agent-chat';
import { ContextWindowBar } from './context-window-bar';
import {
  type ChatContext,
  type ContextProviderData,
  stripContextFromMessageText
} from '@/packages/hooks/ai/chat-context';
import {
  getActiveMentionToken,
  flattenMentionItems,
  filterMentionItems,
  replaceMentionTokenWithCaret,
  type MentionItem,
  type MentionMatch
} from '@/packages/hooks/ai/chat-mentions';
import {
  useChatPage,
  useChatMessagesScroll,
  useContextSearch,
  formatTime
} from '@/packages/hooks/ai/use-chat-page';

function ChatMentionSuggestions({
  open,
  items,
  highlightIndex,
  onHoverIndex,
  onPick,
  noResultsText
}: {
  open: boolean;
  items: MentionItem[];
  highlightIndex: number;
  onHoverIndex: (index: number) => void;
  onPick: (item: MentionItem) => void;
  noResultsText: string;
}) {
  const listRef = React.useRef<HTMLDivElement | null>(null);
  const optionRefs = React.useRef<Array<HTMLButtonElement | null>>([]);

  React.useEffect(() => {
    if (!open || items.length === 0) return;
    const option = optionRefs.current[highlightIndex];
    if (!option) return;
    option.scrollIntoView({ block: 'nearest' });
  }, [open, highlightIndex, items.length]);

  React.useEffect(() => {
    optionRefs.current = optionRefs.current.slice(0, items.length);
  }, [items.length]);

  if (!open) return null;

  return (
    <div
      ref={listRef}
      role="listbox"
      className="z-10 mt-1 max-h-48 overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden rounded-lg border border-border bg-popover text-popover-foreground shadow-md"
    >
      {items.length === 0 ? (
        <div className="px-3 py-2.5 text-xs text-muted-foreground">{noResultsText}</div>
      ) : (
        items.map((item, index) => (
          <button
            ref={(el) => {
              optionRefs.current[index] = el;
            }}
            key={item.key}
            type="button"
            role="option"
            aria-selected={index === highlightIndex}
            onMouseEnter={() => onHoverIndex(index)}
            onMouseDown={(e) => {
              e.preventDefault();
              onPick(item);
            }}
            className={cn(
              'flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors',
              index === highlightIndex
                ? 'bg-primary/10 text-foreground ring-1 ring-primary/30'
                : 'hover:bg-muted/60'
            )}
          >
            <span className="min-w-0 flex-1 truncate">{item.label}</span>
            <span className="shrink-0 text-xs text-muted-foreground">{item.type}</span>
          </button>
        ))
      )}
    </div>
  );
}

function useChatTextareaMentions({
  inputValue,
  setInputValue,
  contextProviders,
  onAddContext,
  textareaRef,
  onChange,
  onKeyDown
}: {
  inputValue: string;
  setInputValue: (value: string) => void;
  contextProviders: ContextProviderData[];
  onAddContext: (ctx: ChatContext) => void;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
}) {
  const { t } = useTranslation();
  const noResultsText = t('ai.mentions.noResults' as Parameters<typeof t>[0]);

  const [liveMatch, setLiveMatch] = React.useState<MentionMatch | null>(null);
  const [highlightIndex, setHighlightIndex] = React.useState(0);
  const highlightIndexRef = React.useRef(0);
  const [dismissedMentionStart, setDismissedMentionStart] = React.useState<number | null>(null);
  const pendingCaretAfterPickRef = React.useRef<{ pos: number; expectedValue: string } | null>(
    null
  );

  React.useEffect(() => {
    highlightIndexRef.current = highlightIndex;
  }, [highlightIndex]);

  const allMentionItems = React.useMemo(
    () => flattenMentionItems(contextProviders),
    [contextProviders]
  );

  const filteredItems = React.useMemo(() => {
    if (!liveMatch) return [];
    return filterMentionItems(allMentionItems, liveMatch.query);
  }, [allMentionItems, liveMatch]);

  const listOpen =
    liveMatch !== null &&
    !(dismissedMentionStart !== null && liveMatch.start === dismissedMentionStart);

  React.useEffect(() => {
    setHighlightIndex(0);
  }, [liveMatch?.start, liveMatch?.query]);

  const reconcileFromValueAndCursor = React.useCallback((value: string, cursor: number) => {
    const c = Math.max(0, Math.min(cursor, value.length));
    const m = getActiveMentionToken(value, c);
    setLiveMatch(m);
    setDismissedMentionStart((ds) => {
      if (ds === null) return null;
      if (!m || m.start !== ds) return null;
      return ds;
    });
  }, []);

  const syncLiveMatch = React.useCallback(
    (el: HTMLTextAreaElement) => {
      reconcileFromValueAndCursor(el.value, el.selectionStart ?? 0);
    },
    [reconcileFromValueAndCursor]
  );

  React.useLayoutEffect(() => {
    const pending = pendingCaretAfterPickRef.current;
    if (pending && inputValue !== pending.expectedValue) {
      pendingCaretAfterPickRef.current = null;
    }

    const el = textareaRef.current;
    const stillPending = pendingCaretAfterPickRef.current;

    if (stillPending && inputValue === stillPending.expectedValue) {
      pendingCaretAfterPickRef.current = null;
      const p = Math.max(0, Math.min(stillPending.pos, inputValue.length));
      if (el) {
        el.focus();
        el.setSelectionRange(p, p);
      }
      reconcileFromValueAndCursor(inputValue, p);
      return;
    }

    let cursor = inputValue.length;
    if (el && el.value === inputValue) {
      cursor = Math.min(el.selectionStart ?? inputValue.length, inputValue.length);
    }
    reconcileFromValueAndCursor(inputValue, cursor);
  }, [inputValue, reconcileFromValueAndCursor, textareaRef]);

  const pickItem = React.useCallback(
    (item: MentionItem) => {
      const el = textareaRef.current;
      if (!el) return;
      const match = getActiveMentionToken(el.value, el.selectionStart ?? 0);
      if (!match) return;

      const value = el.value;
      const { value: next, caret: pos } = replaceMentionTokenWithCaret(value, match);
      onAddContext(item.ctx);
      setDismissedMentionStart(null);

      pendingCaretAfterPickRef.current = { pos, expectedValue: next };
      setInputValue(next);
    },
    [onAddContext, setInputValue, textareaRef]
  );

  const wrappedOnChange = React.useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      onChange(e);
      syncLiveMatch(e.currentTarget);
    },
    [onChange, syncLiveMatch]
  );

  const wrappedOnKeyDown = React.useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      const el = e.currentTarget;
      const match = getActiveMentionToken(el.value, el.selectionStart ?? 0);
      const dismissed =
        dismissedMentionStart !== null && match && match.start === dismissedMentionStart;
      const active = Boolean(match && !dismissed);
      const items = match ? filterMentionItems(allMentionItems, match.query) : [];

      if (active && e.key === 'Escape') {
        e.preventDefault();
        e.stopPropagation();
        if (match) setDismissedMentionStart(match.start);
        return;
      }

      if (active && items.length > 0) {
        if (e.key === 'ArrowDown') {
          e.preventDefault();
          e.stopPropagation();
          setHighlightIndex((i) => Math.min(i + 1, items.length - 1));
          return;
        }
        if (e.key === 'ArrowUp') {
          e.preventDefault();
          e.stopPropagation();
          setHighlightIndex((i) => Math.max(i - 1, 0));
          return;
        }
        if (e.key === 'Enter' && !e.shiftKey) {
          e.preventDefault();
          e.stopPropagation();
          const idx = Math.min(highlightIndexRef.current, items.length - 1);
          const item = items[idx];
          if (item) pickItem(item);
          return;
        }
      }

      onKeyDown(e);
      if (e.key === 'ArrowLeft' || e.key === 'ArrowRight' || e.key === 'Home' || e.key === 'End') {
        queueMicrotask(() => {
          const ta = textareaRef.current;
          if (ta) syncLiveMatch(ta);
        });
      }
    },
    [allMentionItems, dismissedMentionStart, onKeyDown, pickItem, syncLiveMatch, textareaRef]
  );

  const wrappedOnSelect = React.useCallback(
    (e: React.SyntheticEvent<HTMLTextAreaElement>) => {
      syncLiveMatch(e.currentTarget);
    },
    [syncLiveMatch]
  );

  const mentionList = (
    <ChatMentionSuggestions
      open={listOpen}
      items={filteredItems}
      highlightIndex={Math.min(highlightIndex, Math.max(filteredItems.length - 1, 0))}
      onHoverIndex={setHighlightIndex}
      onPick={pickItem}
      noResultsText={noResultsText}
    />
  );

  return {
    mentionList,
    textareaHandlers: {
      onChange: wrappedOnChange,
      onKeyDown: wrappedOnKeyDown,
      onSelect: wrappedOnSelect
    }
  };
}

function NixopusIcon({ className }: { className?: string }) {
  const { resolvedTheme } = useTheme();
  const src = resolvedTheme === 'dark' ? '/logo_white.png' : '/logo_black.png';
  return <Image src={src} alt="Nixopus" width={16} height={16} className={className} />;
}

export function ChatPage() {
  const { t } = useTranslation();
  const page = useChatPage();
  const justCreatedThreadRef = React.useRef(false);
  const pendingMessageRef = React.useRef<string | null>(null);

  React.useEffect(() => {
    if (pendingMessageRef.current && page.activeThreadId && !page.isLoadingHistory) {
      const msg = pendingMessageRef.current;
      pendingMessageRef.current = null;
      page.handleSuggestionClick(msg);
    }
  }, [page.activeThreadId, page.isLoadingHistory, page.handleSuggestionClick]);

  React.useEffect(() => {
    if (!page.isLoadingHistory) {
      justCreatedThreadRef.current = false;
    }
  }, [page.isLoadingHistory]);

  const ensureThread = React.useCallback(() => {
    if (!page.activeThreadId) {
      justCreatedThreadRef.current = true;
      page.handleNewChat();
    }
  }, [page.activeThreadId, page.handleNewChat]);

  const handleWelcomeSubmit = React.useCallback(
    (e?: React.FormEvent) => {
      e?.preventDefault();
      const text = page.inputValue.trim();
      if (!text) return;

      if (page.activeThreadId) {
        page.handleSubmit(e);
      } else {
        pendingMessageRef.current = text;
        page.setInputValue('');
        justCreatedThreadRef.current = true;
        page.handleNewChat();
      }
    },
    [
      page.inputValue,
      page.activeThreadId,
      page.handleSubmit,
      page.handleNewChat,
      page.setInputValue
    ]
  );

  const handleWelcomeSuggestionClick = React.useCallback(
    (text: string) => {
      if (page.activeThreadId) {
        page.handleSuggestionClick(text);
      } else {
        pendingMessageRef.current = text;
        justCreatedThreadRef.current = true;
        page.handleNewChat();
      }
    },
    [page.activeThreadId, page.handleSuggestionClick, page.handleNewChat]
  );

  const handleWelcomeKeyDown = React.useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleWelcomeSubmit();
      }
    },
    [handleWelcomeSubmit]
  );

  const showWelcome =
    page.isThreadsInitialized &&
    page.messages.length === 0 &&
    (!page.isLoadingHistory || justCreatedThreadRef.current);

  return (
    <div className="flex h-full w-full overflow-hidden">
      <div className="flex flex-1 flex-col min-w-0">
        {page.activeQuestion && (
          <AgentQuestionModal
            question={page.activeQuestion}
            onSubmit={page.submitQuestionResponse}
            onDismiss={page.dismissQuestion}
          />
        )}
        {!page.isThreadsInitialized ? (
          <MessagesSkeleton />
        ) : page.messages.length > 0 ? (
          <>
            <ChatMessages
              messages={page.messages}
              isStreaming={page.isStreaming}
              scrollRef={page.scrollRef}
              pendingToolApproval={page.pendingToolApproval}
              autoRunTools={page.autoRunTools}
              onApproveToolCall={page.handleApproveToolCall}
              onDeclineToolCall={page.handleDeclineToolCall}
            />
            {page.omStatus && <ContextWindowBar omStatus={page.omStatus} />}
            {page.readOnly ? (
              <div className="shrink-0 border-t border-border/50 bg-amber-500/5 px-4 py-3 text-center">
                <p className="text-xs text-amber-600 dark:text-amber-400">
                  This is an automated incident thread and is read-only.
                </p>
              </div>
            ) : (
              <ChatInput
                inputValue={page.inputValue}
                setInputValue={page.setInputValue}
                isStreaming={page.isStreaming}
                textareaRef={page.textareaRef}
                selectedContexts={page.selectedContexts}
                contextProviders={page.contextProviders}
                autoRunTools={page.autoRunTools}
                onAutoRunToolsChange={page.setAutoRunTools}
                selectedModel={page.selectedModel}
                onModelChange={page.setSelectedModel}
                availableModels={page.availableModels}
                onAddContext={page.addContext}
                onRemoveContext={page.removeContext}
                onSubmit={page.handleSubmit}
                onKeyDown={page.handleKeyDown}
                onChange={page.handleInputChange}
                onStop={page.stopStreaming}
              />
            )}
          </>
        ) : page.isLoadingHistory && !justCreatedThreadRef.current ? (
          <MessagesSkeleton />
        ) : showWelcome ? (
          <ChatWelcomeView
            inputValue={page.inputValue}
            setInputValue={page.setInputValue}
            textareaRef={page.textareaRef}
            showGuidedPrefillHint={page.showGuidedPrefillHint}
            selectedContexts={page.selectedContexts}
            contextProviders={page.contextProviders}
            autoRunTools={page.autoRunTools}
            onAutoRunToolsChange={page.setAutoRunTools}
            selectedModel={page.selectedModel}
            onModelChange={page.setSelectedModel}
            availableModels={page.availableModels}
            onAddContext={page.addContext}
            onRemoveContext={page.removeContext}
            onSubmit={handleWelcomeSubmit}
            onKeyDown={handleWelcomeKeyDown}
            onChange={page.handleInputChange}
            onSuggestionClick={handleWelcomeSuggestionClick}
            onInputFocus={ensureThread}
          />
        ) : (
          <MessagesSkeleton />
        )}
      </div>
    </div>
  );
}

interface ChatMessagesProps {
  messages: ChatMessage[];
  isStreaming: boolean;
  scrollRef: React.RefObject<HTMLDivElement | null>;
  pendingToolApproval?: PendingToolApproval | null;
  autoRunTools?: boolean;
  onApproveToolCall?: () => void;
  onDeclineToolCall?: () => void;
}

function ChatMessages({
  messages,
  isStreaming,
  scrollRef,
  pendingToolApproval,
  autoRunTools,
  onApproveToolCall,
  onDeclineToolCall
}: ChatMessagesProps) {
  const { containerRef } = useChatMessagesScroll(messages);

  return (
    <div ref={containerRef} className="flex-1 overflow-y-auto" {...({ ref: scrollRef } as any)}>
      <div className="max-w-3xl mx-auto px-4 py-6">
        <div className="space-y-6">
          {messages.map((message, index) => {
            const isLastAssistant = message.role === 'assistant' && index === messages.length - 1;
            return (
              <MessageBubble
                key={message.id}
                message={message}
                isStreaming={isStreaming}
                isLastAssistantMessage={isLastAssistant}
              />
            );
          })}
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

interface ChatWelcomeViewProps {
  inputValue: string;
  setInputValue: (value: string) => void;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  selectedContexts: ChatContext[];
  contextProviders: ContextProviderData[];
  autoRunTools: boolean;
  onAutoRunToolsChange: (value: boolean) => void;
  selectedModel: string;
  onModelChange: (model: string) => void;
  availableModels: readonly { id: string; label: string }[];
  onAddContext: (ctx: ChatContext) => void;
  onRemoveContext: (ctx: ChatContext) => void;
  onSubmit: (e?: React.FormEvent) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  onSuggestionClick: (text: string) => void;
  onInputFocus?: () => void;
  showGuidedPrefillHint: boolean;
}

function ChatWelcomeView({
  inputValue,
  setInputValue,
  textareaRef,
  selectedContexts,
  contextProviders,
  autoRunTools,
  onAutoRunToolsChange,
  selectedModel,
  onModelChange,
  availableModels,
  onAddContext,
  onRemoveContext,
  onSubmit,
  onKeyDown,
  onChange,
  onSuggestionClick,
  onInputFocus,
  showGuidedPrefillHint
}: ChatWelcomeViewProps) {
  const { t } = useTranslation();
  const mentions = useChatTextareaMentions({
    inputValue,
    setInputValue,
    contextProviders,
    onAddContext,
    textareaRef,
    onChange,
    onKeyDown
  });
  const currentModelLabel =
    availableModels.find((m) => m.id === selectedModel)?.label ?? selectedModel;

  const suggestions = showGuidedPrefillHint
    ? [
        t('ai.guidedPrefill.suggestions.overview'),
        t('ai.guidedPrefill.suggestions.deploy'),
        t('ai.guidedPrefill.suggestions.tour')
      ]
    : [t('ai.suggestions.deploy'), t('ai.suggestions.logs'), t('ai.suggestions.envVars')];
  const welcomeTitle = showGuidedPrefillHint
    ? t('ai.emptyState.title')
    : t('ai.emptyState.returningTitle');
  const welcomePlaceholder = showGuidedPrefillHint
    ? t('ai.input.placeholder')
    : t('ai.input.returningPlaceholder');

  return (
    <div className="flex flex-1 flex-col items-center justify-center px-4 pb-12">
      <div className="w-full max-w-2xl">
        <div className="flex flex-col items-center gap-3 mb-8">
          <div className="flex items-center justify-center size-12 rounded-2xl bg-primary/10">
            <NixopusIcon className="size-6" />
          </div>
          <h2 className="text-xl font-semibold text-foreground">{welcomeTitle}</h2>
        </div>

        <div className="rounded-2xl border border-border bg-muted/20 shadow-sm focus-within:border-primary/30 focus-within:shadow-md transition-all">
          <form onSubmit={onSubmit}>
            <Textarea
              ref={textareaRef}
              value={inputValue}
              onChange={mentions.textareaHandlers.onChange}
              onKeyDown={mentions.textareaHandlers.onKeyDown}
              onSelect={mentions.textareaHandlers.onSelect}
              onFocus={onInputFocus}
              placeholder={welcomePlaceholder}
              className="min-h-[72px] max-h-[220px] resize-none overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden border-0 bg-transparent px-4 pt-4 pb-2 text-base focus-visible:ring-0 focus-visible:ring-offset-0 shadow-none"
              rows={1}
            />
            <div className="px-4">{mentions.mentionList}</div>
            <div className="flex items-center justify-between px-3 pb-3">
              <div className="flex items-center gap-2 flex-wrap">
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
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className="inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-colors"
                    >
                      <Zap className="size-3" />
                      <span>{currentModelLabel}</span>
                      <ChevronDown className="size-3 opacity-50" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start" className="w-52">
                    {availableModels.map((model) => (
                      <DropdownMenuItem
                        key={model.id}
                        onClick={() => onModelChange(model.id)}
                        className={cn(
                          'flex items-center gap-2',
                          selectedModel === model.id && 'bg-primary/10'
                        )}
                      >
                        <Check
                          className={cn(
                            'size-3.5 shrink-0',
                            selectedModel === model.id ? 'opacity-100 text-primary' : 'opacity-0'
                          )}
                        />
                        <span>{model.label}</span>
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
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
                      className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-colors"
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
              <Button
                type="submit"
                size="icon"
                disabled={!inputValue.trim()}
                className="shrink-0 size-9 rounded-xl"
              >
                <Send className="size-4" />
              </Button>
            </div>
          </form>
        </div>

        <div className="flex flex-wrap items-center justify-center gap-2 mt-4">
          {suggestions.map((suggestion, index) => (
            <button
              key={index}
              type="button"
              onClick={() => onSuggestionClick(suggestion)}
              className="px-3.5 py-2 text-sm rounded-full border border-border/50 bg-background hover:bg-muted/60 text-muted-foreground hover:text-foreground transition-colors"
            >
              {suggestion}
            </button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground mt-4 text-center">{t('ai.input.hint')}</p>
      </div>
    </div>
  );
}

function formatToolName(name: string): string {
  return name
    .replace(/[-_]/g, ' ')
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/^./, (s) => s.toUpperCase());
}

function getToolArgsSummary(args: unknown): string | null {
  if (!args || typeof args !== 'object') return null;
  const a = args as Record<string, unknown>;
  if (a.name && typeof a.name === 'string') return a.name;
  if (a.owner && a.repo) return `${a.owner}/${a.repo}`;
  if (a.id && typeof a.id === 'string') return a.id.length > 12 ? a.id.slice(0, 8) + '...' : a.id;
  return null;
}

function ToolCallIndicator({ part }: { part: Extract<MessagePart, { type: 'tool-call' }> }) {
  const [expanded, setExpanded] = React.useState(false);
  const isRunning = part.status === 'running';
  const name = formatToolName(part.toolName);
  const summary = getToolArgsSummary(part.args);
  const argsObj =
    part.args && typeof part.args === 'object' ? (part.args as Record<string, unknown>) : null;
  const hasDetails = argsObj !== null && Object.keys(argsObj).length > 0;

  return (
    <div className="text-xs text-muted-foreground/70">
      <button
        type="button"
        onClick={() => hasDetails && setExpanded((v) => !v)}
        className={cn(
          'flex items-center gap-1.5 py-1 px-1 rounded-md transition-colors w-full text-left',
          hasDetails && 'hover:text-muted-foreground hover:bg-muted/40 cursor-pointer'
        )}
      >
        {isRunning ? (
          <Loader2 className="size-3 animate-spin shrink-0 text-primary" />
        ) : (
          <Check className="size-3 shrink-0 text-muted-foreground/50" />
        )}
        <span className={cn(isRunning && 'text-muted-foreground')}>{name}</span>
        {summary && <span className="text-muted-foreground/40">— {summary}</span>}
        {hasDetails && (
          <ChevronRight
            className={cn(
              'size-3 shrink-0 ml-auto transition-transform text-muted-foreground/40',
              expanded && 'rotate-90'
            )}
          />
        )}
      </button>
      {expanded && hasDetails && (
        <pre className="mt-1 ml-5 p-2 rounded-md bg-muted/30 text-[10px] leading-relaxed text-muted-foreground/60 overflow-x-auto max-h-32">
          {JSON.stringify(part.args, null, 2)}
        </pre>
      )}
    </div>
  );
}

function CollapsibleTextPart({ content }: { content: string }) {
  const [expanded, setExpanded] = React.useState(false);
  const firstLine = content
    .split('\n')[0]
    .replace(/^[#*>\s-]+/, '')
    .trim();
  const preview = firstLine.slice(0, 120);

  return (
    <div className="text-xs text-muted-foreground/60">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex items-center gap-1.5 py-0.5 px-1 rounded-md hover:text-muted-foreground hover:bg-muted/40 transition-colors w-full text-left"
      >
        <ChevronRight
          className={cn(
            'size-3 shrink-0 transition-transform text-muted-foreground/40',
            expanded && 'rotate-90'
          )}
        />
        <span className="truncate">
          {preview}
          {!expanded && firstLine.length > 120 && '…'}
        </span>
      </button>
      {expanded && (
        <div className="mt-1 ml-5 text-sm text-foreground">
          <Streamdown
            plugins={STREAMDOWN_PLUGINS}
            controls={STREAMDOWN_CONTROLS}
            animated={STREAMDOWN_ANIMATED}
          >
            {content}
          </Streamdown>
        </div>
      )}
    </div>
  );
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard access denied
    }
  };

  return (
    <TooltipProvider delayDuration={0}>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            type="button"
            onClick={handleCopy}
            className="inline-flex items-center justify-center size-7 rounded-md text-muted-foreground/50 hover:text-muted-foreground hover:bg-muted/60 transition-colors"
          >
            {copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
          </button>
        </TooltipTrigger>
        <TooltipContent side="top" className="text-xs">
          {copied ? 'Copied!' : 'Copy'}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

interface MessageBubbleProps {
  message: ChatMessage;
  isStreaming?: boolean;
  isLastAssistantMessage?: boolean;
}

function MessageBubble({
  message,
  isStreaming = false,
  isLastAssistantMessage = false
}: MessageBubbleProps) {
  const isUser = message.role === 'user';
  const hasParts = !isUser && message.parts && message.parts.length > 0;

  if (hasParts) {
    const lastTextIndex = message.parts!.reduce(
      (acc: number, p: MessagePart, i: number) => (p.type === 'text' ? i : acc),
      -1
    );

    return (
      <div className="flex gap-3">
        <Avatar className="size-8 shrink-0 mt-0.5">
          <AvatarFallback className="bg-muted text-muted-foreground text-xs font-medium">
            <NixopusIcon className="size-4" />
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 max-w-[85%] flex flex-col gap-1">
          {message.parts!.map((part, index) => {
            if (part.type === 'text' && part.content) {
              const isLastText = index === lastTextIndex;
              const isActivelyStreaming = isStreaming && isLastAssistantMessage && isLastText;

              if (!isLastText && !isActivelyStreaming) {
                return <CollapsibleTextPart key={index} content={part.content} />;
              }

              return (
                <div key={index} className="text-sm text-foreground">
                  <Streamdown
                    plugins={STREAMDOWN_PLUGINS}
                    controls={STREAMDOWN_CONTROLS}
                    animated={STREAMDOWN_ANIMATED}
                    isAnimating={isActivelyStreaming}
                    caret={isActivelyStreaming ? 'block' : undefined}
                  >
                    {part.content}
                  </Streamdown>
                </div>
              );
            }
            if (part.type === 'tool-call') {
              return <ToolCallIndicator key={index} part={part} />;
            }
            return null;
          })}
          <div className="flex items-center gap-1 mt-1 px-1">
            <span className="text-xs text-muted-foreground">{formatTime(message.timestamp)}</span>
            <CopyButton text={message.content} />
          </div>
        </div>
      </div>
    );
  }

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
                {ctx.meta?.Language && (
                  <span className="text-muted-foreground/60">{ctx.meta.Language}</span>
                )}
                {ctx.meta?.Provider && (
                  <span className="text-muted-foreground/60">{ctx.meta.Provider}</span>
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
          ) : isStreaming && isLastAssistantMessage && !message.content.trim() ? (
            <span className="text-sm text-muted-foreground">
              Thinking
              <span className="inline-flex">
                <span className="animate-pulse" style={{ animationDelay: '0ms' }}>
                  .
                </span>
                <span className="animate-pulse" style={{ animationDelay: '150ms' }}>
                  .
                </span>
                <span className="animate-pulse" style={{ animationDelay: '300ms' }}>
                  .
                </span>
              </span>
            </span>
          ) : (
            <Streamdown
              plugins={STREAMDOWN_PLUGINS}
              controls={STREAMDOWN_CONTROLS}
              animated={STREAMDOWN_ANIMATED}
              isAnimating={isStreaming && isLastAssistantMessage}
              caret={isStreaming && isLastAssistantMessage ? 'block' : undefined}
            >
              {message.content}
            </Streamdown>
          )}
        </div>
        <div
          className={cn(
            'flex items-center gap-1 mt-1 px-1',
            isUser ? 'justify-end' : 'justify-start'
          )}
        >
          <span className="text-xs text-muted-foreground">{formatTime(message.timestamp)}</span>
          {!isUser && <CopyButton text={message.content} />}
        </div>
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
          <span className="text-sm text-muted-foreground">
            Thinking
            <span className="inline-flex">
              <span className="animate-pulse" style={{ animationDelay: '0ms' }}>
                .
              </span>
              <span className="animate-pulse" style={{ animationDelay: '150ms' }}>
                .
              </span>
              <span className="animate-pulse" style={{ animationDelay: '300ms' }}>
                .
              </span>
            </span>
          </span>
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
  const { search, setSearch, filtered } = useContextSearch(provider.items);

  const isSelected = (ctx: ChatContext) =>
    selectedContexts.some((c) => c.type === ctx.type && c.id === ctx.id);

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
                  {item.meta?.Language && (
                    <span className="text-xs text-muted-foreground shrink-0">
                      {item.meta.Language}
                    </span>
                  )}
                  {item.meta?.Provider && (
                    <span className="text-xs text-muted-foreground shrink-0">
                      {item.meta.Provider}
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
  setInputValue: (value: string) => void;
  isStreaming: boolean;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  selectedContexts: ChatContext[];
  contextProviders: ContextProviderData[];
  autoRunTools: boolean;
  onAutoRunToolsChange: (value: boolean) => void;
  selectedModel: string;
  onModelChange: (model: string) => void;
  availableModels: readonly { id: string; label: string }[];
  onAddContext: (ctx: ChatContext) => void;
  onRemoveContext: (ctx: ChatContext) => void;
  onSubmit: (e?: React.FormEvent) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  onStop: () => void;
}

function ChatInput({
  inputValue,
  setInputValue,
  isStreaming,
  textareaRef,
  selectedContexts,
  contextProviders,
  autoRunTools,
  onAutoRunToolsChange,
  selectedModel,
  onModelChange,
  availableModels,
  onAddContext,
  onRemoveContext,
  onSubmit,
  onKeyDown,
  onChange,
  onStop
}: ChatInputProps) {
  const { t } = useTranslation();
  const mentions = useChatTextareaMentions({
    inputValue,
    setInputValue,
    contextProviders,
    onAddContext,
    textareaRef,
    onChange,
    onKeyDown
  });
  const currentModelLabel =
    availableModels.find((m) => m.id === selectedModel)?.label ?? selectedModel;

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
                {ctx.meta?.Language && <span className="text-primary/60">{ctx.meta.Language}</span>}
                {ctx.meta?.Provider && <span className="text-primary/60">{ctx.meta.Provider}</span>}
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
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-muted border border-border/50 transition-colors"
              >
                <Zap className="size-3" />
                <span>{currentModelLabel}</span>
                <ChevronDown className="size-3 opacity-50" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-52">
              {availableModels.map((model) => (
                <DropdownMenuItem
                  key={model.id}
                  onClick={() => onModelChange(model.id)}
                  className={cn(
                    'flex items-center gap-2',
                    selectedModel === model.id && 'bg-primary/10'
                  )}
                >
                  <Check
                    className={cn(
                      'size-3.5 shrink-0',
                      selectedModel === model.id ? 'opacity-100 text-primary' : 'opacity-0'
                    )}
                  />
                  <span>{model.label}</span>
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
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
          <div className="min-w-0 flex-1 flex flex-col gap-0">
            <Textarea
              ref={textareaRef}
              value={inputValue}
              onChange={mentions.textareaHandlers.onChange}
              onKeyDown={mentions.textareaHandlers.onKeyDown}
              onSelect={mentions.textareaHandlers.onSelect}
              placeholder={t('ai.input.placeholder')}
              className="min-h-[64px] max-h-[220px] resize-none overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden bg-muted/50 border-border/50 focus-visible:ring-primary/30"
              disabled={isStreaming}
              rows={1}
            />
            {mentions.mentionList}
          </div>
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

interface QuestionFieldInputProps {
  field: AgentQuestionField;
  value: string;
  onChange: (value: string) => void;
}

function QuestionFieldInput({ field, value, onChange }: QuestionFieldInputProps) {
  switch (field.type) {
    case 'textarea':
      return (
        <Textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={field.placeholder}
          rows={3}
          className="resize-y text-sm"
        />
      );
    case 'select':
      return (
        <select
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          <option value="">{field.placeholder || 'Select...'}</option>
          {field.options?.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      );
    case 'toggle':
      return (
        <button
          type="button"
          onClick={() => onChange(value === 'true' ? 'false' : 'true')}
          className="flex items-center gap-2.5"
        >
          <div
            className={cn(
              'w-9 h-5 rounded-full relative transition-colors',
              value === 'true' ? 'bg-primary' : 'bg-muted-foreground/30'
            )}
          >
            <div
              className={cn(
                'absolute top-0.5 w-4 h-4 rounded-full bg-white shadow-sm transition-transform',
                value === 'true' ? 'translate-x-[18px]' : 'translate-x-0.5'
              )}
            />
          </div>
          <span className="text-sm text-foreground">{value === 'true' ? 'Yes' : 'No'}</span>
        </button>
      );
    default:
      return (
        <Input
          type={field.type === 'password' ? 'password' : 'text'}
          value={value}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
          placeholder={field.placeholder}
          className="text-sm"
        />
      );
  }
}

interface AgentQuestionModalProps {
  question: AgentQuestion;
  onSubmit: (answers: Record<string, string>) => void;
  onDismiss: () => void;
}

function AgentQuestionModal({ question, onSubmit, onDismiss }: AgentQuestionModalProps) {
  const [values, setValues] = React.useState<Record<string, string>>(() => {
    const initial: Record<string, string> = {};
    for (const field of question.fields) {
      initial[field.name] = field.defaultValue ?? (field.type === 'toggle' ? 'false' : '');
    }
    return initial;
  });
  const [errors, setErrors] = React.useState<Record<string, boolean>>({});

  const handleChange = React.useCallback((name: string, value: string) => {
    setValues((prev) => ({ ...prev, [name]: value }));
    setErrors((prev) => ({ ...prev, [name]: false }));
  }, []);

  const handleSubmit = React.useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      const newErrors: Record<string, boolean> = {};
      let hasErrors = false;
      for (const field of question.fields) {
        if (field.required && !values[field.name]?.trim()) {
          newErrors[field.name] = true;
          hasErrors = true;
        }
      }
      if (hasErrors) {
        setErrors(newErrors);
        return;
      }
      onSubmit(values);
    },
    [question.fields, values, onSubmit]
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onDismiss} />
      <div className="relative z-10 w-full max-w-md mx-4 rounded-xl overflow-hidden bg-background border border-border shadow-2xl">
        <div className="px-5 pt-5 pb-3">
          <h2 className="text-base font-semibold text-foreground">{question.title}</h2>
          {question.description && (
            <p className="mt-1 text-sm text-muted-foreground">{question.description}</p>
          )}
        </div>
        <form onSubmit={handleSubmit} className="px-5 pb-5">
          <div className="flex flex-col gap-4 max-h-[50vh] overflow-y-auto pr-1">
            {question.fields.map((field) => (
              <div key={field.name} className="flex flex-col gap-1.5">
                <label className="text-sm font-medium text-foreground">
                  {field.label}
                  {field.required && <span className="text-destructive ml-0.5">*</span>}
                </label>
                <QuestionFieldInput
                  field={field}
                  value={values[field.name] ?? ''}
                  onChange={(v) => handleChange(field.name, v)}
                />
                {errors[field.name] && <span className="text-xs text-destructive">Required</span>}
              </div>
            ))}
          </div>
          <div className="flex justify-end gap-2 mt-5 pt-4 border-t border-border/50">
            <Button type="button" variant="outline" size="sm" onClick={onDismiss}>
              Cancel
            </Button>
            <Button type="submit" size="sm">
              Submit
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

function Skeleton({ className }: { className?: string }) {
  return <div className={cn('animate-pulse rounded-md bg-muted', className)} />;
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
