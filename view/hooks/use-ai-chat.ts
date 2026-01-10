'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { streamAIChat, getThreadMessages } from '@/redux/services/agents/agentsApi';
import type { Message } from '../components/ai/types';

interface UseAIChatOptions {
  open: boolean;
  threadId?: string | null;
  onThreadChange?: (threadId: string) => void;
}

import type { ThreadMessage } from '@/redux/services/agents/agentsApi';

function convertApiMessageToMessage(apiMsg: ThreadMessage): Message {
  let textContent = '';

  if (Array.isArray(apiMsg.content)) {
    textContent = apiMsg.content
      .filter((c) => c.type === 'text' && c.text)
      .map((c) => c.text)
      .join('');
  } else if (typeof apiMsg.content === 'string') {
    textContent = apiMsg.content;
  }

  return {
    id: apiMsg.id,
    role: apiMsg.role as 'user' | 'assistant',
    content: textContent,
    timestamp: new Date(apiMsg.createdAt || Date.now())
  };
}

function createUserMessage(content: string): Message {
  return {
    id: crypto.randomUUID(),
    role: 'user',
    content: content.trim(),
    timestamp: new Date()
  };
}

function createAssistantMessage(id: string): Message {
  return {
    id,
    role: 'assistant',
    content: '',
    blocks: [],
    timestamp: new Date()
  };
}

function updateMessageContent(messages: Message[], messageId: string, content: string): Message[] {
  return messages.map((msg) => {
    if (msg.id !== messageId) return msg;

    const newContent = msg.content + content;
    const blocks = [...(msg.blocks || [])];
    const lastBlock = blocks[blocks.length - 1];

    if (lastBlock && lastBlock.type === 'text') {
      blocks[blocks.length - 1] = {
        type: 'text',
        content: lastBlock.content + content
      };
    } else {
      blocks.push({ type: 'text', content });
    }

    return { ...msg, content: newContent, blocks };
  });
}

function addToolCallToMessage(
  messages: Message[],
  messageId: string,
  toolName: string,
  toolCallId: string,
  args: Record<string, any>
): Message[] {
  return messages.map((msg) => {
    if (msg.id !== messageId) return msg;

    const blocks = [...(msg.blocks || [])];
    blocks.push({
      type: 'tool-call',
      toolName,
      toolCallId,
      arguments: args
    });

    return { ...msg, blocks };
  });
}

function addToolResultToMessage(
  messages: Message[],
  messageId: string,
  toolCallId: string,
  result: any,
  isError?: boolean
): Message[] {
  return messages.map((msg) => {
    if (msg.id !== messageId) return msg;

    const blocks = [...(msg.blocks || [])];
    blocks.push({
      type: 'tool-result',
      toolCallId,
      result,
      isError
    });

    return { ...msg, blocks };
  });
}

function updateMessageOnError(messages: Message[], messageId: string): Message[] {
  return messages.map((msg) => {
    if (msg.id !== messageId) return msg;

    return {
      ...msg,
      content:
        msg.content || 'Sorry, an error occurred while processing your request. Please try again.'
    };
  });
}

function findScrollViewport(scrollRef: React.RefObject<HTMLDivElement | null>): HTMLElement | null {
  if (!scrollRef.current) return null;
  return scrollRef.current.querySelector('[data-radix-scroll-area-viewport]') as HTMLElement;
}

function shouldAutoScroll(
  viewport: HTMLElement,
  messagesCount: number,
  isStreaming: boolean
): boolean {
  const isNearBottom = viewport.scrollHeight - viewport.scrollTop - viewport.clientHeight < 100;
  return isNearBottom || messagesCount <= 1 || isStreaming;
}

export function useAIChat({ open, threadId, onThreadChange }: UseAIChatOptions) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);

  const scrollRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const abortControllerRef = useRef<AbortController | null>(null);
  const currentThreadIdRef = useRef<string | null>(threadId || null);
  const lastMessageContentRef = useRef<string>('');

  const scrollToBottom = useCallback(() => {
    const viewport = findScrollViewport(scrollRef);
    if (!viewport) return;

    if (shouldAutoScroll(viewport, messages.length, isStreaming)) {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          viewport.scrollTop = viewport.scrollHeight;
        });
      });
    }
  }, [messages.length, isStreaming]);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  useEffect(() => {
    if (isStreaming) {
      scrollToBottom();
    }
  }, [isStreaming, scrollToBottom]);

  useEffect(() => {
    if (messages.length === 0) return;

    const lastMessage = messages[messages.length - 1];
    const currentContent = lastMessage.content || '';

    if (isStreaming && currentContent !== lastMessageContentRef.current) {
      lastMessageContentRef.current = currentContent;
      const timeoutId = setTimeout(() => {
        scrollToBottom();
      }, 10);
      return () => clearTimeout(timeoutId);
    }

    if (!isStreaming) {
      lastMessageContentRef.current = currentContent;
    }
  }, [messages, isStreaming, scrollToBottom]);

  useEffect(() => {
    if (open && textareaRef.current) {
      setTimeout(() => textareaRef.current?.focus(), 100);
    }
  }, [open]);

  const loadThreadMessages = useCallback(async (threadIdToLoad: string) => {
    setIsLoadingMessages(true);
    try {
      const threadMessages = await getThreadMessages(threadIdToLoad);

      if (!Array.isArray(threadMessages)) {
        console.warn('Thread messages is not an array:', threadMessages);
        setMessages([]);
        return;
      }

      const convertedMessages = threadMessages.map(convertApiMessageToMessage);
      setMessages(convertedMessages);
    } catch (error) {
      console.error('Error loading thread messages:', error);
      setMessages([]);
    } finally {
      setIsLoadingMessages(false);
    }
  }, []);

  useEffect(() => {
    if (threadId === currentThreadIdRef.current) return;

    currentThreadIdRef.current = threadId || null;

    if (threadId) {
      loadThreadMessages(threadId);
    } else {
      setMessages([]);
    }
  }, [threadId, loadThreadMessages]);

  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  const createStreamingHandlers = useCallback((assistantMessageId: string) => {
    return {
      onContent: (content: string) => {
        setMessages((prev) => updateMessageContent(prev, assistantMessageId, content));
      },
      onToolCall: (toolName: string, toolCallId: string, args: Record<string, any>) => {
        setMessages((prev) =>
          addToolCallToMessage(prev, assistantMessageId, toolName, toolCallId, args)
        );
      },
      onToolResult: (toolCallId: string, result: any, isError?: boolean) => {
        setMessages((prev) =>
          addToolResultToMessage(prev, assistantMessageId, toolCallId, result, isError)
        );
      },
      onDone: () => {
        setIsStreaming(false);
      },
      onError: (error: unknown) => {
        console.error('Error streaming AI response:', error);
        setMessages((prev) => updateMessageOnError(prev, assistantMessageId));
        setIsStreaming(false);
      }
    };
  }, []);

  const prepareStreamingRequest = useCallback(
    (messageToSend: string, assistantMessageId: string) => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }

      const abortController = new AbortController();
      abortControllerRef.current = abortController;

      const history = messages.map((msg) => ({
        role: msg.role,
        content: msg.content
      }));

      const activeThreadId = threadId || currentThreadIdRef.current || crypto.randomUUID();

      if (!currentThreadIdRef.current && onThreadChange) {
        currentThreadIdRef.current = activeThreadId;
        onThreadChange(activeThreadId);
      }

      return {
        request: {
          message: messageToSend,
          history,
          threadId: activeThreadId
        },
        callbacks: createStreamingHandlers(assistantMessageId),
        signal: abortController.signal
      };
    },
    [messages, threadId, onThreadChange, createStreamingHandlers]
  );

  const handleSubmit = useCallback(
    async (e?: React.FormEvent) => {
      e?.preventDefault();
      if (!inputValue.trim() || isStreaming) return;

      const messageToSend = inputValue.trim();
      const userMessage = createUserMessage(messageToSend);
      const assistantMessageId = crypto.randomUUID();
      const assistantMessage = createAssistantMessage(assistantMessageId);

      setMessages((prev) => [...prev, userMessage, assistantMessage]);
      setInputValue('');
      setIsStreaming(true);

      const { request, callbacks, signal } = prepareStreamingRequest(
        messageToSend,
        assistantMessageId
      );

      await streamAIChat(request, callbacks, signal);
    },
    [inputValue, isStreaming, prepareStreamingRequest]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSubmit();
      }
    },
    [handleSubmit]
  );

  const handleSuggestionClick = useCallback((text: string) => {
    setInputValue(text);
    textareaRef.current?.focus();
  }, []);

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputValue(e.target.value);
  }, []);

  return {
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
  };
}
