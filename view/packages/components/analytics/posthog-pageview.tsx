'use client';

import { usePathname, useSearchParams } from 'next/navigation';
import { useEffect, useRef } from 'react';
import { usePostHog } from 'posthog-js/react';

function PostHogPageviewTracker() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const posthog = usePostHog();
  const lastUrlRef = useRef('');

  useEffect(() => {
    if (!posthog || !pathname) return;

    const url = searchParams?.toString() ? `${pathname}?${searchParams.toString()}` : pathname;

    if (url !== lastUrlRef.current) {
      lastUrlRef.current = url;
      posthog.capture('$pageview', { $current_url: window.origin + url });
    }
  }, [pathname, searchParams, posthog]);

  return null;
}

export default PostHogPageviewTracker;
