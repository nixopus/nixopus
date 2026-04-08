'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useAgentChat,
  type ChatMessage,
  type PendingToolApproval,
  type OmStatus,
  type TokenUsage,
  type AgentQuestion
} from './use-agent-chat';
import { useChatThreads, type ChatThread } from './use-chat-threads';
import {
  type ChatContext,
  type ContextProviderData,
  useChatContextProviders
} from './chat-context';
import { useMemorySearch, type MemorySearchResult } from './use-memory-search';
import { getSelfHosted } from '@/redux/conf';

function useLocalStorageState(key: string, defaultValue: boolean) {
  const [value, setValue] = useState(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem(key) === 'true';
    }
    return defaultValue;
  });

  useEffect(() => {
    localStorage.setItem(key, String(value));
  }, [key, value]);

  return [value, setValue] as const;
}

export const AVAILABLE_MODELS = [
  { id: 'openrouter/anthropic/claude-sonnet-4.5', label: 'Claude Sonnet 4.5' },
  { id: 'openrouter/anthropic/claude-sonnet-4', label: 'Claude Sonnet 4' },
  { id: 'openrouter/anthropic/claude-3.7-sonnet', label: 'Claude 3.7 Sonnet' },
  { id: 'openrouter/openai/gpt-4.1', label: 'GPT-4.1' },
  { id: 'openrouter/openai/gpt-4.1-mini', label: 'GPT-4.1 Mini' },
  { id: 'openrouter/openai/gpt-4o-mini', label: 'GPT-4o Mini' },
  { id: 'openrouter/google/gemini-2.5-flash', label: 'Gemini 2.5 Flash' }
] as const;

export type ModelId = (typeof AVAILABLE_MODELS)[number]['id'];

export interface UseChatPageReturn {
  sidebarCollapsed: boolean;
  toggleSidebarCollapse: () => void;
  selectedContexts: ChatContext[];
  addContext: (ctx: ChatContext) => void;
  removeContext: (ctx: ChatContext) => void;
  autoRunTools: boolean;
  setAutoRunTools: (value: boolean) => void;
  selectedModel: string;
  setSelectedModel: (model: string) => void;
  isSelfHosted: boolean;
  contextProviders: ContextProviderData[];
  handleNewChat: () => void;

  threads: ChatThread[];
  activeThreadId: string | null;
  resourceId?: string;
  isThreadsInitialized: boolean;
  setActiveThreadId: (id: string) => void;
  deleteThread: (id: string) => void;
  renameThread: (id: string, title: string) => void;

  messages: ChatMessage[];
  inputValue: string;
  isStreaming: boolean;
  isLoadingHistory: boolean;
  scrollRef: React.RefObject<HTMLDivElement | null>;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  pendingToolApproval: PendingToolApproval | null;
  activeQuestion: AgentQuestion | null;
  omStatus: OmStatus | null;
  handleSubmit: (e?: React.FormEvent) => void;
  handleKeyDown: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
  handleSuggestionClick: (text: string) => void;
  handleInputChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  handleApproveToolCall: () => void;
  handleDeclineToolCall: () => void;
  submitQuestionResponse: (answers: Record<string, string>) => void;
  dismissQuestion: () => void;
  stopStreaming: () => void;
  setInputValue: (value: string) => void;
  readOnly: boolean;
  refreshThreads: () => void;
  isRefreshing: boolean;
}

export function useChatPage(): UseChatPageReturn {
  const { t } = useTranslation();
  const searchParams = useSearchParams();
  const navRouter = useRouter();

  const [sidebarCollapsed, setSidebarCollapsed] = useLocalStorageState(
    'chat_sidebar_collapsed',
    false
  );
  const [selectedContexts, setSelectedContexts] = useState<ChatContext[]>([]);
  const [autoRunTools, setAutoRunTools] = useLocalStorageState('chat_auto_run_tools', true);
  const [selectedModel, setSelectedModel] = useState<string>(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('chat_selected_model') || AVAILABLE_MODELS[0].id;
    }
    return AVAILABLE_MODELS[0].id;
  });
  const [pendingDeployPrompt, setPendingDeployPrompt] = useState<string | null>(null);
  const repoParamsHandledRef = useRef(false);
  const [isSelfHosted, setIsSelfHosted] = useState(false);

  useEffect(() => {
    getSelfHosted().then(setIsSelfHosted);
  }, []);

  useEffect(() => {
    localStorage.setItem('chat_selected_model', selectedModel);
  }, [selectedModel]);

  const contextProviders = useChatContextProviders();
  const threads = useChatThreads();
  const activeThread = threads.activeThread;
  const chat = useAgentChat({
    threadId: threads.activeThreadId,
    resourceId: activeThread?.threadResourceId || threads.resourceId,
    agentId: activeThread?.agentId,
    readOnly: activeThread?.isIncident ?? false,
    contexts: selectedContexts,
    autoRunTools,
    model: selectedModel,
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
    if (description) meta['Description'] = description;
    if (htmlUrl) meta['GitHub URL'] = htmlUrl;

    setSelectedContexts([
      {
        type: 'Repository',
        id: repoId,
        label: repoFullName,
        meta
      }
    ]);

    setPendingDeployPrompt(`Deploy "${repoFullName}" as a new application.`);
    navRouter.replace('/chats');
  }, [threads.isInitialized, searchParams]);

  useEffect(() => {
    if (pendingDeployPrompt && threads.activeThreadId) {
      chat.setInputValue(pendingDeployPrompt);
      setPendingDeployPrompt(null);
      setTimeout(() => chat.textareaRef.current?.focus(), 100);
    }
  }, [pendingDeployPrompt, threads.activeThreadId]);

  const toggleSidebarCollapse = useCallback(() => {
    setSidebarCollapsed((prev) => !prev);
  }, [setSidebarCollapsed]);

  const handleNewChat = useCallback(() => {
    threads.createThread(t('ai.threads.untitledChat'));
  }, [threads, t]);

  const addContext = useCallback((ctx: ChatContext) => {
    setSelectedContexts((prev) => {
      if (prev.some((c) => c.type === ctx.type && c.id === ctx.id)) return prev;
      return [...prev, ctx];
    });
  }, []);

  const removeContext = useCallback((ctx: ChatContext) => {
    setSelectedContexts((prev) => prev.filter((c) => !(c.type === ctx.type && c.id === ctx.id)));
  }, []);

  return {
    sidebarCollapsed,
    toggleSidebarCollapse,
    selectedContexts,
    addContext,
    removeContext,
    autoRunTools,
    setAutoRunTools,
    selectedModel,
    setSelectedModel,
    isSelfHosted,
    contextProviders,
    handleNewChat,

    threads: threads.threads,
    activeThreadId: threads.activeThreadId,
    resourceId: threads.resourceId,
    isThreadsInitialized: threads.isInitialized,
    setActiveThreadId: threads.setActiveThreadId,
    deleteThread: threads.deleteThread,
    renameThread: threads.updateThreadTitle,

    messages: chat.messages,
    inputValue: chat.inputValue,
    isStreaming: chat.isStreaming,
    isLoadingHistory: chat.isLoadingHistory,
    scrollRef: chat.scrollRef,
    textareaRef: chat.textareaRef,
    pendingToolApproval: chat.pendingToolApproval ?? null,
    activeQuestion: chat.activeQuestion ?? null,
    omStatus: chat.omStatus ?? null,
    handleSubmit: chat.handleSubmit,
    handleKeyDown: chat.handleKeyDown,
    handleSuggestionClick: chat.handleSuggestionClick,
    handleInputChange: chat.handleInputChange,
    handleApproveToolCall: chat.handleApproveToolCall,
    handleDeclineToolCall: chat.handleDeclineToolCall,
    submitQuestionResponse: chat.submitQuestionResponse,
    dismissQuestion: chat.dismissQuestion,
    stopStreaming: chat.stopStreaming,
    setInputValue: chat.setInputValue,
    readOnly: chat.readOnly,
    refreshThreads: threads.refreshThreads,
    isRefreshing: threads.isRefreshing
  };
}

export interface UseThreadSidebarSearchReturn {
  searchInputValue: string;
  setSearchInputValue: (value: string) => void;
  memorySearchResults: MemorySearchResult[];
  isSearching: boolean;
  handleSearchInputChange: (value: string) => void;
  handleSearchKeyDown: (key: string) => void;
  handleSelectSearchResult: (
    threadId: string | undefined,
    onSelectThread: (id: string) => void
  ) => void;
}

export function useThreadSidebarSearch(resourceId?: string): UseThreadSidebarSearchReturn {
  const [searchInputValue, setSearchInputValue] = useState('');
  const memorySearch = useMemorySearch(resourceId);

  const handleSearchInputChange = useCallback(
    (value: string) => {
      setSearchInputValue(value);
      if (!value.trim()) memorySearch.clear();
    },
    [memorySearch]
  );

  const handleSearchKeyDown = useCallback(
    (key: string) => {
      if (key === 'Enter') {
        memorySearch.search(searchInputValue);
      }
      if (key === 'Escape') {
        setSearchInputValue('');
        memorySearch.clear();
      }
    },
    [searchInputValue, memorySearch]
  );

  const handleSelectSearchResult = useCallback(
    (threadId: string | undefined, onSelectThread: (id: string) => void) => {
      if (threadId) {
        onSelectThread(threadId);
        setSearchInputValue('');
        memorySearch.clear();
      }
    },
    [memorySearch]
  );

  return {
    searchInputValue,
    setSearchInputValue,
    memorySearchResults: memorySearch.results,
    isSearching: memorySearch.isSearching,
    handleSearchInputChange,
    handleSearchKeyDown,
    handleSelectSearchResult
  };
}

export interface UseChatMessagesScrollReturn {
  containerRef: React.RefObject<HTMLDivElement | null>;
}

export function useChatMessagesScroll(messages: ChatMessage[]): UseChatMessagesScrollReturn {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [messages]);

  return { containerRef };
}

export function useContextSearch(items: ChatContext[]) {
  const [search, setSearch] = useState('');

  const filtered = search.trim()
    ? items.filter((item) => item.label.toLowerCase().includes(search.toLowerCase()))
    : items;

  const resetSearch = useCallback(() => setSearch(''), []);

  return { search, setSearch, filtered, resetSearch };
}

export function formatTime(date: Date): string {
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
