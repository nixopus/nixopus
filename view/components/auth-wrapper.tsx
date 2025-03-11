import { useAppSelector } from '@/redux/hooks';
import { usePathname, useRouter } from 'next/navigation';
import React, { useEffect } from 'react';
import DashboardLayout from './dashboard-layout';
import { WebSocketProvider } from '@/hooks/socket_provider';

function AuthWrapper({ children }: { children: React.ReactNode }) {
  const user = useAppSelector((state) => state.auth.user);
  const authenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (pathname === '/login' || pathname === '/') {
      return;
    }

    if (!authenticated || !user) {
      console.log('not authenticated');
      router.push('/login');
    }
  }, [user, authenticated, router, pathname]);

  useEffect(() => {
    if (authenticated && user && pathname === '/login') {
      router.push('/dashboard');
    }
  }, [authenticated, user, pathname, router]);

  if (!authenticated || !user) {
    return null;
  }

  return (
    <WebSocketProvider>
      <DashboardLayout>{children}</DashboardLayout>
    </WebSocketProvider>
  );
}

export default AuthWrapper;
