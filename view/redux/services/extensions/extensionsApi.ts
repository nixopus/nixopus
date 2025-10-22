import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import {
  Extension,
  ExtensionListParams,
  ExtensionListResponse,
  ExtensionExecution,
  ExtensionCategory
} from '@/redux/types/extension';
import { EXTENSIONURLS } from '@/redux/api-conf';

export const extensionsApi = createApi({
  reducerPath: 'extensionsApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Extensions', 'Extension', 'Execution'],
  endpoints: (builder) => ({
    getExtensions: builder.query<ExtensionListResponse, ExtensionListParams>({
      query: (params) => {
        const searchParams = new URLSearchParams();

        if (params.category) {
          searchParams.append('category', params.category);
        }
        if (params.type) {
          searchParams.append('type', params.type);
        }
        if (params.search) {
          searchParams.append('search', params.search);
        }
        if (params.sort_by) {
          searchParams.append('sort_by', params.sort_by);
        }
        if (params.sort_dir) {
          searchParams.append('sort_dir', params.sort_dir);
        }
        if (params.page) {
          searchParams.append('page', params.page.toString());
        }
        if (params.page_size) {
          searchParams.append('page_size', params.page_size.toString());
        }

        const queryString = searchParams.toString();
        return {
          url: `${EXTENSIONURLS.GET_EXTENSIONS}${queryString ? `?${queryString}` : ''}`,
          method: 'GET'
        };
      },
      providesTags: ['Extensions'],
      transformResponse: (response: ExtensionListResponse) => response
    }),
    getExtensionCategories: builder.query<ExtensionCategory[], void>({
      query: () => ({
        url: EXTENSIONURLS.GET_CATEGORIES,
        method: 'GET'
      }),
      providesTags: ['Extensions']
    }),
    getExtension: builder.query<Extension, { id: string }>({
      query: ({ id }) => ({
        url: EXTENSIONURLS.GET_EXTENSION.replace('{id}', id),
        method: 'GET'
      }),
      providesTags: (result, error, { id }) => [{ type: 'Extension', id }],
      transformResponse: (response: Extension) => response
    }),
    getExtensionByExtensionId: builder.query<Extension, { extensionId: string }>({
      query: ({ extensionId }) => ({
        url: EXTENSIONURLS.GET_EXTENSION_BY_ID.replace('{extension_id}', extensionId),
        method: 'GET'
      }),
      providesTags: (result, error, { extensionId }) => [{ type: 'Extension', id: extensionId }],
      transformResponse: (response: Extension) => response
    }),
    runExtension: builder.mutation<
      ExtensionExecution,
      { extensionId: string; body: FormData | { variables?: Record<string, unknown> } }
    >({
      query: ({ extensionId, body }) => {
        const isFormData = typeof FormData !== 'undefined' && body instanceof FormData;
        return {
          url: EXTENSIONURLS.RUN_EXTENSION.replace('{extension_id}', extensionId),
          method: 'POST',
          body,
          headers: isFormData ? undefined : { 'Content-Type': 'application/json' }
        };
      }
    }),
    forkExtension: builder.mutation<Extension, { extensionId: string; yaml_content?: string }>({
      query: ({ extensionId, ...body }) => ({
        url: EXTENSIONURLS.FORK_EXTENSION.replace('{extension_id}', extensionId),
        method: 'POST',
        body,
        headers: { 'Content-Type': 'application/json' }
      }),
      invalidatesTags: ['Extensions']
    }),
    deleteExtension: builder.mutation<{ status: string }, { id: string }>({
      query: ({ id }) => ({
        url: EXTENSIONURLS.DELETE_EXTENSION.replace('{id}', id),
        method: 'DELETE'
      }),
      invalidatesTags: ['Extensions']
    }),
    cancelExecution: builder.mutation<{ status: string; message: string }, { executionId: string }>(
      {
        query: ({ executionId }) => ({
          url: EXTENSIONURLS.CANCEL_EXECUTION.replace('{execution_id}', executionId),
          method: 'POST'
        })
      }
    ),
    getExecution: builder.query<ExtensionExecution, { executionId: string }>({
      query: ({ executionId }) => ({
        url: EXTENSIONURLS.GET_EXECUTION.replace('{execution_id}', executionId),
        method: 'GET'
      }),
      providesTags: (result, error, { executionId }) => [{ type: 'Execution', id: executionId }],
      transformResponse: (response: ExtensionExecution) => response
    }),
    listExecutions: builder.query<ExtensionExecution[], { extensionId: string }>({
      query: ({ extensionId }) => ({
        url: EXTENSIONURLS.LIST_EXECUTIONS.replace('{extension_id}', extensionId),
        method: 'GET'
      }),
      providesTags: (result, error, { extensionId }) => [{ type: 'Extension', id: extensionId }],
      transformResponse: (response: ExtensionExecution[]) => response
    }),
    getExecutionLogs: builder.query<
      { logs: any[]; next_after: number },
      { executionId: string; afterSeq?: number; limit?: number }
    >({
      query: ({ executionId, afterSeq = 0, limit = 200 }) => ({
        url: `${EXTENSIONURLS.GET_EXECUTION_LOGS.replace('{execution_id}', executionId)}?afterSeq=${afterSeq}&limit=${limit}`,
        method: 'GET'
      }),
      providesTags: (result, error, { executionId }) => [{ type: 'Execution', id: executionId }],
      transformResponse: (response: { logs: any[]; next_after: number }) => response
    })
  })
});

export const {
  useGetExtensionsQuery,
  useGetExtensionCategoriesQuery,
  useGetExtensionQuery,
  useGetExtensionByExtensionIdQuery,
  useRunExtensionMutation,
  useForkExtensionMutation,
  useDeleteExtensionMutation,
  useCancelExecutionMutation,
  useGetExecutionQuery,
  useListExecutionsQuery,
  useGetExecutionLogsQuery
} = extensionsApi;
