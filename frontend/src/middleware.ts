import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Next.js middleware — proxies /api/* requests to the Go backend.
 *
 * Per AGENTS.md D1: "The frontend/ app never calls Supabase directly.
 * All data access goes through Next.js API routes, which call the Go backend."
 *
 * This middleware acts as that bridge: the browser hits the Next.js origin,
 * and the server forwards the request to the Go backend. No CORS needed
 * since the browser sees same-origin requests.
 *
 * The Go backend URL is read from API_PROXY_URL (server-only, not NEXT_PUBLIC_).
 * Falls back to NEXT_PUBLIC_API_URL for backwards compat, then localhost.
 */
const BACKEND_URL =
  process.env.API_PROXY_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:10888';

export async function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl;

  // Only proxy /api/* paths
  if (!pathname.startsWith('/api/')) {
    return NextResponse.next();
  }

  try {
    // Build the target URL on the Go backend
    const targetUrl = `${BACKEND_URL}${pathname}${search}`;

    // Forward relevant headers, but replace Host with the backend's host
    const forwardHeaders = new Headers();
    forwardHeaders.set('Content-Type', request.headers.get('Content-Type') || 'application/json');
    const authHeader = request.headers.get('Authorization');
    if (authHeader) {
      forwardHeaders.set('Authorization', authHeader);
    }

    const response = await fetch(targetUrl, {
      method: request.method,
      headers: forwardHeaders,
      body: request.method !== 'GET' && request.method !== 'HEAD' ? request.body : undefined,
    });

    // Pass through the backend response without adding CORS headers.
    // The browser sees this as same-origin; no Access-Control-Allow-Origin needed.
    return new NextResponse(response.body, {
      status: response.status,
      headers: {
        'Content-Type': response.headers.get('Content-Type') || 'application/json',
      },
    });
  } catch (error) {
    return NextResponse.json(
      {
        success: false,
        error: {
          code: 'PROXY_ERROR',
          message: error instanceof Error ? error.message : 'Backend unreachable',
        },
      },
      { status: 502 },
    );
  }
}

export const config = {
  matcher: '/api/:path*',
};
