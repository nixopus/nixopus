import { baseQueryWithReauth } from '@/redux/base-query';
import { AuditLogsResponse, ActivitiesResponse, ActivityMessage } from '../types/audit';
import { createApi } from '@reduxjs/toolkit/query/react';
import { AUDITURLS } from '@/redux/api-conf';

interface GetActivitiesParams {
  page?: number;
  pageSize?: number;
  search?: string;
  resource_type?: string;
}

export const auditApi = createApi({
  reducerPath: 'auditApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['AuditLogs', 'Activities'],
  endpoints: (builder) => ({
    getRecentAuditLogs: builder.query<ActivityMessage[], void>({
      query: () => ({
        url: AUDITURLS.GET_RECENT_AUDIT_LOGS + '?pageSize=4',
        method: 'GET'
      }),
      transformResponse: (response: ActivitiesResponse) => {
        return response.data.activities;
      },
      providesTags: [{ type: 'Activities', id: 'LIST' }]
    }),
    getActivities: builder.query<ActivitiesResponse['data'], GetActivitiesParams>({
      query: ({ page = 1, pageSize = 10, search, resource_type }) => {
        const params = new URLSearchParams({
          page: page.toString(),
          pageSize: pageSize.toString()
        });

        if (search) params.append('search', search);
        if (resource_type) params.append('resource_type', resource_type);

        return {
          url: AUDITURLS.GET_RECENT_AUDIT_LOGS + '?' + params.toString(),
          method: 'GET'
        };
      },
      transformResponse: (response: ActivitiesResponse) => {
        return response.data;
      },
      providesTags: [{ type: 'Activities', id: 'LIST' }]
    })
  })
});

export const { useGetRecentAuditLogsQuery, useGetActivitiesQuery } = auditApi;
