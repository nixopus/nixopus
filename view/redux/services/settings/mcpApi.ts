import { MCP_SETTINGS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import type {
  MCPProvider,
  MCPServer,
  CreateMCPServerRequest,
  UpdateMCPServerRequest,
  DeleteMCPServerRequest,
  TestMCPServerRequest,
  TestMCPServerResult
} from '@/redux/types/mcp';

export const mcpApi = createApi({
  reducerPath: 'mcpApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['MCPServer'],
  endpoints: (builder) => ({
    getMCPCatalog: builder.query<MCPProvider[], void>({
      query: () => ({ url: MCP_SETTINGS.LIST_CATALOG, method: 'GET' }),
      transformResponse: (response: { data: MCPProvider[] }) => response.data ?? []
    }),
    getMCPServers: builder.query<MCPServer[], void>({
      query: () => ({ url: MCP_SETTINGS.LIST_SERVERS, method: 'GET' }),
      providesTags: [{ type: 'MCPServer', id: 'LIST' }],
      transformResponse: (response: { data: MCPServer[] }) => response.data ?? []
    }),
    addMCPServer: builder.mutation<MCPServer, CreateMCPServerRequest>({
      query: (data) => ({ url: MCP_SETTINGS.ADD_SERVER, method: 'POST', body: data }),
      invalidatesTags: [{ type: 'MCPServer', id: 'LIST' }],
      transformResponse: (response: { data: MCPServer }) => response.data
    }),
    updateMCPServer: builder.mutation<MCPServer, UpdateMCPServerRequest>({
      query: (data) => ({
        url: `${MCP_SETTINGS.UPDATE_SERVER}/${data.id}`,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'MCPServer', id: 'LIST' }],
      transformResponse: (response: { data: MCPServer }) => response.data
    }),
    deleteMCPServer: builder.mutation<void, DeleteMCPServerRequest>({
      query: (data) => ({ url: MCP_SETTINGS.DELETE_SERVER, method: 'DELETE', body: data }),
      invalidatesTags: [{ type: 'MCPServer', id: 'LIST' }]
    }),
    testMCPServer: builder.mutation<TestMCPServerResult, TestMCPServerRequest>({
      query: (data) => ({ url: MCP_SETTINGS.TEST_SERVER, method: 'POST', body: data }),
      transformResponse: (response: { data: TestMCPServerResult }) => response.data
    })
  })
});

export const {
  useGetMCPCatalogQuery,
  useGetMCPServersQuery,
  useAddMCPServerMutation,
  useUpdateMCPServerMutation,
  useDeleteMCPServerMutation,
  useTestMCPServerMutation
} = mcpApi;
