'use client';

import { useState, type FormEvent } from 'react';
import { useAuth } from '@/context/AuthContext';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import { LIMITS } from '@/lib/constants';
import styles from './LandingAuth.module.css';

export default function LandingAuth() {
  const { login, signup } = useAuth();
  const [activeTab, setActiveTab] = useState<'login' | 'signup'>('login');

  // Login form state
  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [loginError, setLoginError] = useState('');
  const [loginLoading, setLoginLoading] = useState(false);

  // Signup form state
  const [signupUsername, setSignupUsername] = useState('');
  const [signupDisplayName, setSignupDisplayName] = useState('');
  const [signupEmail, setSignupEmail] = useState('');
  const [signupPassword, setSignupPassword] = useState('');
  const [signupErrors, setSignupErrors] = useState<Record<string, string>>({});
  const [signupFormError, setSignupFormError] = useState('');
  const [signupLoading, setSignupLoading] = useState(false);

  async function handleLoginSubmit(e: FormEvent) {
    e.preventDefault();
    setLoginError('');

    if (!loginEmail || !loginPassword) {
      setLoginError('Email and password are required.');
      return;
    }

    setLoginLoading(true);
    try {
      await login(loginEmail, loginPassword);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Login failed. Check your credentials.';
      setLoginError(message);
    } finally {
      setLoginLoading(false);
    }
  }

  function validateSignup(): boolean {
    const newErrors: Record<string, string> = {};

    if (signupUsername.length < LIMITS.USERNAME_MIN_CHARS) {
      newErrors.username = `Username must be at least ${LIMITS.USERNAME_MIN_CHARS} characters.`;
    } else if (signupUsername.length > LIMITS.USERNAME_MAX_CHARS) {
      newErrors.username = `Username must be at most ${LIMITS.USERNAME_MAX_CHARS} characters.`;
    } else if (!/^[a-z0-9_]+$/.test(signupUsername)) {
      newErrors.username = 'Username must be lowercase letters, numbers, or underscores.';
    }

    if (signupDisplayName.length > LIMITS.DISPLAY_NAME_MAX_CHARS) {
      newErrors.displayName = `Display name must be at most ${LIMITS.DISPLAY_NAME_MAX_CHARS} characters.`;
    }

    if (!signupEmail.includes('@')) {
      newErrors.email = 'Enter a valid email address.';
    }

    if (signupPassword.length < LIMITS.PASSWORD_MIN_CHARS) {
      newErrors.password = `Password must be at least ${LIMITS.PASSWORD_MIN_CHARS} characters.`;
    }

    setSignupErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }

  async function handleSignupSubmit(e: FormEvent) {
    e.preventDefault();
    setSignupFormError('');

    if (!validateSignup()) return;

    setSignupLoading(true);
    try {
      await signup({
        username: signupUsername,
        email: signupEmail,
        password: signupPassword,
        display_name: signupDisplayName || signupUsername,
      });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Signup failed. Try a different username or email.';
      setSignupFormError(message);
    } finally {
      setSignupLoading(false);
    }
  }

  return (
    <div className={styles.landingContainer}>
      <div className={styles.hero}>
        <h1 className={styles.logo}>RexiO City</h1>
        <p className={styles.tagline}>
          Happening now. Join the conversation today.
        </p>
      </div>

      <div className={styles.card}>
        <div className={styles.tabHeader} role="tablist">
          <button
            className={`${styles.tabBtn} ${activeTab === 'login' ? styles.tabBtnActive : ''}`}
            onClick={() => setActiveTab('login')}
            role="tab"
            aria-selected={activeTab === 'login'}
          >
            Log in
          </button>
          <button
            className={`${styles.tabBtn} ${activeTab === 'signup' ? styles.tabBtnActive : ''}`}
            onClick={() => setActiveTab('signup')}
            role="tab"
            aria-selected={activeTab === 'signup'}
          >
            Sign up
          </button>
        </div>

        {activeTab === 'login' ? (
          <form onSubmit={(e) => void handleLoginSubmit(e)} className={styles.form}>
            {loginError && <div className={styles.formError}>{loginError}</div>}

            <Input
              label="Email"
              type="email"
              placeholder="you@example.com"
              value={loginEmail}
              onChange={(e) => setLoginEmail(e.target.value)}
              autoComplete="email"
              required
            />

            <Input
              label="Password"
              type="password"
              placeholder="Your password"
              value={loginPassword}
              onChange={(e) => setLoginPassword(e.target.value)}
              autoComplete="current-password"
              required
            />

            <Button type="submit" loading={loginLoading} fullWidth>
              Log in
            </Button>
          </form>
        ) : (
          <form onSubmit={(e) => void handleSignupSubmit(e)} className={styles.form}>
            {signupFormError && <div className={styles.formError}>{signupFormError}</div>}

            <Input
              label="Username"
              type="text"
              placeholder="your_username"
              value={signupUsername}
              onChange={(e) => setSignupUsername(e.target.value.toLowerCase())}
              error={signupErrors.username}
              autoComplete="username"
              required
            />

            <Input
              label="Display name"
              type="text"
              placeholder="How others see you (optional)"
              value={signupDisplayName}
              onChange={(e) => setSignupDisplayName(e.target.value)}
              error={signupErrors.displayName}
            />

            <Input
              label="Email"
              type="email"
              placeholder="you@example.com"
              value={signupEmail}
              onChange={(e) => setSignupEmail(e.target.value)}
              error={signupErrors.email}
              autoComplete="email"
              required
            />

            <Input
              label="Password"
              type="password"
              placeholder="At least 8 characters"
              value={signupPassword}
              onChange={(e) => setSignupPassword(e.target.value)}
              error={signupErrors.password}
              autoComplete="new-password"
              required
            />

            <Button type="submit" loading={signupLoading} fullWidth>
              Sign up
            </Button>
          </form>
        )}
      </div>
    </div>
  );
}
