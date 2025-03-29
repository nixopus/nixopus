import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { isTokenExpired, refreshAccessToken, setAuthTokens } from './lib/auth';

export async function middleware(request: NextRequest) {
  const publicPaths = ['/login', '/register', '/api/auth', '/_next', '/static', '/favicon.ico'];

  if (publicPaths.some((path) => request.nextUrl.pathname.startsWith(path))) {
    return NextResponse.next();
  }

  const isGitHubFlow =
    (request.nextUrl.pathname.startsWith('/self-host') ||
      request.nextUrl.pathname.startsWith('/github-callback')) &&
    (request.nextUrl.searchParams.has('code') ||
      request.nextUrl.searchParams.has('installation_id') ||
      request.nextUrl.searchParams.has('setup_action'));

  if (isGitHubFlow) {
    return NextResponse.next();
  }

  const token = request.cookies.get('token')?.value;
  const refreshToken = request.cookies.get('refreshToken')?.value;

  if (!token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  if (isTokenExpired(token) && refreshToken) {
    try {
      const newTokens = await refreshAccessToken(refreshToken);
      const response = NextResponse.next();

      response.cookies.set({
        name: 'token',
        value: newTokens.access_token,
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: newTokens.expires_in || 30 * 24 * 60 * 60
      });

      if (newTokens.refresh_token) {
        response.cookies.set({
          name: 'refreshToken',
          value: newTokens.refresh_token,
          httpOnly: true,
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'strict',
          maxAge: 60 * 24 * 60 * 60
        });
      }

      return response;
    } catch (error) {
      const response = NextResponse.redirect(new URL('/login', request.url));

      response.cookies.delete('token');
      response.cookies.delete('refreshToken');

      return response;
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/((?!api/v1/auth|_next/static|_next/image|favicon.ico).*)',
    '/self-host/:path*',
    '/github-callback/:path*'
  ]
};
