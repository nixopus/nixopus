'use client';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useAppSelector } from '@/redux/hooks';

export default function Home() {
  const router = useRouter();
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const isInitialized = useAppSelector((state) => state.auth.isInitialized);

  useEffect(() => {
    if (isInitialized && isAuthenticated) {
      router.push('/apps');
    } else {
      router.push('/auth');
    }
  }, [isAuthenticated, isInitialized, router]);

  return (
    <div className="flex h-screen flex-col items-center justify-center bg-background">
      <div className="flex items-center gap-1.5">
        <div
          className="app-loading-dot h-1.5 w-1.5 rounded-full bg-primary/60"
          style={{ animationDelay: '0ms' }}
        />
        <div
          className="app-loading-dot h-1.5 w-1.5 rounded-full bg-primary/60"
          style={{ animationDelay: '150ms' }}
        />
        <div
          className="app-loading-dot h-1.5 w-1.5 rounded-full bg-primary/60"
          style={{ animationDelay: '300ms' }}
        />
      </div>
    </div>
  );
}
