'use client';
import { Geist, Geist_Mono } from 'next/font/google';
import './globals.css';
import { ThemeProvider } from '@/components/ui/theme-provider';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '@/redux/store';
import { Toaster } from '@/components/ui/sonner';
import { useAppDispatch } from '@/redux/hooks';
import { useEffect } from 'react';
import { initializeAuth } from '@/redux/features/users/authSlice';
import { usePathname } from 'next/navigation';
import { WebSocketProvider } from '@/hooks/socket-provider';
import DashboardLayout from '@/components/layout/dashboard-layout';
import { FeatureFlagsProvider } from '@/hooks/features_provider';
import { palette } from '@/components/colors';

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

  useEffect(() => {
    dispatch(initializeAuth() as any);
  }, [dispatch]);

  return (
    <>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange  themes={palette}>
        <WebSocketProvider>
          <FeatureFlagsProvider>
            {pathname === '/' || pathname === '/login' || pathname === '/register' ? (
              <>{children}</>
            ) : (
              <DashboardLayout>{children}</DashboardLayout>
            )}
          </FeatureFlagsProvider>
        </WebSocketProvider>
      </ThemeProvider>
      <Toaster />
    </>
  );
};
