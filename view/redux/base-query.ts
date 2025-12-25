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
import { getAccessToken } from 'supertokens-auth-react/recipe/session';

const mutex = new Mutex();
const MAX_RETRIES = 1;

let currentBaseUrl: string | undefined;

const customBaseQuery: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> = async (
  args,
  api,
  extraOptions
) => {
  if (!currentBaseUrl) {
    currentBaseUrl = await getBaseUrl();
  }

  const baseQuery = fetchBaseQuery({
    baseUrl: currentBaseUrl,
    prepareHeaders: async (headers, { getState, endpoint }) => {
      const token = await getAccessToken();
      const organizationId =
        (getState() as RootState).user.activeOrganization?.id ||
        (getState() as RootState).auth.user?.organization_users?.[0]?.organization_id;

      if (token) {
        headers.set('authorization', `Bearer ${token}`);
      }

      if (organizationId) {
        headers.set('X-Organization-Id', organizationId);
      }
      return headers;
    },
    credentials: 'include'
  });

  return baseQuery(args, api, extraOptions);
};

const retryableBaseQuery = retry(customBaseQuery, {
  maxRetries: MAX_RETRIES,
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
