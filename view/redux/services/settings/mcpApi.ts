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

export interface PaginatedResult<T> {
  items: T[];
  totalCount: number;
  page: number;
  pageSize: number;
}

export interface CatalogQueryParams {
  q?: string;
  sortBy?: string;
  sortDir?: string;
  page?: number;
  limit?: number;
}

export interface ServersQueryParams {
  q?: string;
  sortBy?: string;
  sortDir?: string;
  page?: number;
  limit?: number;
}

function buildQueryString(params: Record<string, string | number | undefined>): string {
  const sp = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== '') sp.set(k, String(v));
  });
  const s = sp.toString();
  return s ? `?${s}` : '';
}

export const mcpApi = createApi({
  reducerPath: 'mcpApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['MCPServer', 'MCPCatalog'],
  endpoints: (builder) => ({
    getMCPCatalog: builder.query<PaginatedResult<MCPProvider>, CatalogQueryParams>({
      query: (params = {}) => ({
        url: `${MCP_SETTINGS.LIST_CATALOG}${buildQueryString({
          q: params.q,
          sort_by: params.sortBy,
          sort_dir: params.sortDir,
          page: params.page,
          limit: params.limit
        })}`,
        method: 'GET'
      }),
      providesTags: ['MCPCatalog'],
      transformResponse: (response: {
        data: { items: MCPProvider[]; total_count: number; page: number; page_size: number };
      }) => ({
        items: response.data?.items ?? [],
        totalCount: response.data?.total_count ?? 0,
        page: response.data?.page ?? 1,
        pageSize: response.data?.page_size ?? 0
      })
    }),

    getMCPServers: builder.query<PaginatedResult<MCPServer>, ServersQueryParams>({
      query: (params = {}) => ({
        url: `${MCP_SETTINGS.LIST_SERVERS}${buildQueryString({
          q: params.q,
          sort_by: params.sortBy,
          sort_dir: params.sortDir,
          page: params.page,
          limit: params.limit
        })}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'MCPServer', id: 'LIST' }],
      transformResponse: (response: {
        data: { items: MCPServer[]; total_count: number; page: number; page_size: number };
      }) => ({
        items: response.data?.items ?? [],
        totalCount: response.data?.total_count ?? 0,
        page: response.data?.page ?? 1,
        pageSize: response.data?.page_size ?? 0
      })
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
