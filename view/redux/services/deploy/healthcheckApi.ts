import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { HEALTHCHECKURLS } from '@/redux/api-conf';
import {
  HealthCheck,
  HealthCheckResult,
  HealthCheckStats,
  CreateHealthCheckRequest,
  UpdateHealthCheckRequest,
  ToggleHealthCheckRequest,
  GetHealthCheckResultsRequest,
  GetHealthCheckStatsRequest
} from '@/redux/types/healthcheck';

export const healthcheckApi = createApi({
  reducerPath: 'healthcheckApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['HealthCheck', 'HealthCheckResult', 'HealthCheckStats'],
  endpoints: (builder) => ({
    getHealthCheck: builder.query<HealthCheck, string>({
      query: (applicationId) => ({
        url: `${HEALTHCHECKURLS.GET_HEALTH_CHECK}?application_id=${applicationId}`,
        method: 'GET'
      }),
      providesTags: (result, error, applicationId) => [{ type: 'HealthCheck', id: applicationId }],
      transformResponse: (response: { data: HealthCheck }) => {
        return response.data;
      }
    }),
    createHealthCheck: builder.mutation<HealthCheck, CreateHealthCheckRequest>({
      query: (data) => ({
        url: HEALTHCHECKURLS.CREATE_HEALTH_CHECK,
        method: 'POST',
        body: data
      }),
      invalidatesTags: (result, error, { application_id }) => [
        { type: 'HealthCheck', id: application_id },
        { type: 'HealthCheckStats', id: application_id }
      ],
      transformResponse: (response: { data: HealthCheck }) => {
        return response.data;
      }
    }),
    updateHealthCheck: builder.mutation<HealthCheck, UpdateHealthCheckRequest>({
      query: (data) => ({
        url: HEALTHCHECKURLS.UPDATE_HEALTH_CHECK,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: (result, error, { application_id }) => [
        { type: 'HealthCheck', id: application_id },
        { type: 'HealthCheckStats', id: application_id }
      ],
      transformResponse: (response: { data: HealthCheck }) => {
        return response.data;
      }
    }),
    deleteHealthCheck: builder.mutation<void, string>({
      query: (applicationId) => ({
        url: `${HEALTHCHECKURLS.DELETE_HEALTH_CHECK}?application_id=${applicationId}`,
        method: 'DELETE'
      }),
      invalidatesTags: (result, error, applicationId) => [
        { type: 'HealthCheck', id: applicationId },
        { type: 'HealthCheckStats', id: applicationId },
        { type: 'HealthCheckResult', id: applicationId }
      ]
    }),
    toggleHealthCheck: builder.mutation<HealthCheck, ToggleHealthCheckRequest>({
      query: (data) => ({
        url: HEALTHCHECKURLS.TOGGLE_HEALTH_CHECK,
        method: 'PATCH',
        body: data
      }),
      invalidatesTags: (result, error, { application_id }) => [
        { type: 'HealthCheck', id: application_id },
        { type: 'HealthCheckStats', id: application_id }
      ],
      transformResponse: (response: { data: HealthCheck }) => {
        return response.data;
      }
    }),
    getHealthCheckResults: builder.query<HealthCheckResult[], GetHealthCheckResultsRequest>({
      query: ({ application_id, limit, start_time, end_time }) => {
        const params = new URLSearchParams();
        params.append('application_id', application_id);
        if (limit) params.append('limit', limit.toString());
        if (start_time) params.append('start_time', start_time);
        if (end_time) params.append('end_time', end_time);
        return {
          url: `${HEALTHCHECKURLS.GET_HEALTH_CHECK_RESULTS}?${params.toString()}`,
          method: 'GET'
        };
      },
      providesTags: (result, error, { application_id }) => [
        { type: 'HealthCheckResult', id: application_id }
      ],
      transformResponse: (response: { data: HealthCheckResult[] }) => {
        return response.data;
      }
    }),
    getHealthCheckStats: builder.query<HealthCheckStats, GetHealthCheckStatsRequest>({
      query: ({ application_id, period }) => {
        const params = new URLSearchParams();
        params.append('application_id', application_id);
        if (period) params.append('period', period);
        return {
          url: `${HEALTHCHECKURLS.GET_HEALTH_CHECK_STATS}?${params.toString()}`,
          method: 'GET'
        };
      },
      providesTags: (result, error, { application_id }) => [
        { type: 'HealthCheckStats', id: application_id }
      ],
      transformResponse: (response: { data: HealthCheckStats }) => {
        return response.data;
      }
    })
  })
});

export const {
  useGetHealthCheckQuery,
  useCreateHealthCheckMutation,
  useUpdateHealthCheckMutation,
  useDeleteHealthCheckMutation,
  useToggleHealthCheckMutation,
  useGetHealthCheckResultsQuery,
  useGetHealthCheckStatsQuery
} = healthcheckApi;
