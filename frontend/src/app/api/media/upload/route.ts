/**
 * /api/media/upload — Next.js Route Handler (Node.js runtime)
 *
 * Proxies multipart/form-data file uploads to the Go backend.
 * This MUST be a Route Handler (not middleware) because:
 *   - Next.js Edge middleware cannot stream large binary request bodies.
 *   - Route Handlers run in the Node.js runtime which has no body size limits
 *     imposed by the edge sandbox.
 *
 * The middleware's matcher is '/api/:path*', but Next.js gives Route Handlers
 * priority over middleware for their own paths, so this file intercepts
 * POST /api/media/upload before middleware sees it.
 */

import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL =
  process.env.API_PROXY_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:10888';

export const runtime = 'nodejs'; // Must be Node.js — edge cannot handle large binary bodies
export const maxDuration = 60;   // Allow up to 60s for large video uploads

export async function POST(request: NextRequest) {
  try {
    const targetUrl = `${BACKEND_URL}/api/media/upload`;

    // Forward auth, CSRF, and cookie headers — same as the main proxy middleware.
    const forwardHeaders = new Headers();
    const auth = request.headers.get('Authorization');
    if (auth) forwardHeaders.set('Authorization', auth);
    const csrf = request.headers.get('X-CSRF-Token');
    if (csrf) forwardHeaders.set('X-CSRF-Token', csrf);
    const cookie = request.headers.get('Cookie');
    if (cookie) forwardHeaders.set('Cookie', cookie);
    // Do NOT set Content-Type — let fetch derive it from the FormData body
    // so the multipart boundary is included automatically.

    // Stream the raw request body directly to the Go backend.
    // request.body is a ReadableStream; passing it avoids buffering the whole
    // file in memory, which matters for videos up to 500 MB.
    const response = await fetch(targetUrl, {
      method: 'POST',
      headers: forwardHeaders,
      // Forward the raw body stream (multipart/form-data) including boundary
      body: request.body,
      // @ts-expect-error — duplex is required for streaming bodies in Node.js fetch
      duplex: 'half',
    });

    const responseHeaders = new Headers();
    responseHeaders.set(
      'Content-Type',
      response.headers.get('Content-Type') || 'application/json',
    );
    // Forward Set-Cookie if backend sets any on this endpoint
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
          code: 'UPLOAD_PROXY_ERROR',
          message: error instanceof Error ? error.message : 'Upload proxy failed',
        },
      },
      { status: 502 },
    );
  }
}
