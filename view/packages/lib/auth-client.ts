import { createAuthClient } from 'better-auth/react';

const getBaseURL = () => {
  if (typeof window !== 'undefined') {
    // Use the proxy route in Next.js
    return window.location.origin;
  }
  // Server-side: use the auth service URL directly
  return process.env.AUTH_SERVICE_URL || 'http://localhost:9090/api/auth';
};

export const authClient = createAuthClient({
  baseURL: getBaseURL(),
});
