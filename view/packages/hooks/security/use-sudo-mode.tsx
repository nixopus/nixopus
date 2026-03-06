'use client';

import React, { createContext, useContext, useState, useCallback, useRef, useEffect } from 'react';
import { authClient } from '@/packages/lib/auth-client';
import { PasskeyVerificationDialog } from '@/packages/components/security/passkey-verification-dialog';

const SUDO_WINDOW_MS = 15 * 60 * 1000;

interface SudoModeContextValue {
  isElevated: boolean;
  hasPasskeys: boolean;
  requireSudo: (onSuccess: () => void) => void;
}

const SudoModeContext = createContext<SudoModeContextValue>({
  isElevated: false,
  hasPasskeys: false,
  requireSudo: (onSuccess) => onSuccess()
});

export function SudoModeProvider({ children }: { children: React.ReactNode }) {
  const { data: session } = authClient.useSession();
  const [lastVerifiedAt, setLastVerifiedAt] = useState<number | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [hasPasskeys, setHasPasskeys] = useState(false);
  const [passkeysFetched, setPasskeysFetched] = useState(false);
  const pendingCallback = useRef<(() => void) | null>(null);

  const isAuthenticated = !!(session as any)?.session || !!(session as any)?.user;

  useEffect(() => {
    if (!isAuthenticated) return;
    let cancelled = false;
    const loadPasskeys = async () => {
      try {
        const res = await fetch('/api/auth/passkey/list-user-passkeys', {
          method: 'GET',
          credentials: 'include'
        });
        if (!cancelled) {
          if (res.ok) {
            const data = await res.json();
            setHasPasskeys(Array.isArray(data) && data.length > 0);
          } else {
            setHasPasskeys(false);
          }
          setPasskeysFetched(true);
        }
      } catch {
        if (!cancelled) {
          setHasPasskeys(false);
          setPasskeysFetched(true);
        }
      }
    };
    loadPasskeys();
    return () => {
      cancelled = true;
    };
  }, [isAuthenticated]);

  const isElevated = lastVerifiedAt !== null && Date.now() - lastVerifiedAt < SUDO_WINDOW_MS;

  const requireSudo = useCallback(
    (onSuccess: () => void) => {
      if (!passkeysFetched || !hasPasskeys || isElevated) {
        onSuccess();
        return;
      }
      pendingCallback.current = onSuccess;
      setDialogOpen(true);
    },
    [hasPasskeys, passkeysFetched, isElevated]
  );

  const handleVerified = useCallback(() => {
    setLastVerifiedAt(Date.now());
    setDialogOpen(false);
    pendingCallback.current?.();
    pendingCallback.current = null;
  }, []);

  const handleCancel = useCallback(() => {
    setDialogOpen(false);
    pendingCallback.current = null;
  }, []);

  const refreshPasskeys = useCallback(async () => {
    try {
      const res = await fetch('/api/auth/passkey/list-user-passkeys', {
        method: 'GET',
        credentials: 'include'
      });
      if (res.ok) {
        const data = await res.json();
        setHasPasskeys(Array.isArray(data) && data.length > 0);
      } else {
        setHasPasskeys(false);
      }
    } catch {
      setHasPasskeys(false);
    }
  }, []);

  useEffect(() => {
    window.__refreshSudoPasskeys = refreshPasskeys;
    return () => {
      delete window.__refreshSudoPasskeys;
    };
  }, [refreshPasskeys]);

  return (
    <SudoModeContext.Provider value={{ isElevated, hasPasskeys, requireSudo }}>
      {children}
      <PasskeyVerificationDialog
        open={dialogOpen}
        onVerified={handleVerified}
        onCancel={handleCancel}
      />
    </SudoModeContext.Provider>
  );
}

export function useSudoMode() {
  return useContext(SudoModeContext);
}

declare global {
  interface Window {
    __refreshSudoPasskeys?: () => Promise<void>;
  }
}
