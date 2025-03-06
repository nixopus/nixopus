import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { DEPLOY } from '@/redux/api-conf';
import { Application } from '@/redux/types/applications';

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
    })
  })
});

export const { useGetApplicationsQuery } = deployApi;
