import styles from './Skeleton.module.css';

interface SkeletonProps {
  /** Width in px or CSS value. */
  width?: number | string;
  /** Height in px or CSS value. */
  height?: number | string;
  /** Make the skeleton a circle (for avatars). */
  circle?: boolean;
  className?: string;
}

/** Generic skeleton loader block with shimmer animation. */
export function Skeleton({ width, height, circle, className }: SkeletonProps) {
  return (
    <div
      className={`${styles.skeleton} ${circle ? styles.circle : ''} ${className || ''}`}
      style={{
        width: typeof width === 'number' ? `${width}px` : width,
        height: typeof height === 'number' ? `${height}px` : height,
      }}
      aria-hidden="true"
    />
  );
}

/** Skeleton for a post card — matches the PostCard layout. */
export function PostCardSkeleton() {
  return (
    <div className={styles.postSkeleton} aria-hidden="true">
      <div className={styles.postSkeletonHeader}>
        <Skeleton width={40} height={40} circle className={styles.postSkeletonAvatar} />
        <div className={styles.postSkeletonMeta}>
          <Skeleton className={`${styles.skeleton} ${styles.postSkeletonName}`} />
          <Skeleton className={`${styles.skeleton} ${styles.postSkeletonUsername}`} />
        </div>
      </div>
      <div className={styles.postSkeletonBody}>
        <Skeleton className={`${styles.skeleton} ${styles.text} ${styles.textFull}`} />
        <Skeleton className={`${styles.skeleton} ${styles.text} ${styles.textMedium}`} />
        <Skeleton className={`${styles.skeleton} ${styles.text} ${styles.textShort}`} />
      </div>
      <div className={styles.postSkeletonActions}>
        <Skeleton className={`${styles.skeleton} ${styles.postSkeletonAction}`} />
        <Skeleton className={`${styles.skeleton} ${styles.postSkeletonAction}`} />
        <Skeleton className={`${styles.skeleton} ${styles.postSkeletonAction}`} />
        <Skeleton className={`${styles.skeleton} ${styles.postSkeletonAction}`} />
      </div>
    </div>
  );
}
