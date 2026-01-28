import { NextRequest, NextResponse } from 'next/server';

const getBackendUrl = () => {
  // Use AUTH_SERVICE_URL for server-side (Docker), fallback to NEXT_PUBLIC_AUTH_URL for client-side
  return (
    process.env.AUTH_SERVICE_URL || process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:9090'
  );
};

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ all: string[] }> }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams);
}

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ all: string[] }> }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams);
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ all: string[] }> }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ all: string[] }> }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams);
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ all: string[] }> }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams);
}

async function proxyRequest(request: NextRequest, params: { all: string[] }) {
  const backendUrl = getBackendUrl();
  const path = params.all ? `/${params.all.join('/')}` : '';
  const url = new URL(request.url);

  // Build the backend URL
  const backendAuthUrl = `${backendUrl}/api/auth${path}${url.search ? `?${url.searchParams.toString()}` : ''}`;

  try {
    // Get request body if present
    let body: BodyInit | undefined;

    if (request.method !== 'GET' && request.method !== 'HEAD') {
      const contentType = request.headers.get('content-type');

      // For form data, use formData() directly
      if (
        contentType?.includes('multipart/form-data') ||
        contentType?.includes('application/x-www-form-urlencoded')
      ) {
        body = await request.formData();
      }
      // For JSON, parse and stringify to ensure proper formatting
      else if (contentType?.includes('application/json')) {
        try {
          const json = await request.json();
          body = JSON.stringify(json);
        } catch {
          // If parsing fails, try as text
          body = await request.text();
        }
      }
      // For other types, use text or arrayBuffer
      else {
        body = await request.text();
      }
    }

    // Forward headers (excluding host and connection-specific headers)
    const headers = new Headers();
    request.headers.forEach((value, key) => {
      const lowerKey = key.toLowerCase();
      if (
        lowerKey !== 'host' &&
        lowerKey !== 'connection' &&
        lowerKey !== 'keep-alive' &&
        lowerKey !== 'transfer-encoding' &&
        lowerKey !== 'upgrade'
      ) {
        headers.set(key, value);
      }
    });

    // CRITICAL: Ensure Origin header is forwarded for Better Auth trusted origins validation
    // Better Auth validates requests against trustedOrigins based on the Origin header
    const origin = request.headers.get('origin') || request.headers.get('referer');
    if (origin) {
      // Extract origin from referer if origin header is missing
      const originUrl = origin.startsWith('http') ? origin : `https://${origin}`;
      try {
        const url = new URL(originUrl);
        headers.set('Origin', url.origin);
      } catch {
        // If parsing fails, use as-is
        headers.set('Origin', origin);
      }
    }

    // Explicitly ensure cookies are forwarded from the incoming request
    // In Node.js fetch, credentials: 'include' doesn't automatically forward cookies
    // from the incoming request, so we need to extract and add them manually
    // Check if cookies were already forwarded in the header loop above
    let cookieHeader = headers.get('cookie');

    if (!cookieHeader) {
      // Try to get cookies from request headers first
      cookieHeader = request.headers.get('cookie');
      if (cookieHeader) {
        headers.set('cookie', cookieHeader);
      } else {
        // Fallback: try to get cookies from NextRequest cookies() API
        const cookies = request.cookies.getAll();
        if (cookies.length > 0) {
          const cookieString = cookies.map((c) => `${c.name}=${c.value}`).join('; ');
          headers.set('cookie', cookieString);
          cookieHeader = cookieString;
        }
      }
    }

    const hasCookies = !!cookieHeader;
    const forwardedOrigin = headers.get('origin');
    console.log(
      `[Auth Proxy] ${request.method} ${path} - Cookies: ${hasCookies} (${cookieHeader ? cookieHeader.split(';').length : 0}), Origin: ${forwardedOrigin || 'none'}, Backend: ${backendAuthUrl}`
    );

    // Make the request to the backend
    // Note: credentials: 'include' doesn't work in Node.js fetch for forwarding cookies
    // We handle cookies explicitly above
    const response = await fetch(backendAuthUrl, {
      method: request.method,
      headers,
      body
    });

    // Get response body
    const responseBody = await response.text();

    // Create response with same status and headers
    const nextResponse = new NextResponse(responseBody, {
      status: response.status,
      statusText: response.statusText
    });

    // Forward response headers (especially Set-Cookie for Better Auth)
    // Use append() for Set-Cookie to preserve multiple cookie headers
    response.headers.forEach((value, key) => {
      const lowerKey = key.toLowerCase();
      if (lowerKey === 'set-cookie') {
        // Set-Cookie headers must be appended, not set, to preserve multiple cookies
        // In production, ensure cookies have proper domain/path settings
        // Better Auth should set these, but we ensure they're forwarded correctly
        nextResponse.headers.append(key, value);
      } else {
        nextResponse.headers.set(key, value);
      }
    });

    // Ensure CORS headers are set correctly
    if (!nextResponse.headers.has('Access-Control-Allow-Origin')) {
      const origin = request.headers.get('origin');
      if (origin) {
        nextResponse.headers.set('Access-Control-Allow-Origin', origin);
      }
    }
    nextResponse.headers.set('Access-Control-Allow-Credentials', 'true');

    return nextResponse;
  } catch (error) {
    console.error('Error proxying Better Auth request:', error);
    return NextResponse.json({ error: 'Failed to proxy request to auth server' }, { status: 500 });
  }
}
