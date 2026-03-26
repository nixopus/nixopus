import posthog from 'posthog-js';

let posthogClient: typeof posthog | null = null;

export async function initPostHog(key: string, host: string): Promise<typeof posthog | null> {
  if (typeof window === 'undefined' || !key) return null;

  if (!posthogClient) {
    posthog.init(key, {
      api_host: host || 'https://eu.i.posthog.com',
      ui_host: 'https://eu.posthog.com',
      capture_pageview: false,
      capture_pageleave: true,
      person_profiles: 'identified_only',
      autocapture: true,
      disable_session_recording: false,
      session_recording: {
        maskAllInputs: true
      }
    });
    posthogClient = posthog;
  }

  return posthogClient;
}

export function getPostHogClient(): typeof posthog | null {
  return posthogClient;
}
