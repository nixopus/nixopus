import { createAuthClient } from 'better-auth/react';
// Note: Organization plugin may not be available in current Better Auth version
// We'll use REST API calls through the proxy instead
// If organization plugin becomes available, uncomment below:
// import { organizationClient } from 'better-auth/client/plugins';

const getBaseURL = () => {
  if (typeof window !== 'undefined') {
    // Use the proxy route in Next.js
    return window.location.origin;
  }
  // Server-side: use the auth service URL directly
  return process.env.AUTH_SERVICE_URL || 'http://localhost:9090/api/auth';
};

export const authClient = createAuthClient({
  baseURL: getBaseURL()
  // If organization plugin is available, add it here:
  // plugins: [organizationClient()],
});
