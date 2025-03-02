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

export const setTokensToStorage = (token: string, refreshToken?: string, expiresIn?: number) => {
  if (token) {
    try {
      localStorage.setItem('token', token);
      localStorage.setItem('lastLogin', new Date().toISOString());

      if (refreshToken) {
        localStorage.setItem('refreshToken', refreshToken);
      }

      if (expiresIn) {
        const expiryTime = Date.now() + expiresIn * 1000;
        localStorage.setItem('tokenExpiry', expiryTime.toString());
      }
    } catch (error) {
      console.error('Failed to save tokens to localStorage:', error);
    }
  }
};

const LOGOUT = 'auth/logout';
const SET_CREDENTIALS = 'auth/setCredentials';

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
    }
  });

  let result = await baseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 401) {
    console.log('Token expired, attempting refresh');

    const refreshToken = (api.getState() as RootState).auth.refreshToken;

    if (!refreshToken) {
      api.dispatch({ type: LOGOUT });
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
      setTokensToStorage(
        refreshData.access_token,
        refreshData.refresh_token,
        refreshData.expires_in
      );
      api.dispatch({
        type: SET_CREDENTIALS,
        payload: {
          user: null,
          token: refreshData.access_token,
          refreshToken: refreshData.refresh_token,
          expiresIn: refreshData.expires_in
        }
      });

      result = await baseQuery(args, api, extraOptions);
    } else {
      api.dispatch({ type: LOGOUT });
    }
  }

  return result;
};
