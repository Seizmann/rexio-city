import { Inter } from 'next/font/google';
import type { Metadata } from 'next';
import Providers from './providers';
import './globals.css';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });

export const metadata: Metadata = {
  title: 'RexiO City — Social Platform',
  description:
    'A public social platform combining short-form posting with media presentation. Built by AI.',
  metadataBase: new URL(
    process.env.NEXT_PUBLIC_APP_URL || 'https://city.rexio.pro',
  ),
  icons: {
    icon: '/icon.webp',
    shortcut: '/icon.webp',
    apple: '/icon.webp',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={inter.variable} suppressHydrationWarning>
      <body className={inter.className} suppressHydrationWarning>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
