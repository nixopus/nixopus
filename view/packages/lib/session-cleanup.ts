/**
 * Asks the server to expire Better Auth cookies across every scope they could
 * be stored under. See app/api/session-cleanup/route.ts for why sign-out alone
 * cannot always reach them.
 *
 * Best effort: a failure here must never block signing out or rendering the
 * login page.
 */
export async function clearStaleAuthCookies(): Promise<void> {
  if (typeof window === 'undefined') return;

  const basePath = process.env.NEXT_PUBLIC_BASE_PATH || '';

  try {
    await fetch(`${basePath}/api/session-cleanup`, {
      method: 'POST',
      credentials: 'include',
      cache: 'no-store'
    });
  } catch {
    // Offline or blocked; the local sign-out still proceeds.
  }
}
