import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { MACHINEHOSTURLS } from '@/redux/api-conf';
import type {
  MachineBackup,
  TriggerBackupResponse,
  BackupListParams,
  BackupListResponseData,
  BackupListResponse,
  BackupScheduleData,
  BackupScheduleResponse,
} from '@/redux/types/machine-backup';

export type {
  MachineBackup,
  TriggerBackupResponse,
  BackupListParams,
  BackupListResponseData,
  BackupScheduleData,
};

export const machineBackupApi = createApi({
  reducerPath: 'machineBackupApi',
  baseQuery: baseQueryWithReauth,
  keepUnusedDataFor: 120,
  tagTypes: ['MachineBackups', 'BackupSchedule'],
  endpoints: (builder) => ({
    listBackups: builder.query<BackupListResponseData, BackupListParams | void>({
      query: (params) => {
        const p = params || {};
        const searchParams = new URLSearchParams();
        if (p.page) searchParams.set('page', String(p.page));
        if (p.page_size) searchParams.set('page_size', String(p.page_size));
        if (p.search) searchParams.set('search', p.search);
        if (p.sort_by) searchParams.set('sort_by', p.sort_by);
        if (p.sort_order) searchParams.set('sort_order', p.sort_order);
        if (p.status) searchParams.set('status', p.status);
        if (p.server_id) searchParams.set('server_id', p.server_id);
        const qs = searchParams.toString();
        return { url: `${MACHINEHOSTURLS.LIST_BACKUPS}${qs ? `?${qs}` : ''}`, method: 'GET' };
      },
      providesTags: ['MachineBackups'],
      transformResponse: (response: BackupListResponse) => response.data,
    }),
    triggerBackup: builder.mutation<TriggerBackupResponse, { server_id?: string } | undefined>({
      query: (params) => ({
        url: `${MACHINEHOSTURLS.BACKUP}${params?.server_id ? `?server_id=${params.server_id}` : ''}`,
        method: 'POST',
      }),
      invalidatesTags: ['MachineBackups'],
    }),
    getBackupSchedule: builder.query<BackupScheduleData, { server_id?: string } | void>({
      query: (params) => {
        const qs = params?.server_id ? `?server_id=${params.server_id}` : '';
        return { url: `${MACHINEHOSTURLS.GET_BACKUP_SCHEDULE}${qs}`, method: 'GET' };
      },
      providesTags: ['BackupSchedule'],
      transformResponse: (response: BackupScheduleResponse) => response.data,
    }),
    updateBackupSchedule: builder.mutation<BackupScheduleData, BackupScheduleData & { server_id?: string }>({
      query: ({ server_id, ...payload }) => {
        const qs = server_id ? `?server_id=${server_id}` : '';
        return { url: `${MACHINEHOSTURLS.UPDATE_BACKUP_SCHEDULE}${qs}`, method: 'PUT', body: payload };
      },
      invalidatesTags: ['BackupSchedule'],
      transformResponse: (response: BackupScheduleResponse) => response.data,
    }),
  }),
});

export const {
  useListBackupsQuery,
  useTriggerBackupMutation,
  useGetBackupScheduleQuery,
  useUpdateBackupScheduleMutation,
} = machineBackupApi;
