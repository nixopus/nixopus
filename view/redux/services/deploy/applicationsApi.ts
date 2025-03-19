import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { DEPLOY } from '@/redux/api-conf';
import {
  Application,
  CreateApplicationRequest,
  ReDeployApplicationRequest,
  UpdateDeploymentRequest
} from '@/redux/types/applications';

export const deployApi = createApi({
  reducerPath: 'deployApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Deploy'],
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
  useRestartApplicationMutation
} = deployApi;
