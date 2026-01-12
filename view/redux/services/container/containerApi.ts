import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { CONTAINERURLS } from '@/redux/api-conf';

export interface Container {
  id: string;
  name: string;
  image: string;
  status: string;
  state: string;
  created: string;
  command: string;
  ip_address: string;
  labels?: { [key: string]: string };
  ports: {
    private_port: number;
    public_port: number;
    type: string;
  }[];
  host_config: {
    memory: number;
    memory_swap: number;
    cpu_shares: number;
  };
}

export interface ContainerGroup {
  application_id: string;
  application_name: string;
  containers: Container[];
}

export interface UpdateContainerResourcesRequest {
  containerId: string;
  memory: number; // Memory limit in bytes (0 = unlimited)
  memory_swap: number; // Total memory limit (memory + swap) in bytes (0 = unlimited, -1 = unlimited swap)
  cpu_shares: number; // CPU shares (relative weight)
}

export interface UpdateContainerResourcesResponse {
  container_id: string;
  memory: number;
  memory_swap: number;
  cpu_shares: number;
  warnings?: string[];
}

export type ContainerListParams = {
  page: number;
  page_size: number;
  search?: string;
  sort_by?: 'name' | 'status';
  sort_order?: 'asc' | 'desc';
};

export const containerApi = createApi({
  reducerPath: 'containerApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Container'],
  endpoints: (builder) => ({
    getContainers: builder.query<
      {
        containers: Container[];
        groups?: ContainerGroup[];
        ungrouped?: Container[];
        total_count: number;
        group_count?: number;
        page: number;
        page_size: number;
      },
      ContainerListParams
    >({
      query: ({ page, page_size, search, sort_by, sort_order }) => ({
        url: CONTAINERURLS.GET_CONTAINERS,
        method: 'GET',
        params: { page, page_size, search, sort_by, sort_order }
      }),
      providesTags: [{ type: 'Container', id: 'LIST' }],
      transformResponse: (response: {
        data: {
          containers?: Container[];
          groups?: ContainerGroup[];
          ungrouped?: Container[];
          total_count: number;
          group_count?: number;
          page: number;
          page_size: number;
        };
      }) => {
        return {
          containers: response.data.containers ?? [],
          groups: response.data.groups,
          ungrouped: response.data.ungrouped,
          total_count: response.data.total_count,
          group_count: response.data.group_count,
          page: response.data.page,
          page_size: response.data.page_size
        };
      }
    }),
    getContainer: builder.query<Container, string>({
      query: (containerId) => ({
        url: CONTAINERURLS.GET_CONTAINER.replace('{container_id}', containerId),
        method: 'GET'
      }),
      providesTags: (result, error, id) => [{ type: 'Container', id }],
      transformResponse: (response: { data: Container }) => {
        return response.data;
      }
    }),
    startContainer: builder.mutation<void, string>({
      query: (containerId) => ({
        url: CONTAINERURLS.START_CONTAINER.replace('{container_id}', containerId),
        method: 'POST'
      }),
      invalidatesTags: (result, error, id) => [
        { type: 'Container', id },
        { type: 'Container', id: 'LIST' }
      ]
    }),
    stopContainer: builder.mutation<void, string>({
      query: (containerId) => ({
        url: CONTAINERURLS.STOP_CONTAINER.replace('{container_id}', containerId),
        method: 'POST'
      }),
      invalidatesTags: (result, error, id) => [
        { type: 'Container', id },
        { type: 'Container', id: 'LIST' }
      ]
    }),
    removeContainer: builder.mutation<void, string>({
      query: (containerId) => ({
        url: CONTAINERURLS.REMOVE_CONTAINER.replace('{container_id}', containerId),
        method: 'DELETE'
      }),
      invalidatesTags: (result, error, id) => [
        { type: 'Container', id },
        { type: 'Container', id: 'LIST' }
      ]
    }),
    getContainerLogs: builder.query<string, { containerId: string; tail?: number }>({
      query: ({ containerId, tail }) => ({
        url: CONTAINERURLS.GET_CONTAINER_LOGS.replace('{container_id}', containerId),
        method: 'POST',
        body: {
          id: containerId,
          tail: tail,
          stdout: true,
          stderr: true
        }
      }),
      transformResponse: (response: { data: string }) => {
        return response.data;
      }
    }),
    updateContainerResources: builder.mutation<
      UpdateContainerResourcesResponse,
      UpdateContainerResourcesRequest
    >({
      query: ({ containerId, memory, memory_swap, cpu_shares }) => ({
        url: CONTAINERURLS.UPDATE_CONTAINER_RESOURCES.replace('{container_id}', containerId),
        method: 'PUT',
        body: {
          memory,
          memory_swap,
          cpu_shares
        }
      }),
      invalidatesTags: (result, error, { containerId }) => [
        { type: 'Container', id: containerId },
        { type: 'Container', id: 'LIST' }
      ],
      transformResponse: (response: { data: UpdateContainerResourcesResponse }) => {
        return response.data;
      }
    })
  })
});

export const {
  useGetContainersQuery,
  useGetContainerQuery,
  useStartContainerMutation,
  useStopContainerMutation,
  useRemoveContainerMutation,
  useGetContainerLogsQuery,
  useUpdateContainerResourcesMutation
} = containerApi;
