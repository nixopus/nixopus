import { NextRequest } from 'next/server';

export const runtime = 'edge';

const AGENT_URL = process.env.NEXT_PUBLIC_AGENT_URL || '';

async function proxy(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  if (!AGENT_URL) {
    return new Response(JSON.stringify({ error: 'Agent URL not configured' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' }
    });
  }

  const { path } = await params;
  const targetPath = path.join('/');
  const url = new URL(`/${targetPath}`, AGENT_URL);
  req.nextUrl.searchParams.forEach((value, key) => {
    url.searchParams.set(key, value);
  });

  const headers = new Headers();
  const authorization = req.headers.get('authorization');
  const contentType = req.headers.get('content-type');
  const orgId = req.headers.get('x-organization-id');
  if (authorization) headers.set('Authorization', authorization);
  if (contentType) headers.set('Content-Type', contentType);
  if (orgId) headers.set('X-Organization-Id', orgId);

  const init: RequestInit = {
    method: req.method,
    headers
  };

  if (req.method !== 'GET' && req.method !== 'HEAD') {
    init.body = req.body;
    // @ts-expect-error -- edge runtime supports duplex streaming
    init.duplex = 'half';
  }

  const upstream = await fetch(url.toString(), init);

  return new Response(upstream.body, {
    status: upstream.status,
    headers: {
      'Content-Type': upstream.headers.get('Content-Type') || 'application/json',
      'Cache-Control': 'no-cache'
    }
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const DELETE = proxy;
export const PATCH = proxy;
