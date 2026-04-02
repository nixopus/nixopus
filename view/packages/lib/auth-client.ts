import { createAuthClient } from 'better-auth/react';
import { emailOTPClient, organizationClient, apiKeyClient } from 'better-auth/client/plugins';
import { dodopaymentsClient } from '@dodopayments/better-auth';
import { passkeyClient } from '@better-auth/passkey/client';

const getBaseURL = () => {
  if (typeof window !== 'undefined') {
    return window.location.origin;
  }
  return process.env.AUTH_SERVICE_URL || 'http://localhost:9090/api/auth';
};

const getApiKey = () => {
  if (typeof window !== 'undefined') {
    return (window as any).__NIXOPUS_API_KEY__ || '';
  }
  return process.env.NEXT_PUBLIC_AUTH_API_KEY || '';
};

export const authClient = createAuthClient({
  baseURL: getBaseURL(),
  fetchOptions: {
    onRequest: (ctx: any) => {
      const apiKey = getApiKey();
      if (apiKey) {
        ctx.headers.set('x-api-key', apiKey);
      }
      return ctx;
    }
  },
  plugins: [
    emailOTPClient(),
    organizationClient(),
    dodopaymentsClient(),
    apiKeyClient(),
    passkeyClient()
  ]
});
