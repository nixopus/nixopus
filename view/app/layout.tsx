'use client';
import { Geist, Geist_Mono } from 'next/font/google';
import './globals.css';
import { ThemeProvider } from '@nixopus/ui';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '@/redux/store';
import { Toaster } from '@nixopus/ui';
import { useAppDispatch, useAppSelector } from '@/redux/hooks';
import { useEffect, useState, useMemo } from 'react';
import { initializeAuth } from '@/redux/features/users/authSlice';
import { usePathname, useRouter } from 'next/navigation';
import { WebSocketProvider } from '@/packages/hooks/shared/socket-provider';
import { FeatureFlagsProvider } from '@/packages/hooks/shared/features_provider';
import { SystemStatsProvider } from '@/packages/hooks/shared/system-stats-provider';
import { palette } from '@/packages/utils/colors';
import { authClient } from '@/packages/lib/auth-client';
import { SettingsModalProvider } from '@/packages/hooks/shared/use-settings-modal';
import AppLayout from '@/packages/layouts/layout';
import { SettingsModal } from '@/packages/components/settings';
import Image from 'next/image';
import { useTheme } from 'next-themes';

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin']
});

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin']
});

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <Layout>{children}</Layout>;
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

const Layout = ({ children }: { children: React.ReactNode }) => {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        suppressHydrationWarning
      >
        <Provider store={store}>
          <PersistGate loading={null} persistor={persistor}>
            <ChildrenWrapper>{children}</ChildrenWrapper>
          </PersistGate>
        </Provider>
      </body>
    </html>
  );
};

const ChildrenWrapper = ({ children }: { children: React.ReactNode }) => {
  const dispatch = useAppDispatch();
  const pathname = usePathname();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const isInitialized = useAppSelector((state) => state.auth.isInitialized);

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

  useEffect(() => {
    const initAuth = async () => {
      await dispatch(initializeAuth() as any);
      setIsLoading(false);
    };
    initAuth();
  }, [dispatch]);

  // Re-check session when pathname changes to catch login state changes
  // This is a fallback in case Redux state hasn't updated yet
  useEffect(() => {
    if (!isInitialized) return;

    const checkSession = async () => {
      try {
        const session = await authClient.getSession();
        const hasSession = !!session?.data?.session;
        // If session exists but Redux says not authenticated, re-initialize
        // This can happen right after login before Redux state updates
        if (hasSession && !isAuthenticated) {
          await dispatch(initializeAuth() as any);
        }
      } catch (error) {
        // Session check failed, rely on Redux state
      }
    };

    // Only check on public routes to avoid unnecessary checks
    if (isPublicRoute) {
      checkSession();
    }
  }, [pathname, isAuthenticated, isInitialized, isPublicRoute, dispatch]);

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
        router.push('/apps');
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
          <AppLoadingScreen />
        ) : (
          <SettingsModalProvider>
            <WebSocketProvider>
              <FeatureFlagsProvider>
                {isPublicRoute ? (
                  <>{children}</>
                ) : (
                  <SystemStatsProvider>
                    <AppLayout>{children}</AppLayout>
                  </SystemStatsProvider>
                )}
              </FeatureFlagsProvider>
            </WebSocketProvider>
            <SettingsModal />
          </SettingsModalProvider>
        )}
      </ThemeProvider>
      <Toaster />
    </>
  );
};
