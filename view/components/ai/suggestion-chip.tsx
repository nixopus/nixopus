'use client';

import React from 'react';

interface SuggestionChipProps {
  text: string;
  onClick: () => void;
}

export function SuggestionChip({ text, onClick }: SuggestionChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="px-4 py-3 text-sm text-left rounded-lg border border-border/50 bg-muted/30 hover:bg-muted/60 hover:border-border transition-colors text-muted-foreground hover:text-foreground"
    >
      {text}
    </button>
  );
}
