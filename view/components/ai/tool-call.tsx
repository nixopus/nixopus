'use client';

import React, { useState } from 'react';
import { Grid3x3, ChevronDown, ChevronRight } from 'lucide-react';

interface ToolCallProps {
  toolName: string;
  toolCallId: string;
  arguments?: Record<string, any>;
  result?: any;
  isError?: boolean;
  isComplete?: boolean;
}

function getChevronIcon(hasResult: boolean, isExpanded: boolean) {
  if (!hasResult) {
    return '›';
  }
  if (isExpanded) {
    return <ChevronDown className="size-3" />;
  }
  return <ChevronRight className="size-3" />;
}

function formatResult(result: any, isError: boolean): string {
  if (typeof result === 'string') {
    return result;
  }
  if (isError) {
    return JSON.stringify(result);
  }
  return JSON.stringify(result, null, 2);
}

function ToolCallResult({ result, isError }: { result: any; isError?: boolean }) {
  const hasError = isError ?? false;
  const formattedResult = formatResult(result, hasError);

  if (hasError) {
    return <div className="text-xs text-destructive/80 font-mono">Error: {formattedResult}</div>;
  }

  return (
    <pre className="text-xs text-muted-foreground font-mono overflow-x-auto whitespace-pre-wrap">
      {formattedResult}
    </pre>
  );
}

export function ToolCall({ toolName, result, isError, isComplete }: ToolCallProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasResult = Boolean(result !== undefined && isComplete);

  const handleClick = () => {
    if (hasResult) {
      setIsExpanded(!isExpanded);
    }
  };

  return (
    <div className="my-1">
      <div
        className="flex items-center gap-2 cursor-pointer hover:opacity-80 transition-opacity"
        onClick={handleClick}
      >
        <span className="text-muted-foreground/60 text-sm">
          {getChevronIcon(hasResult, isExpanded)}
        </span>
        <div className="flex items-center gap-2 px-2 py-1 rounded-md bg-muted/40">
          <Grid3x3 className="size-3.5 text-amber-500 shrink-0" />
          <span className="text-xs font-mono text-foreground">{toolName}</span>
        </div>
      </div>

      {isExpanded && hasResult && (
        <div className="ml-6 mt-1 px-2 py-1.5 rounded-md bg-muted/20 border border-border/30">
          <ToolCallResult result={result} isError={isError} />
        </div>
      )}
    </div>
  );
}
