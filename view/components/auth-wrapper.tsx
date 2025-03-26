import { useAppSelector } from '@/redux/hooks';
import { usePathname, useRouter } from 'next/navigation';
import React, { useEffect, useState } from 'react';
import DashboardLayout from './dashboard-layout';
import { WebSocketProvider } from '@/hooks/socket_provider';
import Loading from './ui/loading';

const PUBLIC_ROUTES = ['/login', '/register', '/forgot-password', '/'];

function AuthWrapper({ children }: { children: React.ReactNode }) {
  const user = useAppSelector((state) => state.auth.user);
  const authenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const isInitialized = useAppSelector((state) => state.auth.isInitialized);
  const router = useRouter();
  const pathname = usePathname();

  const [isLoading, setIsLoading] = useState(true);

  const isPublicRoute = PUBLIC_ROUTES.includes(pathname);

  useEffect(() => {
    if (!isInitialized) {
      return;
    }

    if (!authenticated && !isPublicRoute) {
      router.push('/login');
    } else if (authenticated && pathname === '/login') {
      router.push('/dashboard');
    }

    setIsLoading(false);
  }, [authenticated, isInitialized, isPublicRoute, pathname, router]);

  if (!isInitialized || isLoading) {
    return <Loading />;
  }

  if (isPublicRoute) {
    return <>{children}</>;
  }

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
