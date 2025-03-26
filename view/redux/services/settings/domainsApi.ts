import { DOMAIN_SETTINGS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { Domain } from '@/redux/types/domain';

export const domainsApi = createApi({
  reducerPath: 'domainsApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Domains'],
  endpoints: (builder) => ({
    getAllDomains: builder.query<Domain[], { organizationId: string }>({
      query: ({ organizationId }) => ({
        url: DOMAIN_SETTINGS.GET_DOMAINS + `?id=${organizationId}`,
        method: 'GET'
      }),
      providesTags: [{ type: 'Domains', id: 'LIST' }],
      transformResponse: (response: { data: Domain[] }) => {
        return response.data;
      }
    }),
    createDomain: builder.mutation<null, { name: string; organization_id: string }>({
      query: (data) => ({
        url: DOMAIN_SETTINGS.ADD_DOMAIN,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'Domains', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    updateDomain: builder.mutation<null, { name: string; id: string }>({
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
    deleteDomain: builder.mutation<null, string>({
      query: (id) => ({
        url: DOMAIN_SETTINGS.DELETE_DOMAIN,
        method: 'DELETE',
        body: { id }
      }),
      invalidatesTags: [{ type: 'Domains', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    generateRandomSubdomain: builder.query<string, void>({
      query: (id) => ({
        url: DOMAIN_SETTINGS.GENERATE_RANDOM_SUBDOMAIN,
        method: 'GET'
      })
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
