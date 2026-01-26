'use client';
import React from 'react';
import { authClient } from '@/packages/lib/auth-client';

export const BetterAuthProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
  return <>{children}</>;
};
