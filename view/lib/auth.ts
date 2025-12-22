import { parseCookies, setCookie, destroyCookie } from 'nookies';
import { jwtDecode } from 'jwt-decode';

interface DecodedToken {
  exp: number;
}

interface AuthTokens {
  access_token: string;
  refresh_token?: string;
  expires_in?: number;
}

export function getToken(ctx?: any): string | null {
  if (ctx?.cookies) {
    return ctx.cookies.get('token')?.value || null;
  }
  const cookies = parseCookies(ctx);
  return cookies.token || null;
}

export function getRefreshToken(ctx?: any): string | null {
  if (ctx?.cookies) {
    return ctx.cookies.get('refreshToken')?.value || null;
  }
  const cookies = parseCookies(ctx);
  return cookies.refreshToken || null;
}

export function isTokenExpired(token: string | null): boolean {
  try {
    if (!token) return true;

    if (!token.includes('.')) {
      console.error('Invalid token format: missing separator');
      return true;
    }

    const decoded = jwtDecode<DecodedToken>(token);
    if (!decoded.exp) {
      console.error('Invalid token: missing expiration');
      return true;
    }

    console.log(decoded.exp * 1000, Date.now() + 1 * 60 * 1000);

    return decoded.exp * 1000 < Date.now() + 1 * 60 * 1000;
  } catch (error) {
    console.error('Error checking token expiration:', error);
    return true;
  }
}

export function setAuthTokens(tokens: AuthTokens, ctx?: any): void {
  const { access_token, refresh_token, expires_in } = tokens;

  if (ctx?.cookies) {
    ctx.cookies.set('token', access_token, {
      maxAge: expires_in || 7 * 24 * 60 * 60,
      path: '/',
      secure: false,
      sameSite: 'lax'
    });

    if (refresh_token) {
      ctx.cookies.set('refreshToken', refresh_token, {
        maxAge: 14 * 24 * 60 * 60,
        path: '/',
        secure: false,
        sameSite: 'lax'
      });
    }
  } else {
    setCookie(ctx, 'token', access_token, {
      maxAge: expires_in || 7 * 24 * 60 * 60,
      path: '/',
      secure: false,
      sameSite: 'lax'
    });

    if (refresh_token) {
      setCookie(ctx, 'refreshToken', refresh_token, {
        maxAge: 14 * 24 * 60 * 60,
        path: '/',
        secure: false,
        sameSite: 'lax'
      });
    }
  }
}

export function clearAuthTokens(ctx?: any): void {
  if (ctx?.cookies) {
    ctx.cookies.delete('token', { path: '/' });
    ctx.cookies.delete('refreshToken', { path: '/' });
  } else {
    // Client-side context
    destroyCookie(ctx, 'token', { path: '/' });
    destroyCookie(ctx, 'refreshToken', { path: '/' });
  }
}
