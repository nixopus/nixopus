import { AUTHURLS } from '@/redux/api-conf';
import { BASE_URL } from '@/redux/conf';
import { RootState } from '@/redux/store';
import { AuthResponse } from '@/redux/types/user';
import {
  BaseQueryFn,
  FetchArgs,
  FetchBaseQueryError,
  fetchBaseQuery
} from '@reduxjs/toolkit/query/react';
import { Mutex } from 'async-mutex';
import { setAuthTokens } from '@/lib/auth';

const mutex = new Mutex();

export const baseQueryWithReauth: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  const baseQuery = fetchBaseQuery({
    baseUrl: BASE_URL,
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as RootState).auth.token;
      if (token) {
        headers.set('authorization', `Bearer ${token}`);
      }
      return headers;
    },
    credentials: 'include'
  });

  await mutex.waitForUnlock();
  let result = await baseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 401) {
    if (!mutex.isLocked()) {
      const release = await mutex.acquire();

      try {
        const refreshToken = (api.getState() as RootState).auth.refreshToken;

        if (!refreshToken) {
          api.dispatch({ type: 'auth/logout' });
          return result;
        }

        const refreshResult = await baseQuery(
          {
            url: AUTHURLS.REFRESH_TOKEN,
            method: 'POST',
            body: { refresh_token: refreshToken }
          },
          api,
          extraOptions
        );

        if (refreshResult.data) {
          const refreshData = refreshResult.data as AuthResponse;

          setAuthTokens({
            access_token: refreshData.access_token,
            refresh_token: refreshData.refresh_token,
            expires_in: refreshData.expires_in
          });

          api.dispatch({
            type: 'auth/setCredentials',
            payload: {
              user: null,
              token: refreshData.access_token,
              refreshToken: refreshData.refresh_token,
              expiresIn: refreshData.expires_in
            }
          });

          result = await baseQuery(args, api, extraOptions);
        } else {
          api.dispatch({ type: 'auth/logout' });
        }
      } finally {
        release();
      }
    } else {
      await mutex.waitForUnlock();
      result = await baseQuery(args, api, extraOptions);
    }
  }

  return result;
};