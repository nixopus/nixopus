'use client';
import { Geist, Geist_Mono } from 'next/font/google';
import './globals.css';
import { ThemeProvider } from '@/components/ui/theme-provider';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '@/redux/store';
import { Toaster } from '@/components/ui/sonner';
import { useAppDispatch, useAppSelector } from '@/redux/hooks';
import { useEffect } from 'react';
import { initializeAuth } from '@/redux/features/users/authSlice';
import { usePathname, useRouter } from 'next/navigation';
import { WebSocketProvider } from '@/hooks/socket-provider';
import DashboardLayout from '@/components/layout/dashboard-layout';
import { FeatureFlagsProvider } from '@/hooks/features_provider';
import { palette } from '@/components/colors';
import { SuperTokensProvider } from '@/components/supertokensProvider';
import { useSessionContext } from 'supertokens-auth-react/recipe/session';

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

const Layout = ({ children }: { children: React.ReactNode }) => {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        suppressHydrationWarning
      >
        <Provider store={store}>
          <PersistGate loading={null} persistor={persistor}>
            <SuperTokensProvider>
              <ChildrenWrapper>{children}</ChildrenWrapper>
            </SuperTokensProvider>
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
  const session = useSessionContext();

  useEffect(() => {
    dispatch(initializeAuth() as any);
  }, [dispatch]);

  const PUBLIC_ROUTES = [
    '/login',
    '/register',
    '/auth',
    '/reset-password',
    '/verify-email',
    '/auth/organization-invite'
  ];
  const isPublicRoute = PUBLIC_ROUTES.some(
    (route) => pathname === route || pathname.startsWith(route + '/')
  );

  useEffect(() => {
    if (session.loading) return;

    const sessionExists = 'doesSessionExist' in session ? session.doesSessionExist : false;

    if (!isPublicRoute && !sessionExists) {
      router.push('/auth');
    } else if (isPublicRoute && sessionExists) {
      if (
        pathname === '/' ||
        pathname === '/auth' ||
        pathname === '/login' ||
        pathname === '/register'
      ) {
        router.push('/dashboard');
      }
    }
  }, [pathname, session.loading, router, isPublicRoute]);

  if (session.loading) {
    return (
      <div className="flex h-screen flex-col items-center justify-center bg-background">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600"></div>
      </div>
    );
  }

  return (
    <>
      <ThemeProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        disableTransitionOnChange
        themes={palette}
      >
        <WebSocketProvider>
          <FeatureFlagsProvider>
            {isPublicRoute ? <>{children}</> : <DashboardLayout>{children}</DashboardLayout>}
          </FeatureFlagsProvider>
        </WebSocketProvider>
      </ThemeProvider>
      <Toaster />
    </>
  );
};
