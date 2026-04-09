import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { SERVERURLS } from '@/redux/api-conf';
import type {
  GetServersResponse,
  GetServersParams,
  CreateMachineRequest,
  CreateMachineResponse,
  MachineVerifyResponse,
  ProvisionMachineRequest,
  ProvisionStatusResponse,
  MachineSshStatusResponse,
  DeleteMachineResponse
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
      })
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
  useDeleteMachineMutation,
  useGetMachineSshStatusQuery,
  useLazyGetMachineSshStatusQuery
} = machinesApi;
