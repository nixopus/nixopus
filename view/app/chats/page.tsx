'use client';

import React from 'react';
import dynamic from 'next/dynamic';

const ChatPage = dynamic(
  () => import('@/packages/components/ai-chat').then((m) => ({ default: m.ChatPage })),
  { ssr: false }
);

export default function ChatsPage() {
  return (
    <div className="h-full w-full">
      <ChatPage />
    </div>
  );
}
