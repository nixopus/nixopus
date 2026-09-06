import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { getSessionCookie } from 'better-auth/cookies';
import { getPluginPrivatePaths, getPluginPublicPaths } from '@/plugins/registry';

const BASE_PATH = process.env.BASE_PATH || '';

const CORE_PUBLIC_PATHS = ['/auth', '/login', '/register', '/reset-password', '/verify-email'];
const CORE_PRIVATE_PATHS = [
  '/apps',
  '/charts',
  '/chats',
  '/extensions',
  '/machines',
  '/settings',
  '/activities',
  '/domains',
  '/api-keys',
  '/security',
  '/backups'
];

const PUBLIC_PATHS = [...CORE_PUBLIC_PATHS, ...getPluginPublicPaths()];
const PRIVATE_PATHS = [...CORE_PRIVATE_PATHS, ...getPluginPrivatePaths()];
const AUTH_BYPASS_PATHS = ['/apps/github-callback', '/charts/github-callback'];

function getPath(pathname: string) {
  if (BASE_PATH && pathname.startsWith(BASE_PATH)) {
    return pathname.slice(BASE_PATH.length) || '/';
  }
  return pathname;
}

function isPublicPath(path: string) {
  return PUBLIC_PATHS.some((p) => path === p || path.startsWith(p + '/'));
}

function isPrivatePath(path: string) {
  return PRIVATE_PATHS.some((p) => path === p || path.startsWith(p + '/'));
}

export function middleware(request: NextRequest) {
  const path = getPath(request.nextUrl.pathname);
  // A session cookie may be expired, revoked, or an orphan left at a scope
  // sign-out cannot clear, so it only gates private pages — never redirects
  // anyone off the login page, which would make a stale cookie loop
  // /auth -> /chats -> /auth until the browser gives up. Sending a signed-in
  // user to /chats is client-layout.tsx's job; it checks a verified session.
  // getSessionCookie() also matches the __Secure- prefixed name.
  const hasAuth = !!getSessionCookie(request);

  const isAuthBypass = AUTH_BYPASS_PATHS.some((p) => path === p || path.startsWith(p + '/'));
  if (!hasAuth && (path === '/' || isPrivatePath(path)) && !isPublicPath(path) && !isAuthBypass) {
    const url = request.nextUrl.clone();
    url.pathname = BASE_PATH ? `${BASE_PATH}/auth` : '/auth';
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico|logo_.*\\.png|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)'
  ]
};
