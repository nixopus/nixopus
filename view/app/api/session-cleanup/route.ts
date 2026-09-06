import { NextRequest, NextResponse } from 'next/server';

/**
 * Expires Better Auth cookies across every scope they could be stored under.
 *
 * Sign-out deletes cookies using the attributes configured right now. A cookie
 * stored under a different domain or path — left by an AUTH_COOKIE_DOMAIN
 * change, or by the app being reachable on both the apex and a subdomain — is
 * a different cookie to the browser, so that deletion never matches it. The
 * orphan keeps being sent and no client code can remove it, because HttpOnly
 * cookies are unwritable from document.cookie. Only a server response clears
 * them.
 */

const BASE_PATH = process.env.BASE_PATH || '';

const BASE_COOKIE_NAMES = [
  'better-auth.session_token',
  'better-auth.session_data',
  'better-auth.dont_remember'
];

// A cookie at /auth is not sent to /chats and is not cleared by a Path=/
// deletion, so cover the auth landings as well as the root.
const PATHS = ['/', '/auth', '/login', '/register'].flatMap((path) =>
  BASE_PATH ? [path, `${BASE_PATH}${path === '/' ? '/' : path}`] : [path]
);

const SECURE_PREFIX = '__Secure-';
const COOKIE_PREFIX = 'better-auth.';

function isSameOrigin(request: NextRequest): boolean {
  const origin = request.headers.get('origin');
  if (!origin) return true; // Same-origin GET/POST may omit Origin entirely.
  try {
    return new URL(origin).host === request.headers.get('host');
  } catch {
    return false;
  }
}

function isSecureRequest(request: NextRequest): boolean {
  const forwarded = request.headers.get('x-forwarded-proto');
  if (forwarded) return forwarded.split(',')[0].trim() === 'https';
  return request.nextUrl.protocol === 'https:';
}

// The known names plus any Better Auth cookie this browser actually sent,
// which picks up chunked variants (session_data.0, session_data.1, ...)
// without emitting headers for chunks that do not exist.
function cookieNames(request: NextRequest, secure: boolean): string[] {
  const names = new Set(BASE_COOKIE_NAMES);

  for (const cookie of request.cookies.getAll()) {
    const bare = cookie.name.startsWith(SECURE_PREFIX)
      ? cookie.name.slice(SECURE_PREFIX.length)
      : cookie.name;
    if (bare.startsWith(COOKIE_PREFIX)) names.add(bare);
  }

  // A __Secure- cookie can only be set over https, so only clear one when the
  // request is secure — otherwise the browser rejects the header.
  return secure ? [...names].flatMap((n) => [n, `${SECURE_PREFIX}${n}`]) : [...names];
}

// Host-only (undefined) plus every parent domain the cookie could be scoped
// to: app.example.com yields undefined, .app.example.com, .example.com. Stops
// at two labels so we never target a public suffix like .com, which browsers
// reject. localhost and bare IPs cannot carry domain cookies at all.
function candidateDomains(host: string | null): (string | undefined)[] {
  const hostname = (host || '').split(':')[0].toLowerCase();
  const domains: (string | undefined)[] = [undefined];

  if (!hostname || hostname === 'localhost' || /^[\d.]+$/.test(hostname)) {
    return domains;
  }

  const labels = hostname.split('.');
  for (let i = 0; i <= labels.length - 2; i++) {
    domains.push(`.${labels.slice(i).join('.')}`);
  }
  return domains;
}

function expiredCookie(name: string, path: string, domain: string | undefined): string {
  const parts = [`${name}=`, 'Max-Age=0', 'Expires=Thu, 01 Jan 1970 00:00:00 GMT', `Path=${path}`];
  if (domain) parts.push(`Domain=${domain}`);
  parts.push('HttpOnly', 'SameSite=Lax');
  // Browsers reject a __Secure- cookie set without Secure, which would make
  // the deletion silently do nothing.
  if (name.startsWith(SECURE_PREFIX)) parts.push('Secure');
  return parts.join('; ');
}

export async function POST(request: NextRequest) {
  // Without this a cross-site POST could force-log-out a signed-in user.
  if (!isSameOrigin(request)) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }

  const response = NextResponse.json({ cleared: true });
  const secure = isSecureRequest(request);
  const domains = candidateDomains(request.headers.get('host'));

  for (const name of cookieNames(request, secure)) {
    for (const path of PATHS) {
      for (const domain of domains) {
        response.headers.append('Set-Cookie', expiredCookie(name, path, domain));
      }
    }
  }

  return response;
}
