'use client';

import { PostHogProvider as PHProvider, usePostHog } from 'posthog-js/react';
import { useEffect, useRef, useState, Suspense } from 'react';
import type posthogLib from 'posthog-js';
import { useAppSelector } from '@/redux/hooks';
import { initPostHog } from '@/packages/lib/posthog';
import { getPostHogKey, getPostHogHost, getSelfHosted } from '@/redux/conf';
import PostHogPageviewTracker from '@/packages/components/analytics/posthog-pageview';

function PostHogIdentify() {
  const posthog = usePostHog();
  const { isAuthenticated, isInitialized, isLoading, user } = useAppSelector((state) => state.auth);

  const prevAuthRef = useRef(false);
  const prevUserPropsRef = useRef('');

  useEffect(() => {
    if (!posthog || !isInitialized || isLoading) return;

    if (isAuthenticated && user) {
      const userProps = JSON.stringify({
        id: user.id,
        email: user.email,
        name: user.username
      });

      if (userProps !== prevUserPropsRef.current) {
        prevUserPropsRef.current = userProps;
        posthog.identify(user.id, {
          email: user.email,
          name: user.username
        });
      }
      prevAuthRef.current = true;
    } else if (prevAuthRef.current) {
      posthog.reset();
      prevAuthRef.current = false;
      prevUserPropsRef.current = '';
    }
  }, [isAuthenticated, isInitialized, isLoading, user, posthog]);

  return null;
}

export function CSPostHogProvider({ children }: { children: React.ReactNode }) {
  const [client, setClient] = useState<typeof posthogLib | null>(null);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    let cancelled = false;
    async function init() {
      const [key, host, selfHosted] = await Promise.all([
        getPostHogKey(),
        getPostHogHost(),
        getSelfHosted()
      ]);
      if (cancelled) return;

      if (key && !selfHosted) {
        const ph = await initPostHog(key, host);
        if (!cancelled) setClient(ph);
      }
      if (!cancelled) setLoaded(true);
    }
    init();
    return () => {
      cancelled = true;
    };
  }, []);

  if (!loaded || !client) {
    return <>{children}</>;
  }

  return (
    <PHProvider client={client}>
      <PostHogIdentify />
      <Suspense fallback={null}>
        <PostHogPageviewTracker />
      </Suspense>
      {children}
    </PHProvider>
  );
}
