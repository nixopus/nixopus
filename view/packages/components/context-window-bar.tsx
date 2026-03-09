'use client';

import React, { useState } from 'react';
import { ChevronDown, ChevronUp, Brain, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { type OmStatus } from '@/packages/hooks/ai/use-agent-chat';

function formatTokens(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

function TokenBar({
  label,
  tokens,
  threshold,
  color
}: {
  label: string;
  tokens: number;
  threshold: number;
  color: 'primary' | 'amber';
}) {
  const pct = threshold > 0 ? Math.min((tokens / threshold) * 100, 100) : 0;
  const barColor =
    color === 'primary'
      ? pct > 80
        ? 'bg-orange-500'
        : 'bg-primary'
      : pct > 80
        ? 'bg-orange-500'
        : 'bg-amber-500';

  return (
    <div className="flex items-center gap-2 min-w-0">
      <span className="text-[10px] font-medium text-muted-foreground whitespace-nowrap uppercase tracking-wider">
        {label}
      </span>
      <div className="flex-1 h-1.5 rounded-full bg-muted/50 min-w-[60px] max-w-[120px] overflow-hidden">
        <div
          className={cn('h-full rounded-full transition-all duration-500', barColor)}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-[10px] tabular-nums text-muted-foreground whitespace-nowrap">
        {formatTokens(tokens)}/{formatTokens(threshold)}
      </span>
    </div>
  );
}

interface ContextWindowBarProps {
  omStatus: OmStatus;
}

export function ContextWindowBar({ omStatus }: ContextWindowBarProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="shrink-0 border-t border-border/30">
      <div className="max-w-3xl mx-auto px-4">
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="w-full flex items-center gap-3 py-1.5 group"
        >
          <Brain className="size-3.5 text-muted-foreground/70" />

          <div className="flex items-center gap-4 flex-1 min-w-0">
            <TokenBar
              label="Context"
              tokens={omStatus.messages.tokens}
              threshold={omStatus.messages.threshold}
              color="primary"
            />
            <TokenBar
              label="Memory"
              tokens={omStatus.observations.tokens}
              threshold={omStatus.observations.threshold}
              color="amber"
            />
          </div>

          {omStatus.isObserving && (
            <span className="flex items-center gap-1 text-[10px] text-amber-500 font-medium animate-pulse">
              <Loader2 className="size-3 animate-spin" />
              Summarizing
            </span>
          )}

          {omStatus.observationsText ? (
            expanded ? (
              <ChevronDown className="size-3 text-muted-foreground/50 group-hover:text-muted-foreground transition-colors" />
            ) : (
              <ChevronUp className="size-3 text-muted-foreground/50 group-hover:text-muted-foreground transition-colors" />
            )
          ) : null}
        </button>

        {expanded && omStatus.observationsText && (
          <div className="pb-2">
            <div className="rounded-md border border-border/50 bg-muted/30 p-3 max-h-48 overflow-y-auto">
              <p className="text-[10px] uppercase tracking-wider text-muted-foreground/70 font-medium mb-1.5">
                Agent Memory
              </p>
              <pre className="text-xs text-muted-foreground whitespace-pre-wrap font-mono leading-relaxed">
                {omStatus.observationsText}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
