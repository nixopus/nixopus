'use client';

import React, { useState, useCallback } from 'react';
import { AIContent } from '@/packages/components/ai-sheet';
import { ThreadsSidebar } from '@/packages/components/ai-sheet';

export default function ChatPage() {
  const [selectedThreadId, setSelectedThreadId] = useState<string | null>(null);

  const handleThreadSelect = useCallback((threadId: string | null) => {
    setSelectedThreadId(threadId);
  }, []);

  const handleNewThread = useCallback(() => {
    setSelectedThreadId(null);
  }, []);

  const handleThreadChange = useCallback((threadId: string) => {
    setSelectedThreadId(threadId);
  }, []);

  return (
    <div className="flex h-[calc(100vh-4rem)] overflow-hidden bg-background">
      <aside className="w-64 shrink-0 hidden md:flex flex-col h-full overflow-hidden">
        <ThreadsSidebar
          selectedThreadId={selectedThreadId}
          onThreadSelect={handleThreadSelect}
          onNewThread={handleNewThread}
          className="h-full"
        />
      </aside>
      <main className="flex-1 flex flex-col h-full overflow-hidden min-w-0">
        <AIContent
          open={true}
          className="h-full"
          threadId={selectedThreadId}
          onThreadChange={handleThreadChange}
        />
      </main>
    </div>
  );
}
