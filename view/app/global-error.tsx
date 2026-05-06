'use client';

import { useEffect } from 'react';
import { Geist } from 'next/font/google';

// Re-declare font here — root layout is unavailable when this renders
const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin']
});

interface GlobalErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function GlobalError({ error, reset }: GlobalErrorProps) {
  useEffect(() => {
    console.error('[Global Error Boundary]', error);
  }, [error]);

  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} antialiased`}
        suppressHydrationWarning
        style={{ margin: 0, fontFamily: 'var(--font-geist-sans, sans-serif)' }}
      >
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '100svh',
            padding: '24px',
            textAlign: 'center',
            gap: '24px'
          }}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="48"
            height="48"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            style={{ color: '#ef4444' }}
          >
            <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" />
            <path d="M12 9v4" />
            <path d="M12 17h.01" />
          </svg>
          <h1 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0 }}>Something went wrong</h1>
          <p style={{ color: '#6b7280', maxWidth: '400px', margin: 0 }}>
            A critical error occurred. Please try reloading the page.
          </p>
          {error.digest && (
            <p style={{ fontSize: '0.75rem', fontFamily: 'monospace', color: '#9ca3af' }}>
              Error code: {error.digest}
            </p>
          )}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: '8px',
              width: '100%',
              maxWidth: '280px'
            }}
          >
            <button
              onClick={reset}
              style={{
                padding: '10px 16px',
                borderRadius: '6px',
                border: 'none',
                background: '#111827',
                color: '#fff',
                cursor: 'pointer',
                fontSize: '0.875rem',
                fontWeight: 500
              }}
            >
              Try again
            </button>
            <button
              onClick={() => {
                window.location.href = '/chats';
              }}
              style={{
                padding: '10px 16px',
                borderRadius: '6px',
                border: '1px solid #e5e7eb',
                background: 'transparent',
                cursor: 'pointer',
                fontSize: '0.875rem',
                fontWeight: 500
              }}
            >
              Go to Dashboard
            </button>
          </div>
        </div>
      </body>
    </html>
  );
}
