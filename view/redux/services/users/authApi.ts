import { AUTHURLS } from '@/redux/api-conf';
import { baseQueryWithReauth } from '@/redux/base-query';
import { TwoFactorSetupResponse, User } from '@/redux/types/user';
import { createApi } from '@reduxjs/toolkit/query/react';

export const authApi = createApi({
  reducerPath: 'authApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Authentication'],
  endpoints: (builder) => ({
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
  useGetUserDetailsQuery,
  useVerifyEmailMutation,
  useSendVerificationEmailMutation,
  useSetupTwoFactorMutation,
  useVerifyTwoFactorMutation,
  useDisableTwoFactorMutation,
  useIsAdminRegisteredQuery
} = authApi;
