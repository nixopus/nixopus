import { DOMAIN_SETTINGS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import {
  Domain,
  RandomSubdomainResponse,
  CreateDomainRequest,
  UpdateDomainRequest,
  DeleteDomainRequest
} from '@/redux/types/domain';

export const domainsApi = createApi({
  reducerPath: 'domainsApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Domains'],
  endpoints: (builder) => ({
    getAllDomains: builder.query<Domain[], void>({
      query: () => ({
        url: DOMAIN_SETTINGS.GET_DOMAINS,
        method: 'GET'
      }),
      providesTags: [{ type: 'Domains', id: 'LIST' }],
      transformResponse: (response: { data: Domain[] }) => {
        return response.data;
      }
    }),
    createDomain: builder.mutation<{ id: string }, CreateDomainRequest>({
      query: (data) => ({
        url: DOMAIN_SETTINGS.ADD_DOMAIN,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Domains', id: 'LIST' }],
      transformResponse: (response: { data: { id: string } }) => {
        return response.data;
      }
    }),
    updateDomain: builder.mutation<null, UpdateDomainRequest>({
      query: (data) => ({
        url: DOMAIN_SETTINGS.UPDATE_DOMAIN,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'Domains', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    deleteDomain: builder.mutation<null, DeleteDomainRequest>({
      query: (data) => ({
        url: DOMAIN_SETTINGS.DELETE_DOMAIN,
        method: 'DELETE',
        body: data
      }),
      invalidatesTags: [{ type: 'Domains', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    generateRandomSubdomain: builder.query<RandomSubdomainResponse, void>({
      query: () => ({
        url: DOMAIN_SETTINGS.GENERATE_RANDOM_SUBDOMAIN,
        method: 'GET'
      }),
      transformResponse: (response: { data: RandomSubdomainResponse }) => {
        return response.data;
      }
    })
  })
});

export const {
  useGetAllDomainsQuery,
  useCreateDomainMutation,
  useUpdateDomainMutation,
  useDeleteDomainMutation,
  useGenerateRandomSubdomainQuery
} = domainsApi;
