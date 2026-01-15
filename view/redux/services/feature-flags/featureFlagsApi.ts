import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { GetFeatureFlagsResponse, UpdateFeatureFlagRequest } from '@/packages/types/feature-flags';
import { FEATURE_FLAGS } from '@/redux/api-conf';

export const FeatureFlagsApi = createApi({
  reducerPath: 'FeatureFlagsApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['FeatureFlags'],
  endpoints: (builder) => ({
    getAllFeatureFlags: builder.query<GetFeatureFlagsResponse, void>({
      query: () => ({
        url: FEATURE_FLAGS.GET_FEATURE_FLAGS,
        method: 'GET'
      }),
      providesTags: [{ type: 'FeatureFlags', id: 'LIST' }],
      transformResponse: (response: { data: GetFeatureFlagsResponse }) => {
        return response.data;
      }
    }),
    updateFeatureFlag: builder.mutation<null, UpdateFeatureFlagRequest>({
      query: (data) => ({
        url: FEATURE_FLAGS.UPDATE_FEATURE_FLAG,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'FeatureFlags', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    checkFeatureEnabled: builder.query<{ is_enabled: boolean }, string>({
      query: (featureName) => ({
        url: FEATURE_FLAGS.CHECK_FEATURE_ENABLED,
        method: 'GET',
        params: { feature_name: featureName }
      }),
      providesTags: [{ type: 'FeatureFlags', id: 'LIST' }],
      transformResponse: (response: { data: { is_enabled: boolean } }) => {
        return response.data;
      }
    })
  })
});

export const {
  useGetAllFeatureFlagsQuery,
  useUpdateFeatureFlagMutation,
  useCheckFeatureEnabledQuery
} = FeatureFlagsApi;
