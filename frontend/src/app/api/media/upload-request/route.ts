/**
 * /api/media/upload-request — Request presigned URL for direct R2 upload
 *
 * Flow:
 * 1. Frontend sends metadata (filename, content-type, size)
 * 2. Next.js proxies to Go backend
 * 3. Go backend generates presigned URL for direct R2 PUT
 * 4. Frontend receives URL and streams file directly to R2
 *
 * This bypasses Vercel's 4.5MB function body limit since no file bytes
 * pass through the Next.js function.
 */

import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL =
  process.env.API_PROXY_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:10888';

export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';

interface UploadRequest {
  filename: string;
  content_type: string;
  size: number;
}

export async function POST(request: NextRequest) {
  try {
    const targetUrl = `${BACKEND_URL}/api/media/upload-request`;

    // Parse the JSON body to validate it
    const bodyText = await request.text();

    // Parse and validate the JSON
    let parsedBody: UploadRequest | null = null;
    try {
      parsedBody = JSON.parse(bodyText) as UploadRequest;
    } catch {
      return NextResponse.json(
        {
          success: false,
          error: {
            code: 'INVALID_JSON',
            message: 'Invalid JSON in request body',
          },
        },
        { status: 400 },
      );
    }

    // Validate required fields
    if (!parsedBody.filename || !parsedBody.content_type || !parsedBody.size) {
      return NextResponse.json(
        {
          success: false,
          error: {
            code: 'INVALID_INPUT',
            message: 'filename, content_type, and size are required',
          },
        },
        { status: 400 },
      );
    }

    // Forward auth headers
    const forwardHeaders = new Headers();
    const auth = request.headers.get('Authorization');
    if (auth) forwardHeaders.set('Authorization', auth);
    const csrf = request.headers.get('X-CSRF-Token');
    if (csrf) forwardHeaders.set('X-CSRF-Token', csrf);
    const cookie = request.headers.get('Cookie');
    if (cookie) forwardHeaders.set('Cookie', cookie);
    forwardHeaders.set('Content-Type', 'application/json');

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
