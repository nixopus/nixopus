'use client';

import React from 'react';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Bot } from 'lucide-react';

export function StreamingIndicator() {
  return (
    <div className="flex gap-3">
      <Avatar className="size-8 shrink-0">
        <AvatarFallback className="bg-muted text-muted-foreground">
          <Bot className="size-4" />
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
