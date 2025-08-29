import { baseQueryWithReauth } from '@/redux/base-query';
import { AuditLogsResponse } from '../types/audit';
import { createApi } from '@reduxjs/toolkit/query/react';
import { AUDITURLS } from '@/redux/api-conf';

export const auditApi = createApi({
  reducerPath: 'auditApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['AuditLogs'],
  endpoints: (builder) => ({
    getRecentAuditLogs: builder.query<AuditLogsResponse['data'], void>({
      query: () => ({
        url: AUDITURLS.GET_RECENT_AUDIT_LOGS + '?page=1&pageSize=4', // TODO : ALLOW user to view more audit logs
        method: 'GET'
      }),
      transformResponse: (response: AuditLogsResponse) => {
        return response.data;
      },
      providesTags: [{ type: 'AuditLogs', id: 'LIST' }]
    })
  })
});

export const { useGetRecentAuditLogsQuery } = auditApi;
