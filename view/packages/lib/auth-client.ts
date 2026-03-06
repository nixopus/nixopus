import { createAuthClient } from 'better-auth/react';
import { emailOTPClient } from 'better-auth/client/plugins';
import { passkeyClient } from '@better-auth/passkey/client';

const getBaseURL = () => {
  if (typeof window !== 'undefined') {
    return window.location.origin;
  }
  return process.env.AUTH_SERVICE_URL || 'http://localhost:9090/api/auth';
};

export const authClient = createAuthClient({
  baseURL: getBaseURL(),
  plugins: [emailOTPClient(), passkeyClient()]
});
