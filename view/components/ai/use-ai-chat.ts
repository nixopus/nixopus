'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import type { Message } from './types';

interface UseAIChatOptions {
  open: boolean;
}

export function useAIChat({ open }: UseAIChatOptions) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const scrollToBottom = useCallback(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  useEffect(() => {
    if (open && textareaRef.current) {
      setTimeout(() => textareaRef.current?.focus(), 100);
    }
  }, [open]);

  const handleSubmit = useCallback(
    (e?: React.FormEvent) => {
      e?.preventDefault();
      if (!inputValue.trim() || isStreaming) return;

      const userMessage: Message = {
        id: crypto.randomUUID(),
        role: 'user',
        content: inputValue.trim(),
        timestamp: new Date()
      };

      setMessages((prev) => [...prev, userMessage]);
      setInputValue('');
      setIsStreaming(true);

      // Placeholder for SSE integration
      // This will be replaced with actual API call
      setTimeout(() => {
        const assistantMessage: Message = {
          id: crypto.randomUUID(),
          role: 'assistant',
          content:
            'This is a placeholder response. SSE integration will provide real streaming responses from the AI backend.',
          timestamp: new Date()
        };
        setMessages((prev) => [...prev, assistantMessage]);
        setIsStreaming(false);
      }, 1000);
    },
    [inputValue, isStreaming]
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
    scrollRef,
    textareaRef,
    handleSubmit,
    handleKeyDown,
    handleSuggestionClick,
    handleInputChange
  };
}
