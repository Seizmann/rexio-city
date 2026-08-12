'use client';

import { useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { ROUTES } from '@/lib/constants';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import styles from '../auth.module.css';

export default function LoginPage() {
  const router = useRouter();
  const { login, isAuthenticated } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  // If already authenticated, redirect to home
  if (isAuthenticated) {
    router.replace(ROUTES.HOME);
    return null;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');

    if (!email || !password) {
      setError('Email and password are required.');
      return;
    }

    setLoading(true);
    try {
      await login(email, password);
      router.push(ROUTES.HOME);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed. Check your credentials.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className={styles.authPage}>
      <div className={styles.card}>
        <div className={styles.logo}>
          <div className={styles.logoText}>RexiO City</div>
          <div className={styles.logoSub}>Log in to your account</div>
        </div>

        <form onSubmit={(e) => void handleSubmit(e)} className={styles.form}>
          {error && <div className={styles.formError}>{error}</div>}

          <Input
            label="Email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
            required
          />

          <Input
            label="Password"
            type="password"
            placeholder="Your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            required
          />

          <Button type="submit" loading={loading} fullWidth>
            Log in
          </Button>
        </form>

        <div className={styles.footer}>
          No account yet?{' '}
          <Link href={ROUTES.SIGNUP} className={styles.footerLink}>
            Sign up
          </Link>
        </div>
      </div>
    </div>
  );
}
