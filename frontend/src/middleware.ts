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

    // Forward relevant headers, but replace Host with the backend's host.
    // Must include: Authorization (JWT), X-CSRF-Token (double-submit CSRF),
    // Cookie (so backend can read rexio_csrf and rexio_refresh cookies).
    // Content-Type must be forwarded as-is (multipart/form-data needs the boundary param).
    const forwardHeaders = new Headers();
    const contentType = request.headers.get('Content-Type');
    if (contentType) {
      forwardHeaders.set('Content-Type', contentType);
    }
    const authHeader = request.headers.get('Authorization');
    if (authHeader) {
      forwardHeaders.set('Authorization', authHeader);
    }
    const csrfHeader = request.headers.get('X-CSRF-Token');
    if (csrfHeader) {
      forwardHeaders.set('X-CSRF-Token', csrfHeader);
    }
    // Forward cookies so the Go backend can read rexio_csrf (CSRF) and
    // rexio_refresh (refresh token rotation) cookies.
    const cookieHeader = request.headers.get('Cookie');
    if (cookieHeader) {
      forwardHeaders.set('Cookie', cookieHeader);
    }

    const response = await fetch(targetUrl, {
      method: request.method,
      headers: forwardHeaders,
      body: request.method !== 'GET' && request.method !== 'HEAD' ? request.body : undefined,
      // duplex required when streaming a request body (e.g. file uploads)
      // @ts-expect-error — duplex is valid in Node.js fetch but not in TypeScript's DOM lib yet
      duplex: 'half',
    });


    // Pass through the backend response without adding CORS headers.
    // The browser sees this as same-origin; no Access-Control-Allow-Origin needed.
    // Must forward Set-Cookie so the backend can set httpOnly refresh token cookie
    // and the readable rexio_csrf cookie on login/refresh/logout responses.
    const responseHeaders = new Headers();
    responseHeaders.set('Content-Type', response.headers.get('Content-Type') || 'application/json');
    // Forward all Set-Cookie headers (can be multiple — refresh token + CSRF cookie)
    response.headers.forEach((value, key) => {
      if (key.toLowerCase() === 'set-cookie') {
        responseHeaders.append('Set-Cookie', value);
      }
    });
    return new NextResponse(response.body, {
      status: response.status,
      headers: responseHeaders,
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
