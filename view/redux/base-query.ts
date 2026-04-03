import { getBaseUrl } from '@/redux/conf';
import { RootState } from '@/redux/store';
import {
  BaseQueryFn,
  FetchArgs,
  FetchBaseQueryError,
  fetchBaseQuery,
  retry
} from '@reduxjs/toolkit/query/react';
import { Mutex } from 'async-mutex';
import { authClient } from '@/packages/lib/auth-client';
import { getAdvancedSettings } from '@/packages/utils/advanced-settings';

const mutex = new Mutex();

function getMaxRetries(): number {
  try {
    return getAdvancedSettings().apiRetryAttempts;
  } catch {
    return 1;
  }
}

let currentBaseUrl: string | undefined;

export async function preloadBaseUrl() {
  if (!currentBaseUrl) {
    currentBaseUrl = await getBaseUrl();
  }
}

const sharedPrepareHeaders = async (
  headers: Headers,
  { getState }: { getState: () => unknown }
) => {
  try {
    const state = getState() as RootState;
    let token = state.auth.token;
    if (!token) {
      const session = await authClient.getSession();
      token = session?.data?.session?.token;
    }
    const organizationId =
      state.user.activeOrganization?.id || state.orgs.organizations[0]?.organization.id;

    if (token) {
      headers.set('authorization', `Bearer ${token}`);
    }

    if (organizationId) {
      headers.set('X-Organization-Id', organizationId);
    }

    const advancedSettings = getAdvancedSettings();
    if (advancedSettings.disableApiCache) {
      headers.set('X-Disable-Cache', 'true');
    }
  } catch (error) {
    console.error('Error getting session token:', error);
  }

  return headers;
};

const proxiedBaseQuery = fetchBaseQuery({
  baseUrl: '',
  prepareHeaders: sharedPrepareHeaders,
  credentials: 'include'
});

let apiBaseQuery: ReturnType<typeof fetchBaseQuery> | null = null;

async function getApiBaseQuery() {
  if (apiBaseQuery) return apiBaseQuery;
  if (!currentBaseUrl) {
    currentBaseUrl = await getBaseUrl();
  }
  apiBaseQuery = fetchBaseQuery({
    baseUrl: currentBaseUrl,
    prepareHeaders: sharedPrepareHeaders,
    credentials: 'include'
  });
  return apiBaseQuery;
}

const PROXIED_PREFIXES = ['/api/auth', '/api/credits', '/api/trail', '/api/agent'];

const customBaseQuery: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> = async (
  args,
  api,
  extraOptions
) => {
  const url = typeof args === 'string' ? args : args.url;
  const isProxiedEndpoint = PROXIED_PREFIXES.some((prefix) => url.startsWith(prefix));
  const baseQuery = isProxiedEndpoint ? proxiedBaseQuery : await getApiBaseQuery();

  return baseQuery(args, api, extraOptions);
};

const retryableBaseQuery = retry(customBaseQuery, {
  maxRetries: getMaxRetries(),
  backoff: async (attempt) => {
    const delay = Math.min(1000 * 2 ** attempt, 30000);
    await new Promise((resolve) => setTimeout(resolve, delay));
  }
});

export const baseQueryWithReauth: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  await mutex.waitForUnlock();

  try {
    let result = await retryableBaseQuery(args, api, extraOptions);

    if (result.error) {
      if (result.error.status === 401) {
        console.warn('Unauthorized request, logging out');
        api.dispatch({ type: 'auth/logoutUser' });
      }

      if (result.error.status === 429) {
        console.warn('Rate limit exceeded, waiting before retry');
        await new Promise((resolve) => setTimeout(resolve, 5000));
        result = await retryableBaseQuery(args, api, extraOptions);
      }
    }

    return result;
  } catch (error) {
    console.error('API request failed:', error);
    return {
      error: {
        status: 'FETCH_ERROR',
        error: 'Network request failed'
      }
    };
  }
};

export { customBaseQuery as baseQuery };
