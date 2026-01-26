import { NextRequest, NextResponse } from 'next/server';

const getBackendUrl = () => {
  return process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:9090';
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

    // Make the request to the backend
    const response = await fetch(backendAuthUrl, {
      method: request.method,
      headers,
      body,
      // Important: include credentials (cookies) for Better Auth
      credentials: 'include'
    });

    // Get response body
    const responseBody = await response.text();

    // Create response with same status and headers
    const nextResponse = new NextResponse(responseBody, {
      status: response.status,
      statusText: response.statusText
    });

    // Forward response headers (especially Set-Cookie for Better Auth)
    response.headers.forEach((value, key) => {
      nextResponse.headers.set(key, value);
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
