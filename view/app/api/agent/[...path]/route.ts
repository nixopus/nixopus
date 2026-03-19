import { NextRequest } from 'next/server';

export const runtime = 'edge';

const AGENT_URL = process.env.AGENT_URL || '';

async function proxy(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  if (!AGENT_URL) {
    return new Response(JSON.stringify({ error: 'Agent URL not configured' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' }
    });
  }

  const { path } = await params;
  const targetPath = path.join('/');
  const base = new URL(AGENT_URL);
  const url = new URL(base.toString());
  const basePath = base.pathname.replace(/\/+$/, '');
  const normalizedTargetPath =
    basePath.endsWith('/api') && targetPath.startsWith('api/') ? targetPath.slice(4) : targetPath;
  url.pathname = `${basePath}/${normalizedTargetPath}`;
  req.nextUrl.searchParams.forEach((value, key) => {
    url.searchParams.set(key, value);
  });

  const headers = new Headers();
  const authorization = req.headers.get('authorization');
  const contentType = req.headers.get('content-type');
  const cookie = req.headers.get('cookie');
  const accept = req.headers.get('accept');
  const orgId = req.headers.get('x-organization-id');
  const appId = req.headers.get('x-application-id');
  if (authorization) headers.set('Authorization', authorization);
  if (contentType) headers.set('Content-Type', contentType);
  if (cookie) headers.set('Cookie', cookie);
  if (accept) headers.set('Accept', accept);
  if (orgId) headers.set('X-Organization-Id', orgId);
  if (appId) headers.set('X-Application-Id', appId);

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

  const upstreamContentType = upstream.headers.get('Content-Type') || 'application/json';
  const isStreaming =
    upstreamContentType.includes('text/event-stream') ||
    upstreamContentType.includes('text/plain') ||
    upstreamContentType.includes('application/x-ndjson');

  const responseHeaders: Record<string, string> = {
    'Content-Type': isStreaming ? 'text/event-stream; charset=utf-8' : upstreamContentType,
    'Cache-Control': 'no-cache, no-transform',
    'X-Content-Type-Options': 'nosniff'
  };

  if (isStreaming) {
    responseHeaders['Connection'] = 'keep-alive';
    responseHeaders['X-Accel-Buffering'] = 'no';
  }

  return new Response(upstream.body, {
    status: upstream.status,
    headers: responseHeaders
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const DELETE = proxy;
export const PATCH = proxy;
