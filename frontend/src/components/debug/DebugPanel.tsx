'use client';

import { useState, useEffect } from 'react';

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
  const [show, setShow] = useState(true);

  useEffect(() => {
    // Force initial load
    setTimeout(() => setEntries([...debugEntries]), 0);
    setDebugCallback(setEntries);

    // Poll every 500ms to ensure updates
    const interval = setInterval(() => {
      setEntries([...debugEntries]);
    }, 500);

    return () => {
      clearInterval(interval);
    };
  }, []);

  // Always show the panel (even if empty) so user can see it
  return (
    <div style={{
      position: 'fixed',
      bottom: 70,
      left: 0,
      right: 0,
      maxHeight: '30vh',
      overflow: 'auto',
      backgroundColor: 'rgba(0,0,0,0.95)',
      color: '#0f0',
      fontFamily: 'monospace',
      fontSize: 11,
      zIndex: 9999,
      padding: '8px',
      borderTop: '2px solid #0f0',
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4, fontWeight: 'bold' }}>
        <span>🐛 DEBUG ({entries.length})</span>
        <div style={{ display: 'flex', gap: '4px' }}>
          <button
            onClick={() => setShow(!show)}
            style={{ background: show ? '#333' : '#0f0', color: show ? '#fff' : '#000', border: 'none', padding: '2px 6px', cursor: 'pointer', borderRadius: 4, fontSize: 10 }}
          >
            {show ? 'Hide' : 'Show'}
          </button>
          <button
            onClick={() => { debugEntries = []; setEntries([]); }}
            style={{ background: 'red', color: 'white', border: 'none', padding: '2px 6px', cursor: 'pointer', borderRadius: 4, fontSize: 10 }}
          >
            ✕
          </button>
        </div>
      </div>
      {show && (entries.length === 0 ? (
        <div style={{ color: '#888', padding: '4px 0' }}>No debug events yet...</div>
      ) : (
        entries.map((entry, i) => (
          <div key={i} style={{
            color: entry.type === 'error' ? '#f44' : entry.type === 'upload' ? '#fa0' : '#0f0',
            padding: '2px 0',
            borderBottom: '1px solid #333',
            wordBreak: 'break-all',
          }}>
            {entry.timestamp.slice(11, 23)} [{entry.type.toUpperCase()}] {entry.message}
          </div>
        ))
      ))}
    </div>
  );
}
