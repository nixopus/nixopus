import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { DEPLOY } from '@/redux/api-conf';
import {
  Application,
  ApplicationLogsResponse,
  CreateApplicationRequest,
  ReDeployApplicationRequest,
  UpdateDeploymentRequest,
  ApplicationDeployment
} from '@/redux/types/applications';

export const deployApi = createApi({
  reducerPath: 'deployApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Applications', 'Deploy'],
  endpoints: (builder) => ({
    getApplications: builder.query<
      { applications: Application[]; total_count: number },
      { page: number; limit: number }
    >({
      query: ({ page, limit }) => ({
        url: `${DEPLOY.GET_APPLICATIONS}?page=${page}&page_size=${limit}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: {
        data: { applications: Application[]; total_count: number };
      }) => {
        return response.data;
      }
    }),
    createDeployment: builder.mutation<Application, CreateApplicationRequest>({
      query: (data) => ({
        url: DEPLOY.CREATE_APPLICATION,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: Application }) => {
        return response.data;
      }
    }),
    updateDeployment: builder.mutation<Application, UpdateDeploymentRequest>({
      query: (data) => ({
        url: DEPLOY.UPDATE_APPLICATION,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: Application }) => {
        return response.data;
      }
    }),
    redeployApplication: builder.mutation<Application, ReDeployApplicationRequest>({
      query: (data) => ({
        url: `${DEPLOY.REDEPLOY_APPLICATION}`,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: Application }) => {
        return response.data;
      }
    }),
    getApplicationDeploymentById: builder.query<Application, { id: string }>({
      query: ({ id }) => ({
        url: `${DEPLOY.DEPLOYMENT}/${id}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: Application }) => {
        return response.data;
      }
    }),
    getApplicationById: builder.query<Application, { id: string }>({
      query: ({ id }) => ({
        url: `${DEPLOY.GET_APPLICATION}?id=${id}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: Application }) => {
        return response.data;
      }
    }),
    deleteApplication: builder.mutation<null, { id: string }>({
      query: ({ id }) => ({
        url: `${DEPLOY.DELETE_APPLICATION}`,
        body: { id },
        method: 'DELETE'
      }),
      invalidatesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    rollbackApplication: builder.mutation<null, { id: string }>({
      query: ({ id }) => ({
        url: `${DEPLOY.ROLLBACK_APPLICATION}`,
        body: { id },
        method: 'POST'
      }),
      invalidatesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    restartApplication: builder.mutation<null, { id: string }>({
      query: ({ id }) => ({
        url: `${DEPLOY.RESTART_APPLICATION}`,
        body: { id },
        method: 'POST'
      }),
      invalidatesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    getApplicationLogs: builder.query<
      ApplicationLogsResponse,
      {
        id: string;
        page: number;
        page_size: number;
        level?: string;
        search_term?: string;
        start_time?: string;
        end_time?: string;
      }
    >({
      query: ({ id, page, page_size, level, search_term, start_time, end_time }) => ({
        url: DEPLOY.GET_APPLICATION_LOGS.replace('{application_id}', id),
        method: 'GET',
        params: {
          page,
          page_size,
          level,
          search_term,
          start_time,
          end_time
        }
      }),
      providesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: ApplicationLogsResponse }) => {
        return response.data;
      }
    }),
    getDeploymentLogs: builder.query<
      ApplicationLogsResponse,
      {
        id: string;
        page: number;
        page_size: number;
        search_term?: string;
      }
    >({
      query: ({ id, page, page_size, search_term }) => ({
        url: DEPLOY.GET_DEPLOYMENT_LOGS.replace('{deployment_id}', id),
        method: 'GET',
        params: {
          page,
          page_size,
          search_term
        }
      }),
      transformResponse: (response: { data: ApplicationLogsResponse }) => response.data,
      providesTags: (result, error, arg) => [{ type: 'Deploy' as const, id: arg.id }]
    }),
    getApplicationDeployments: builder.query<
      { deployments: ApplicationDeployment[]; total_count: number },
      {
        id: string;
        page: number;
        limit: number;
      }
    >({
      query: ({ id, page, limit }) => ({
        url: `${DEPLOY.GET_APPLICATION_DEPLOYMENTS}?id=${id}&page=${page}&limit=${limit}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: {
        data: { deployments: ApplicationDeployment[]; total_count: number };
      }) => {
        return response.data;
      }
    }),
    updateApplicationLabels: builder.mutation<string[], { id: string; labels: string[] }>({
      query: ({ id, labels }) => ({
        url: `${DEPLOY.UPDATE_APPLICATION_LABELS}?id=${id}`,
        method: 'PUT',
        body: { labels }
      }),
      invalidatesTags: [{ type: 'Deploy', id: 'LIST' }],
      transformResponse: (response: { data: string[] }) => {
        return response.data;
      }
    })
  })
});

export const {
  useGetApplicationsQuery,
  useCreateDeploymentMutation,
  useGetApplicationByIdQuery,
  useUpdateDeploymentMutation,
  useRedeployApplicationMutation,
  useGetApplicationDeploymentByIdQuery,
  useDeleteApplicationMutation,
  useRollbackApplicationMutation,
  useRestartApplicationMutation,
  useGetApplicationLogsQuery,
  useGetDeploymentLogsQuery,
  useGetApplicationDeploymentsQuery,
  useUpdateApplicationLabelsMutation
} = deployApi;
