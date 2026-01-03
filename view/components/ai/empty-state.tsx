'use client';

import React from 'react';
import { Bot } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { SuggestionChip } from './suggestion-chip';

interface EmptyStateProps {
  onSuggestionClick: (text: string) => void;
}

export function EmptyState({ onSuggestionClick }: EmptyStateProps) {
  const { t } = useTranslation();

  const suggestions = [
    t('ai.suggestions.deploy'),
    t('ai.suggestions.logs'),
    t('ai.suggestions.envVars')
  ];

  return (
    <div className="flex flex-col items-center justify-center py-16 px-4">
      <div className="flex items-center justify-center size-16 rounded-2xl bg-primary/10 mb-6">
        <Bot className="size-8 text-primary" />
      </div>
      <h3 className="text-lg font-semibold text-foreground mb-2">{t('ai.emptyState.title')}</h3>
      <p className="text-sm text-muted-foreground text-center max-w-sm">
        {t('ai.emptyState.description')}
      </p>
      <div className="grid grid-cols-1 gap-2 mt-8 w-full max-w-sm">
        {suggestions.map((suggestion, index) => (
          <SuggestionChip
            key={index}
            text={suggestion}
            onClick={() => onSuggestionClick(suggestion)}
          />
        ))}
      </div>
    </div>
  );
}
