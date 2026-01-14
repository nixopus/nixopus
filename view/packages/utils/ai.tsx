import React from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';

export function formatTime(date: Date): string {
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function formatResult(result: any, isError: boolean): string {
  if (typeof result === 'string') {
    return result;
  }
  if (isError) {
    return JSON.stringify(result);
  }
  return JSON.stringify(result, null, 2);
}

export function getChevronIcon(hasResult: boolean, isExpanded: boolean): React.ReactNode {
  if (!hasResult) {
    return '›';
  }
  if (isExpanded) {
    return <ChevronDown className="size-3" />;
  }
  return <ChevronRight className="size-3" />;
}
