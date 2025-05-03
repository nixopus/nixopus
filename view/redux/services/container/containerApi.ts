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

export const containerApi = createApi({
  reducerPath: 'containerApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Container'],
  endpoints: (builder) => ({
    getContainers: builder.query<Container[], void>({
      query: () => ({
        url: CONTAINERURLS.GET_CONTAINERS,
        method: 'GET'
      }),
      providesTags: [{ type: 'Container', id: 'LIST' }],
      transformResponse: (response: { data: Container[] }) => {
        return response.data;
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
      query: ({ containerId, tail = 100 }) => ({
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
    })
  })
});

export const {
  useGetContainersQuery,
  useGetContainerQuery,
  useStartContainerMutation,
  useStopContainerMutation,
  useRemoveContainerMutation,
  useGetContainerLogsQuery
} = containerApi;
