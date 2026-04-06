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
  server_id?: string;
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
  server_id?: string;
};

export type ContainerIdArg = string | { containerId: string; server_id?: string };

function resolveContainerIdArg(arg: ContainerIdArg): { containerId: string; server_id?: string } {
  return typeof arg === 'string' ? { containerId: arg } : arg;
}

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
      query: ({ page, page_size, search, sort_by, sort_order, server_id }) => ({
        url: CONTAINERURLS.GET_CONTAINERS,
        method: 'GET',
        params: {
          page,
          page_size,
          search,
          sort_by,
          sort_order,
          ...(server_id ? { server_id } : {})
        }
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
    getContainer: builder.query<Container, ContainerIdArg>({
      query: (arg) => {
        const { containerId, server_id } = resolveContainerIdArg(arg);
        return {
          url: `${CONTAINERURLS.GET_CONTAINER.replace('{container_id}', containerId)}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
          method: 'GET'
        };
      },
      providesTags: (result, error, arg) => [
        { type: 'Container', id: resolveContainerIdArg(arg).containerId }
      ],
      transformResponse: (response: { data: Container }) => {
        return response.data;
      }
    }),
    startContainer: builder.mutation<void, ContainerIdArg>({
      query: (arg) => {
        const { containerId, server_id } = resolveContainerIdArg(arg);
        return {
          url: `${CONTAINERURLS.START_CONTAINER.replace('{container_id}', containerId)}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
          method: 'POST'
        };
      },
      invalidatesTags: (result, error, arg) => {
        const { containerId } = resolveContainerIdArg(arg);
        return [
          { type: 'Container', id: containerId },
          { type: 'Container', id: 'LIST' }
        ];
      }
    }),
    stopContainer: builder.mutation<void, ContainerIdArg>({
      query: (arg) => {
        const { containerId, server_id } = resolveContainerIdArg(arg);
        return {
          url: `${CONTAINERURLS.STOP_CONTAINER.replace('{container_id}', containerId)}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
          method: 'POST'
        };
      },
      invalidatesTags: (result, error, arg) => {
        const { containerId } = resolveContainerIdArg(arg);
        return [
          { type: 'Container', id: containerId },
          { type: 'Container', id: 'LIST' }
        ];
      }
    }),
    removeContainer: builder.mutation<void, ContainerIdArg>({
      query: (arg) => {
        const { containerId, server_id } = resolveContainerIdArg(arg);
        return {
          url: `${CONTAINERURLS.REMOVE_CONTAINER.replace('{container_id}', containerId)}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (result, error, arg) => {
        const { containerId } = resolveContainerIdArg(arg);
        return [
          { type: 'Container', id: containerId },
          { type: 'Container', id: 'LIST' }
        ];
      }
    }),
    getContainerLogs: builder.query<
      string,
      { containerId: string; tail?: number; server_id?: string }
    >({
      query: ({ containerId, tail, server_id }) => ({
        url: `${CONTAINERURLS.GET_CONTAINER_LOGS.replace('{container_id}', containerId)}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
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
      query: ({ containerId, memory, memory_swap, cpu_shares, server_id }) => ({
        url: `${CONTAINERURLS.UPDATE_CONTAINER_RESOURCES.replace('{container_id}', containerId)}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
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
  useLazyGetContainerLogsQuery,
  useUpdateContainerResourcesMutation
} = containerApi;
