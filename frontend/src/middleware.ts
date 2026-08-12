import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// API proxy to forward requests to Go backend
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:10888';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  
  // Skip if not an API path
  if (!pathname.startsWith('/api/')) {
    return NextResponse.next();
  }

  try {
    // Forward the request to the Go backend
    const response = await fetch(`${API_URL}${pathname}`, {
      method: request.method,
      headers: {
        'Content-Type': 'application/json',
        ...Object.fromEntries(request.headers),
      },
      body: request.body,
    });

    // Return the response from backend
    return new NextResponse(response.body, {
      status: response.status,
      headers: {
        'Content-Type': response.headers.get('Content-Type') || 'application/json',
        'Access-Control-Allow-Origin': '*',
      },
    });
  } catch (error) {
    return NextResponse.json(
      { error: 'Proxy error', message: error instanceof Error ? error.message : 'Unknown error' },
      { status: 502 }
    );
  }
}

export const config = {
  matcher: '/api/:path*',
};
