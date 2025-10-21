'use client';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { Loader2 } from 'lucide-react';

export default function Home() {
  const router = useRouter();
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const isInitialized = useAppSelector((state) => state.auth.isInitialized);

  useEffect(() => {
    if (isInitialized && isAuthenticated) {
      router.push('/dashboard');
    } else {
      router.push('/auth');
    }
  }, [isAuthenticated, isInitialized, router]);

  return (
    <div className="flex h-screen flex-col items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin" />
    </div>
  );
}
