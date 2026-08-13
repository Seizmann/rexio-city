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

  // Only proxy /api/* and /uploads/* paths
  if (!pathname.startsWith('/api/') && !pathname.startsWith('/uploads/')) {
    return NextResponse.next();
  }

  // DEBUG: Log request size for mobile debugging
  const cookieHeader = request.headers.get('Cookie');
  const userAgent = request.headers.get('User-Agent') || 'unknown';
  // Count all header bytes manually
  let headerSize = 0;
  const headerPairs: string[] = [];
  request.headers.forEach((value, key) => {
    headerSize += key.length + value.length;
    headerPairs.push(`${key}: ${value.slice(0, 50)}...`);
  });
  console.log(`[middleware] ${pathname} | UA: ${userAgent.slice(0, 50)} | Headers: ${headerSize}B | Cookie: ${cookieHeader ? cookieHeader.length : 0}B`);
  if (headerSize > 5000) {
    console.log(`[middleware] Headers for ${pathname}:`, headerPairs.join(' | '));
  }

  // Block requests with excessive header sizes (Vercel limit is ~8KB)
  if (headerSize > 8000) {
    console.error('[middleware] Header size too large:', headerSize, 'bailing out');
    return new NextResponse(JSON.stringify({ success: false, error: { code: 'HEADER_TOO_LARGE', message: 'Request headers too large. Please refresh the page and try again.' } }), {
      status: 431,
      headers: { 'Content-Type': 'application/json' },
    });
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

    // IMPORTANT: Use getSetCookie() — NOT headers.forEach().
    // Node.js fetch merges all Set-Cookie values into one comma-joined string when
    // accessed via forEach/get, breaking multi-cookie responses (refresh + CSRF).
    // getSetCookie() returns a proper string[] preserving each cookie separately.
    //
    // Also strip the Domain= attribute from each cookie. The Go backend sets
    // Domain=rexio.pro (or empty = backend host). The browser must store the cookie
    // against the frontend origin (dev-city.rexio.pro / city.rexio.pro), so we let
    // the browser infer the domain from the response origin rather than using the
    // backend's domain directive.
    const setCookies = (response.headers as unknown as { getSetCookie?: () => string[] }).getSetCookie?.() ?? [];
    for (const cookie of setCookies) {
      // Strip Domain= attribute — let browser bind cookie to frontend origin
      const stripped = cookie.replace(/;?\s*Domain=[^;]*/gi, '');
      responseHeaders.append('Set-Cookie', stripped);
    }

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
  // Proxy /api/* (except /api/media/upload handled by Node.js Route Handlers) and /uploads/*
  // Exclude: /api/media/upload, /api/media/upload-request, /api/media/upload-complete
  matcher: ['/api/((?!media/upload(?:-request|-complete)?).*)', '/uploads/:path*'],
};
