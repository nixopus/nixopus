'use client';
import { ThemeProvider } from '@nixopus/ui';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '@/redux/store';
import { Toaster } from '@nixopus/ui';
import { useAppDispatch, useAppSelector } from '@/redux/hooks';
import { useEffect, useState, useMemo } from 'react';
import { initializeAuth } from '@/redux/features/users/authSlice';
import { preloadBaseUrl } from '@/redux/base-query';
import { usePathname, useRouter } from 'next/navigation';
import { WebSocketProvider } from '@/packages/hooks/shared/socket-provider';
import { FeatureFlagsProvider } from '@/packages/hooks/shared/features_provider';
import { SystemStatsProvider } from '@/packages/hooks/shared/system-stats-provider';
import { palette } from '@/packages/utils/colors';
import { authClient } from '@/packages/lib/auth-client';
import AppLayout from '@/packages/layouts/layout';
import { SudoModeProvider } from '@/packages/hooks/security/use-sudo-mode';
import { CSPostHogProvider } from '@/packages/components/analytics/posthog-provider';
import Image from 'next/image';
import { useTheme } from 'next-themes';
import { Skeleton } from '@nixopus/ui';
import { getPluginProviders } from '@/plugins/registry';

function PluginProviderWrapper({ children }: { children: React.ReactNode }) {
  const Providers = getPluginProviders();
  if (!Providers) return <>{children}</>;
  return <Providers>{children}</Providers>;
}

const AppLoadingScreen = () => {
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const logoSrc = mounted && resolvedTheme === 'dark' ? '/logo_white.png' : '/logo_black.png';

  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-background">
      <div className="app-loading-logo">
        <Image src={logoSrc} alt="Nixopus" width={48} height={48} priority />
      </div>
      <div className="mt-8 flex items-center gap-1.5">
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
};

const AppShellSkeleton = () => (
  <div className="flex h-screen w-screen bg-background">
    <div className="flex w-12 shrink-0 flex-col border-r border-border bg-sidebar">
      <div className="flex h-12 items-center justify-center border-b border-border">
        <Skeleton className="h-4 w-4 rounded" />
      </div>
      <div className="flex flex-1 flex-col gap-2 p-2">
        <Skeleton className="h-8 w-8 rounded-md" />
        <Skeleton className="h-8 w-8 rounded-md" />
        <Skeleton className="h-8 w-8 rounded-md" />
      </div>
    </div>
    <div className="flex flex-1 flex-col">
      <header className="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4">
        <Skeleton className="h-4 w-24" />
        <div className="ml-auto flex gap-2">
          <Skeleton className="h-8 w-8 rounded-md" />
          <Skeleton className="h-8 w-20 rounded-md" />
        </div>
      </header>
      <main className="flex-1 overflow-auto p-6">
        <div className="flex flex-col gap-6">
          <div className="flex items-center justify-between">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-9 w-32 rounded-md" />
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-32 rounded-lg" />
            ))}
          </div>
        </div>
      </main>
    </div>
  </div>
);

export default function ClientLayout({ children }: { children: React.ReactNode }) {
  return (
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        <CSPostHogProvider>
          <ChildrenWrapper>{children}</ChildrenWrapper>
        </CSPostHogProvider>
      </PersistGate>
    </Provider>
  );
}

const ChildrenWrapper = ({ children }: { children: React.ReactNode }) => {
  const dispatch = useAppDispatch();
  const pathname = usePathname();
  const router = useRouter();
  const authState = useAppSelector((state) => state.auth);
  const { isAuthenticated, isInitialized, user } = authState;
  const [isLoading, setIsLoading] = useState(true);

  const PUBLIC_ROUTES = [
    '/login',
    '/register',
    '/auth',
    '/reset-password',
    '/verify-email',
    '/auth/organization-invite'
  ];
  const isPublicRoute = useMemo(
    () => PUBLIC_ROUTES.some((route) => pathname === route || pathname.startsWith(route + '/')),
    [pathname]
  );

  const hasPersistedAuth = isAuthenticated && !!user;

  useEffect(() => {
    const initAuth = async () => {
      try {
        await preloadBaseUrl();
        await dispatch(initializeAuth() as any);
      } finally {
        setIsLoading(false);
      }
    };
    initAuth();
  }, [dispatch]);

  useEffect(() => {
    if (hasPersistedAuth && !isPublicRoute) {
      setIsLoading(false);
    }
  }, [hasPersistedAuth, isPublicRoute]);

  useEffect(() => {
    if ((hasPersistedAuth || isAuthenticated) && (pathname === '/' || pathname === '/auth')) {
      router.prefetch('/chats');
    }
  }, [hasPersistedAuth, isAuthenticated, pathname, router]);

  // Re-check session only on auth/login routes to catch post-login state before Redux updates
  const isAuthCheckRoute =
    pathname === '/auth' || pathname === '/login' || pathname === '/register';
  useEffect(() => {
    if (!isInitialized || !isAuthCheckRoute) return;

    const checkSession = async () => {
      try {
        const session = await authClient.getSession();
        const hasSession = !!session?.data?.session;
        if (hasSession && !isAuthenticated) {
          await dispatch(initializeAuth() as any);
        }
      } catch {
        // Session check failed, rely on Redux state
      }
    };

    checkSession();
  }, [pathname, isAuthenticated, isInitialized, isAuthCheckRoute, dispatch]);

  useEffect(() => {
    if (isLoading || !isInitialized) return;

    // Prevent redirect loops by checking if we're already on the target route
    if (!isPublicRoute && !isAuthenticated) {
      if (pathname !== '/auth') {
        router.push('/auth');
      }
    } else if (isPublicRoute && isAuthenticated) {
      if (
        pathname === '/' ||
        pathname === '/auth' ||
        pathname === '/login' ||
        pathname === '/register'
      ) {
        router.push('/chats');
      }
    }
  }, [pathname, isLoading, isInitialized, router, isPublicRoute, isAuthenticated]);

  return (
    <>
      <ThemeProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        disableTransitionOnChange
        themes={palette}
      >
        {isLoading ? (
          isPublicRoute ? (
            <AppLoadingScreen />
          ) : (
            <AppShellSkeleton />
          )
        ) : (
          <WebSocketProvider>
            <FeatureFlagsProvider>
              {isPublicRoute ? (
                <>{children}</>
              ) : (
                <SystemStatsProvider>
                  <SudoModeProvider>
                    <AppLayout>
                      <PluginProviderWrapper>{children}</PluginProviderWrapper>
                    </AppLayout>
                  </SudoModeProvider>
                </SystemStatsProvider>
              )}
            </FeatureFlagsProvider>
          </WebSocketProvider>
        )}
      </ThemeProvider>
      <Toaster />
    </>
  );
};
