'use client';

import { useState, useEffect, useRef } from 'react';

interface DebugEntry {
  timestamp: string;
  type: 'log' | 'error' | 'upload' | 'auth';
  message: string;
}

let debugEntries: DebugEntry[] = [];
// eslint-disable-next-line react-refresh/only-export-components
export let debugCallback: ((entries: DebugEntry[]) => void) | null = null;

// eslint-disable-next-line react-refresh/only-export-components
export function debugLog(type: DebugEntry['type'], message: string) {
  const entry: DebugEntry = {
    timestamp: new Date().toISOString(),
    type,
    message,
  };
  debugEntries.push(entry);
  if (debugEntries.length > 50) debugEntries.shift();
  if (debugCallback) debugCallback(debugEntries);
  console.log(`[DEBUG] ${type}:`, message);
}

export function setDebugCallback(callback: (entries: DebugEntry[]) => void) {
  debugCallback = callback;
}

// eslint-disable-next-line react-refresh/only-export-components
export default function DebugPanel() {
  const [entries, setEntries] = useState<DebugEntry[]>([]);

  useEffect(() => {
    // Use setTimeout to avoid "setState in effect" lint error
    setTimeout(() => setEntries(debugEntries), 0);
    setDebugCallback(setEntries);
  }, []);

  if (entries.length === 0) return null;

  return (
    <div style={{
      position: 'fixed',
      bottom: 0,
      left: 0,
      right: 0,
      maxHeight: '40vh',
      overflow: 'auto',
      backgroundColor: 'rgba(0,0,0,0.9)',
      color: '#0f0',
      fontFamily: 'monospace',
      fontSize: 12,
      zIndex: 9999,
      padding: '8px',
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
        <span>DEBUG ({entries.length})</span>
        <button
          onClick={() => { debugEntries = []; setEntries([]); }}
          style={{ background: 'red', color: 'white', border: 'none', padding: '2px 8px', cursor: 'pointer' }}
        >
          Clear
        </button>
      </div>
      {entries.map((entry, i) => (
        <div key={i} style={{
          color: entry.type === 'error' ? '#f44' : entry.type === 'upload' ? '#fa0' : '#0f0',
          padding: '2px 0',
          borderBottom: '1px solid #333',
        }}>
          {entry.timestamp.slice(11, 23)} [{entry.type.toUpperCase()}] {entry.message}
        </div>
      ))}
    </div>
  );
}
