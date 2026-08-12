'use client';

import { useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { ROUTES, LIMITS } from '@/lib/constants';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import styles from '../auth.module.css';

export default function SignupPage() {
  const router = useRouter();
  const { signup, isAuthenticated } = useAuth();

  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState('');
  const [loading, setLoading] = useState(false);

  // If already authenticated, redirect to home
  if (isAuthenticated) {
    router.replace(ROUTES.HOME);
    return null;
  }

  function validate(): boolean {
    const newErrors: Record<string, string> = {};

    if (username.length < LIMITS.USERNAME_MIN_CHARS) {
      newErrors.username = `Username must be at least ${LIMITS.USERNAME_MIN_CHARS} characters.`;
    } else if (username.length > LIMITS.USERNAME_MAX_CHARS) {
      newErrors.username = `Username must be at most ${LIMITS.USERNAME_MAX_CHARS} characters.`;
    } else if (!/^[a-z0-9_]+$/.test(username)) {
      newErrors.username = 'Username must be lowercase letters, numbers, or underscores.';
    }

    if (displayName.length > LIMITS.DISPLAY_NAME_MAX_CHARS) {
      newErrors.displayName = `Display name must be at most ${LIMITS.DISPLAY_NAME_MAX_CHARS} characters.`;
    }

    if (!email.includes('@')) {
      newErrors.email = 'Enter a valid email address.';
    }

    if (password.length < LIMITS.PASSWORD_MIN_CHARS) {
      newErrors.password = `Password must be at least ${LIMITS.PASSWORD_MIN_CHARS} characters.`;
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError('');

    if (!validate()) return;

    setLoading(true);
    try {
      await signup({
        username,
        email,
        password,
        display_name: displayName || username,
      });
      router.push(ROUTES.HOME);
    } catch (err) {
      setFormError(
        err instanceof Error ? err.message : 'Signup failed. Try a different username or email.',
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className={styles.authPage}>
      <div className={styles.card}>
        <div className={styles.logo}>
          <div className={styles.logoText}>RexiO City</div>
          <div className={styles.logoSub}>Create your account</div>
        </div>

        <form onSubmit={(e) => void handleSubmit(e)} className={styles.form}>
          {formError && <div className={styles.formError}>{formError}</div>}

          <Input
            label="Username"
            type="text"
            placeholder="your_username"
            value={username}
            onChange={(e) => setUsername(e.target.value.toLowerCase())}
            error={errors.username}
            autoComplete="username"
            required
          />

          <Input
            label="Display name"
            type="text"
            placeholder="How others see you (optional)"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            error={errors.displayName}
          />

          <Input
            label="Email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={errors.email}
            autoComplete="email"
            required
          />

          <Input
            label="Password"
            type="password"
            placeholder="At least 8 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
            autoComplete="new-password"
            required
          />

          <Button type="submit" loading={loading} fullWidth>
            Sign up
          </Button>
        </form>

        <div className={styles.footer}>
          Already have an account?{' '}
          <Link href={ROUTES.LOGIN} className={styles.footerLink}>
            Log in
          </Link>
        </div>
      </div>
    </div>
  );
}
