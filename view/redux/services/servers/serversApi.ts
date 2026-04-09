import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { SERVERURLS } from '@/redux/api-conf';
import type { GetServersResponse, GetServersParams } from '@/redux/types/servers';

export const machinesApi = createApi({
  reducerPath: 'machinesApi',
  baseQuery: baseQueryWithReauth,
  keepUnusedDataFor: 600,
  tagTypes: ['Server'],
  endpoints: (builder) => ({
    getServers: builder.query<GetServersResponse, GetServersParams | void>({
      query: (params) => ({
        url: SERVERURLS.GET_SERVERS,
        method: 'GET',
        params: params ?? undefined
      }),
      providesTags: ['Server'],
      transformResponse: (response: {
        status: string;
        message: string;
        data: GetServersResponse;
      }) => response.data
    })
  })
});

export const { useGetServersQuery, useLazyGetServersQuery } = machinesApi;
