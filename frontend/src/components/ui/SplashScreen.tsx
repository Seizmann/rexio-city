/**
 * SplashScreen — shown during initial auth hydration (isLoading state).
 *
 * Mimics Twitter/X's approach: full-screen dark background with the
 * app icon centered, animating in with a spring pop-in effect.
 * Replaces the old plain spinner.
 */

import styles from './SplashScreen.module.css';

export default function SplashScreen() {
  return (
    <div className={styles.splash} suppressHydrationWarning>
      <div className={styles.iconWrap}>
        <img
          src="/icon.webp"
          alt="Rexio City"
          className={styles.icon}
          draggable={false}
        />
      </div>
    </div>
  );
}
