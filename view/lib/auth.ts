import { parseCookies, setCookie, destroyCookie } from 'nookies';
import { jwtDecode } from 'jwt-decode';
import { BASE_URL } from '@/redux/conf';
import { AUTHURLS } from '@/redux/api-conf';

interface DecodedToken {
  exp: number;
}

interface AuthTokens {
  access_token: string;
  refresh_token?: string;
  expires_in?: number;
}

export function getToken(ctx?: any): string | null {
  const cookies = parseCookies(ctx);
  return cookies.token || null;
}

export function getRefreshToken(ctx?: any): string | null {
  const cookies = parseCookies(ctx);
  return cookies.refreshToken || null;
}

export function isTokenExpired(token: string | null): boolean {
  try {
    if (!token) return true;
    
    const decoded = jwtDecode<DecodedToken>(token);
    return decoded.exp * 1000 < Date.now() + 60000;
  } catch (error) {
    return true;
  }
}

export function setAuthTokens(tokens: AuthTokens, ctx?: any): void {
  const { access_token, refresh_token, expires_in } = tokens;
  
  setCookie(ctx, 'token', access_token, {
    maxAge: expires_in || 30 * 24 * 60 * 60, 
    path: '/',
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict'
  });
  
  if (refresh_token) {
    setCookie(ctx, 'refreshToken', refresh_token, {
      maxAge: 60 * 24 * 60 * 60, 
      path: '/',
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict'
    });
  }
}

export function clearAuthTokens(ctx?: any): void {
  destroyCookie(ctx, 'token', { path: '/' });
  destroyCookie(ctx, 'refreshToken', { path: '/' });
}

export async function refreshAccessToken(refreshToken: string): Promise<AuthTokens> {
  const response = await fetch(`${BASE_URL}${AUTHURLS.REFRESH_TOKEN}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!response.ok) {
    throw new Error('Failed to refresh token');
  }

  return await response.json();
}