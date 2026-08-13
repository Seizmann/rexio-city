/**
 * /api/media/upload-complete — Confirm direct R2 upload and get media URL
 *
 * Flow:
 * 1. Frontend receives presigned URL from upload-request
 * 2. Frontend PUTs file directly to R2
 * 3. Frontend calls this endpoint to confirm upload
 * 4. Go backend verifies file size in R2
 * 5. Returns media URL for post creation
 */

import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL =
  process.env.API_PROXY_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:10888';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  try {
    const targetUrl = `${BACKEND_URL}/api/media/upload-complete`;

    // Forward auth headers
    const forwardHeaders = new Headers();
    const auth = request.headers.get('Authorization');
    if (auth) forwardHeaders.set('Authorization', auth);
    const csrf = request.headers.get('X-CSRF-Token');
    if (csrf) forwardHeaders.set('X-CSRF-Token', csrf);
    const cookie = request.headers.get('Cookie');
    if (cookie) forwardHeaders.set('Cookie', cookie);

    // Read body as text to avoid type issues
    const bodyText = await request.text();
    const response = await fetch(targetUrl, {
      method: 'POST',
      headers: forwardHeaders,
      body: bodyText,
    });

    const responseHeaders = new Headers();
    responseHeaders.set(
      'Content-Type',
      response.headers.get('Content-Type') || 'application/json',
    );
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
          message: error instanceof Error ? error.message : 'Request failed',
        },
      },
      { status: 502 },
    );
  }
}
