/**
 * Search API route — proxies to the Go backend.
 *
 * Per AGENTS.md D1: Frontend never calls Supabase directly.
 * This route proxies to the Go backend's /api/search endpoint.
 */

import { NextRequest, NextResponse } from 'next/server';
import type { SearchResponse } from '@/lib/types';

const BACKEND_URL =
  process.env.API_PROXY_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:10888';

export async function GET(request: NextRequest) {
  const { searchParams } = request.nextUrl;
  const searchQuery = searchParams.get('q') || '';
  const type = searchParams.get('type') || '';
  const page = searchParams.get('page') || '1';
  const perPage = searchParams.get('per_page') || '10';

  // Build backend query params
  const backendParams = new URLSearchParams();
  if (searchQuery) backendParams.set('q', searchQuery);
  if (type) backendParams.set('type', type);
  backendParams.set('page', page);
  backendParams.set('per_page', perPage);

  try {
    const response = await fetch(
      `${BACKEND_URL}/api/search?${backendParams.toString()}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        // Forward auth cookies if present
        credentials: 'include',
      }
    );

    // Clone the response to read it
    const data = await response.json() as SearchResponse;

    return NextResponse.json(data, {
      status: response.status,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  } catch {
    return NextResponse.json(
      {
        success: false,
        data: null,
        meta: null,
        error: {
          code: 'NETWORK_ERROR',
          message: 'Failed to fetch search results',
        },
      },
      { status: 500 }
    );
  }
}
