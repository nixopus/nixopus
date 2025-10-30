import { SERVER_SETTINGS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { 
  Server, 
  CreateServerRequest, 
  UpdateServerRequest, 
  UpdateServerStatusRequest,
  DeleteServerRequest,
  CreateServerResponse,
  GetServersRequest,
  ServerListResponse
} from '@/redux/types/server';

export const serversApi = createApi({
  reducerPath: 'serversApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Servers'],
  endpoints: (builder) => ({
    getAllServers: builder.query<ServerListResponse, GetServersRequest | void>({
      query: (params) => {
        const searchParams = new URLSearchParams();
        
        if (params) {
          if (params.page) searchParams.append('page', params.page.toString());
          if (params.page_size) searchParams.append('page_size', params.page_size.toString());
          if (params.search) searchParams.append('search', params.search);
          if (params.sort_by) searchParams.append('sort_by', params.sort_by);
          if (params.sort_order) searchParams.append('sort_order', params.sort_order);
        }
        
        const queryString = searchParams.toString();
        
        return {
          url: `${SERVER_SETTINGS.GET_SERVERS}${queryString ? `?${queryString}` : ''}`,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'Servers', id: 'LIST' }],
      transformResponse: (response: { data: ServerListResponse }) => {
        return response.data;
      }
    }),
    createServer: builder.mutation<CreateServerResponse, CreateServerRequest>({
      query: (data) => ({
        url: SERVER_SETTINGS.CREATE_SERVER,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Servers', id: 'LIST' }],
      transformResponse: (response: { data: CreateServerResponse }) => {
        return response.data;
      }
    }),
    updateServer: builder.mutation<null, UpdateServerRequest>({
      query: (data) => ({
        url: SERVER_SETTINGS.UPDATE_SERVER,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'Servers', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    updateServerStatus: builder.mutation<Server, UpdateServerStatusRequest>({
      query: (data) => ({
        url: SERVER_SETTINGS.UPDATE_SERVER_STATUS,
        method: 'PATCH',
        body: data
      }),
      invalidatesTags: [{ type: 'Servers', id: 'LIST' }, { type: 'Servers', id: 'ACTIVE' }],
      transformResponse: (response: { data: Server }) => {
        return response.data;
      }
    }),
    deleteServer: builder.mutation<null, DeleteServerRequest>({
      query: (data) => ({
        url: SERVER_SETTINGS.DELETE_SERVER,
        method: 'DELETE',
        body: data
      }),
      invalidatesTags: [{ type: 'Servers', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    getActiveServer: builder.query<Server | null, void>({
      query: () => ({
        url: `${SERVER_SETTINGS.GET_SERVERS}?page=1&page_size=100`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Servers', id: 'ACTIVE' }],
      transformResponse: (response: { data: ServerListResponse }) => {
        const servers = response.data.servers as Server[];
        return servers.find(server => server.status === 'active') || null;
      }
    })
  })
});

export const {
  useGetAllServersQuery,
  useCreateServerMutation,
  useUpdateServerMutation,
  useUpdateServerStatusMutation,
  useDeleteServerMutation,
  useGetActiveServerQuery
} = serversApi;
