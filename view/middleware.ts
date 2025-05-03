import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { clearAuthTokens } from './lib/auth';
import { defaultLocale, locales } from './lib/i18n/config';
import { jwtDecode } from 'jwt-decode';

interface DecodedToken {
  exp: number;
  '2fa_enabled': boolean;
  '2fa_verified': boolean;
}

export async function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname;

  if (pathname === '/logout') {
    const response = NextResponse.redirect(new URL('/login', request.url));
    clearAuthTokens(response);
    return response;
  }

  const publicPaths = [
    '/login',
    '/register',
    '/api/auth',
    '/_next',
    '/static',
    '/favicon.ico',
    '/reset-password',
    '/verify-email',
    '/login'
  ];

  const isPublicPath = publicPaths.some((path) => pathname.includes(path));
  if (isPublicPath) {
    return NextResponse.next();
  }

  const isGitHubFlow =
    (pathname.includes('/self-host') || pathname.includes('/github-callback')) &&
    (request.nextUrl.searchParams.has('code') ||
      request.nextUrl.searchParams.has('installation_id') ||
      request.nextUrl.searchParams.has('setup_action'));

  const isPasswordResetFlow =
    pathname.includes('/reset-password') && request.nextUrl.searchParams.has('token');

  const isVerificationFlow =
    pathname.includes('/verify-email') && request.nextUrl.searchParams.has('token');

  if (isGitHubFlow || isPasswordResetFlow || isVerificationFlow) {
    return NextResponse.next();
  }

  const token = request.cookies.get('token')?.value;
  if (!token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  try {
    const decoded = jwtDecode<DecodedToken>(token);
    if (decoded['2fa_enabled'] && !decoded['2fa_verified']) {
      return NextResponse.redirect(new URL('/login', request.url));
    }
  } catch (error) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
    '/self-host/:path*',
    '/github-callback/:path*'
  ]
};

export function i18nMiddleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname;
  const pathnameIsMissingLocale = locales.every(
    (locale) => !pathname.startsWith(`/${locale}/`) && pathname !== `/${locale}`
  );

  if (pathnameIsMissingLocale) {
    const locale = defaultLocale;
    return NextResponse.redirect(new URL(`/${locale}${pathname}`, request.url));
  }
}
