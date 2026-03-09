'use client';

import React, { Suspense } from 'react';
import { ChatPage } from '@/packages/components/ai-chat';

export default function ChatsPage() {
  return (
    <div className="h-[calc(100vh-4rem)] w-full">
      <Suspense>
        <ChatPage />
      </Suspense>
    </div>
  );
}
