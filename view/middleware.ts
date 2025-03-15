import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { isTokenExpired, refreshAccessToken, setAuthTokens } from './lib/auth';

export async function middleware(request: NextRequest) {
    const publicPaths = ['/login', '/register', '/api/auth', '/_next', '/static', '/favicon.ico'];
    if (publicPaths.some(path => request.nextUrl.pathname.startsWith(path))) {
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
            return NextResponse.redirect(new URL('/login', request.url));
        }
    }

    return NextResponse.next();
}

export const config = {
    matcher: [
        '/((?!api/v1/auth|_next/static|_next/image|favicon.ico).*)',
    ],
};