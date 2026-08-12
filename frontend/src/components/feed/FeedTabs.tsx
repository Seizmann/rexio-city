'use client';

import React from 'react';
import styles from './FeedTabs.module.css';

interface FeedTabsProps {
  activeTab: 'following' | 'foryou';
  onTabChange: (tab: 'following' | 'foryou') => void;
}

export default function FeedTabs({ activeTab, onTabChange }: FeedTabsProps) {
  return (
    <div className={styles.tabs}>
      <button 
        className={`${styles.tab} ${activeTab === 'following' ? styles.active : ''}`}
        onClick={() => onTabChange('following')}
      >
        Following
        {activeTab === 'following' && <div className={styles.indicator} />}
      </button>
      <button 
        className={`${styles.tab} ${activeTab === 'foryou' ? styles.active : ''}`}
        onClick={() => onTabChange('foryou')}
      >
        For You
        {activeTab === 'foryou' && <div className={styles.indicator} />}
      </button>
    </div>
  );
}
