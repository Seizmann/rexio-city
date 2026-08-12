/**
 * Auth group layout — wraps login and signup pages.
 * No top bar or navigation; just the centered card layout.
 * Server component (no 'use client' needed).
 */
export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
