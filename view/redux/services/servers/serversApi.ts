import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { MACHINEHOSTURLS, SERVERURLS } from '@/redux/api-conf';

const machineHostQuery = (server_id?: string) =>
  server_id ? `?server_id=${encodeURIComponent(server_id)}` : '';
import type {
  GetServersResponse,
  GetServersParams,
  CreateMachineRequest,
  CreateMachineResponse,
  MachineVerifyResponse,
  RenameMachineRequest,
  RenameMachineResponse,
  ProvisionMachineRequest,
  ProvisionStatusResponse,
  MachineSshStatusResponse,
  DeleteMachineResponse,
  MachineState,
  MachineStateResponse,
  MachineActionResponse
} from '@/redux/types/servers';

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
    }),
    createMachine: builder.mutation<CreateMachineResponse, CreateMachineRequest>({
      query: (data) => ({
        url: SERVERURLS.CREATE_MACHINE,
        method: 'POST',
        body: data
      }),
      invalidatesTags: ['Server']
    }),
    verifyMachine: builder.mutation<MachineVerifyResponse, string>({
      query: (id) => ({
        url: `${SERVERURLS.VERIFY_MACHINE}/${id}/verify`,
        method: 'POST'
      }),
      invalidatesTags: ['Server']
    }),
    provisionMachine: builder.mutation<ProvisionStatusResponse, ProvisionMachineRequest>({
      query: (data) => ({
        url: SERVERURLS.PROVISION_MACHINE,
        method: 'POST',
        body: data
      }),
      invalidatesTags: ['Server']
    }),
    getProvisionStatus: builder.query<ProvisionStatusResponse, string>({
      query: (id) => ({
        url: `${SERVERURLS.PROVISION_STATUS}/${id}/status`,
        method: 'GET'
      })
    }),
    renameMachine: builder.mutation<RenameMachineResponse, RenameMachineRequest>({
      query: ({ id, name }) => ({
        url: `${SERVERURLS.VERIFY_MACHINE}/${id}/rename`,
        method: 'PATCH',
        body: { name }
      }),
      invalidatesTags: ['Server']
    }),
    deleteMachine: builder.mutation<DeleteMachineResponse, string>({
      query: (id) => ({
        url: `${SERVERURLS.DELETE_MACHINE}/${id}`,
        method: 'DELETE'
      }),
      invalidatesTags: ['Server']
    }),
    getMachineSshStatus: builder.query<MachineSshStatusResponse, string>({
      query: (id) => ({
        url: `${SERVERURLS.SSH_STATUS}/${id}/ssh/status`,
        method: 'GET'
      }),
      transformResponse: (response: { status: string; data: MachineSshStatusResponse }) =>
        response.data
    }),
    getMachineStatus: builder.query<MachineState, { server_id?: string } | void>({
      query: (params) => ({
        url: `${MACHINEHOSTURLS.STATUS}${machineHostQuery(params?.server_id)}`,
        method: 'GET'
      }),
      transformResponse: (response: MachineStateResponse) => response.data
    }),
    getMachineStats: builder.query<unknown, { server_id?: string } | void>({
      query: (params) => ({
        url: `${MACHINEHOSTURLS.STATS}${machineHostQuery(params?.server_id)}`,
        method: 'GET'
      })
    }),
    execMachine: builder.mutation<unknown, { command: string; server_id?: string }>({
      query: ({ command, server_id }) => ({
        url: `${MACHINEHOSTURLS.EXEC}${machineHostQuery(server_id)}`,
        method: 'POST',
        body: { command }
      })
    }),
    restartMachine: builder.mutation<MachineActionResponse, { server_id?: string } | void>({
      query: (params) => ({
        url: `${MACHINEHOSTURLS.RESTART}${machineHostQuery(params?.server_id)}`,
        method: 'POST'
      })
    }),
    pauseMachine: builder.mutation<MachineActionResponse, { server_id?: string } | void>({
      query: (params) => ({
        url: `${MACHINEHOSTURLS.PAUSE}${machineHostQuery(params?.server_id)}`,
        method: 'POST'
      })
    }),
    resumeMachine: builder.mutation<MachineActionResponse, { server_id?: string } | void>({
      query: (params) => ({
        url: `${MACHINEHOSTURLS.RESUME}${machineHostQuery(params?.server_id)}`,
        method: 'POST'
      })
    }),
    triggerMachineBackup: builder.mutation<unknown, { server_id?: string } | void>({
      query: (params) => ({
        url: `${MACHINEHOSTURLS.BACKUP}${machineHostQuery(params?.server_id)}`,
        method: 'POST'
      })
    }),
    getMachineMetrics: builder.query<
      unknown,
      { server_id?: string; from?: string; to?: string; limit?: number } | void
    >({
      query: (params) => {
        const { server_id, from, to, limit } = params ?? {};
        const p = new URLSearchParams();
        if (from) p.set('from', from);
        if (to) p.set('to', to);
        if (limit != null) p.set('limit', String(limit));
        if (server_id) p.set('server_id', server_id);
        const qs = p.toString();
        return {
          url: `${MACHINEHOSTURLS.METRICS}${qs ? `?${qs}` : ''}`,
          method: 'GET'
        };
      }
    }),
    getMachineEvents: builder.query<
      unknown,
      { server_id?: string; from?: string; to?: string; limit?: number } | void
    >({
      query: (params) => {
        const { server_id, from, to, limit } = params ?? {};
        const p = new URLSearchParams();
        if (from) p.set('from', from);
        if (to) p.set('to', to);
        if (limit != null) p.set('limit', String(limit));
        if (server_id) p.set('server_id', server_id);
        const qs = p.toString();
        return {
          url: `${MACHINEHOSTURLS.EVENTS}${qs ? `?${qs}` : ''}`,
          method: 'GET'
        };
      }
    })
  })
});

export const {
  useGetServersQuery,
  useLazyGetServersQuery,
  useCreateMachineMutation,
  useVerifyMachineMutation,
  useProvisionMachineMutation,
  useGetProvisionStatusQuery,
  useRenameMachineMutation,
  useDeleteMachineMutation,
  useGetMachineSshStatusQuery,
  useLazyGetMachineSshStatusQuery,
  useGetMachineStatusQuery,
  useLazyGetMachineStatusQuery,
  useGetMachineStatsQuery,
  useLazyGetMachineStatsQuery,
  useExecMachineMutation,
  useRestartMachineMutation,
  usePauseMachineMutation,
  useResumeMachineMutation,
  useTriggerMachineBackupMutation,
  useGetMachineMetricsQuery,
  useLazyGetMachineMetricsQuery,
  useGetMachineEventsQuery,
  useLazyGetMachineEventsQuery
} = machinesApi;
