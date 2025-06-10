import { AUTHURLS } from '@/redux/api-conf';
import { baseQueryWithReauth } from '@/redux/base-query';
import {
  AuthResponse,
  LoginPayload,
  RefreshTokenPayload,
  TwoFactorLoginPayload,
  TwoFactorSetupResponse,
  User
} from '@/redux/types/user';
import { createApi } from '@reduxjs/toolkit/query/react';

export const authApi = createApi({
  reducerPath: 'authApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Authentication'],
  endpoints: (builder) => ({
    registerUser: builder.mutation<AuthResponse, { email: string; password: string }>({
      query(credentials) {
        return {
          url: AUTHURLS.USER_REGISTER,
          method: 'POST',
          body: {
            ...credentials,
            type: 'admin',
            username: credentials.email.split('@')[0],
            organization: ''
          }
        };
      },
      transformResponse: (response: { data: AuthResponse }) => {
        return { ...response.data };
      },
      invalidatesTags: [{ type: 'Authentication', id: 'LIST' }]
    }),
    loginUser: builder.mutation<AuthResponse, LoginPayload>({
      query(credentials) {
        return {
          url: AUTHURLS.USER_LOGIN,
          method: 'POST',
          body: credentials
        };
      },
      transformResponse: (response: { data: AuthResponse }) => {
        return { ...response.data };
      },
      invalidatesTags: [{ type: 'Authentication', id: 'LIST' }]
    }),
    logout: builder.mutation<void, { refresh_token: string }>({
      query({ refresh_token }) {
        return {
          url: AUTHURLS.LOGOUT,
          method: 'POST',
          body: { refresh_token }
        };
      },
      invalidatesTags: [{ type: 'Authentication', id: 'LIST' }]
    }),
    getUserDetails: builder.query<User, void>({
      query: () => ({
        url: AUTHURLS.USER_DETAILS,
        method: 'GET'
      }),
      providesTags: [{ type: 'Authentication', id: 'LIST' }],
      transformResponse: (response: { data: User }) => {
        return { ...response.data };
      }
    }),
    refreshToken: builder.mutation<AuthResponse, RefreshTokenPayload>({
      query: (payload) => ({
        url: AUTHURLS.REFRESH_TOKEN,
        method: 'POST',
        body: payload
      }),
      transformResponse: (response: { data: AuthResponse }) => {
        return { ...response.data };
      }
    }),
    resetPassword: builder.mutation<void, { token: string; password: string }>({
      query({ token, password }) {
        return {
          url: `${AUTHURLS.RESET_PASSWORD}?token=${token}`,
          method: 'POST',
          body: { password }
        };
      }
    }),
    verifyEmail: builder.mutation<void, { token: string }>({
      query({ token }) {
        return {
          url: `${AUTHURLS.VERIFY_EMAIL}?token=${token}`,
          method: 'GET'
        };
      }
    }),
    sendVerificationEmail: builder.mutation<void, void>({
      query() {
        return {
          url: AUTHURLS.SEND_VERIFICATION,
          method: 'POST'
        };
      }
    }),
    setupTwoFactor: builder.mutation<TwoFactorSetupResponse, void>({
      query: () => ({
        url: AUTHURLS.SETUP_TWO_FACTOR,
        method: 'POST'
      }),
      transformResponse: (response: { data: TwoFactorSetupResponse }) => {
        return { ...response.data };
      },
      invalidatesTags: [{ type: 'Authentication', id: 'LIST' }]
    }),
    verifyTwoFactor: builder.mutation<void, { code: string }>({
      query: (body) => ({
        url: AUTHURLS.VERIFY_TWO_FACTOR,
        method: 'POST',
        body
      }),
      invalidatesTags: [{ type: 'Authentication', id: 'LIST' }]
    }),
    disableTwoFactor: builder.mutation<void, void>({
      query: () => ({
        url: AUTHURLS.DISABLE_TWO_FACTOR,
        method: 'POST'
      }),
      invalidatesTags: [{ type: 'Authentication', id: 'LIST' }]
    }),
    twoFactorLogin: builder.mutation<AuthResponse, TwoFactorLoginPayload>({
      query: (credentials) => ({
        url: AUTHURLS.TWO_FACTOR_LOGIN,
        method: 'POST',
        body: credentials
      }),
      transformResponse: (response: { data: AuthResponse }) => {
        return { ...response.data };
      },
      invalidatesTags: [{ type: 'Authentication', id: 'LIST' }]
    }),
    isAdminRegistered: builder.query<boolean, void>({
      query: () => ({
        url: AUTHURLS.IS_ADMIN_REGISTERED,
        method: 'GET'
      }),
      transformResponse: (response: { data: { admin_registered: boolean } }) => {
        return response.data.admin_registered;
      }
    })
  })
});

export const {
  useRegisterUserMutation,
  useLoginUserMutation,
  useLogoutMutation,
  useGetUserDetailsQuery,
  useRefreshTokenMutation,
  useResetPasswordMutation,
  useVerifyEmailMutation,
  useSendVerificationEmailMutation,
  useSetupTwoFactorMutation,
  useVerifyTwoFactorMutation,
  useDisableTwoFactorMutation,
  useTwoFactorLoginMutation,
  useIsAdminRegisteredQuery
} = authApi;
