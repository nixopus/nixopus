import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { CUSTOM_DOMAIN_URLS } from '@/redux/api-conf';
import type {
  CustomDomain,
  AddCustomDomainResponse,
  DNSCheckResponse,
} from '@/redux/types/custom-domains';

export const customDomainsApi = createApi({
  reducerPath: 'customDomainsApi',
  baseQuery: baseQueryWithReauth,
  keepUnusedDataFor: 300,
  tagTypes: ['CustomDomains'],
  endpoints: (builder) => ({
    listAllDomains: builder.query<CustomDomain[], void>({
      query: () => ({ url: CUSTOM_DOMAIN_URLS.LIST, method: 'GET' }),
      providesTags: [{ type: 'CustomDomains', id: 'LIST' }],
      transformResponse: (response: { data: CustomDomain[] }) => response.data,
    }),
    addCustomDomain: builder.mutation<AddCustomDomainResponse, { name: string }>({
      query: (data) => ({ url: CUSTOM_DOMAIN_URLS.ADD, method: 'POST', body: data }),
      invalidatesTags: [{ type: 'CustomDomains', id: 'LIST' }],
    }),
    verifyCustomDomain: builder.mutation<{ data: CustomDomain }, { id: string }>({
      query: (data) => ({ url: CUSTOM_DOMAIN_URLS.VERIFY, method: 'POST', body: data }),
      invalidatesTags: [{ type: 'CustomDomains', id: 'LIST' }],
      transformResponse: (response: { data: CustomDomain }) => ({ data: response.data }),
    }),
    removeCustomDomain: builder.mutation<void, { id: string }>({
      query: (data) => ({ url: CUSTOM_DOMAIN_URLS.REMOVE, method: 'DELETE', body: data }),
      invalidatesTags: [{ type: 'CustomDomains', id: 'LIST' }],
      async onQueryStarted({ id }, { dispatch, queryFulfilled }) {
        const patch = dispatch(
          customDomainsApi.util.updateQueryData('listAllDomains', undefined, (draft) => {
            const idx = draft.findIndex((d) => d.id === id);
            if (idx !== -1) draft.splice(idx, 1);
          }),
        );
        try {
          await queryFulfilled;
        } catch {
          patch.undo();
        }
      },
    }),
    checkDNSStatus: builder.query<DNSCheckResponse, { id: string }>({
      query: ({ id }) => ({ url: `${CUSTOM_DOMAIN_URLS.DNS_CHECK}?id=${id}`, method: 'GET' }),
    }),
  }),
});

export const {
  useListAllDomainsQuery,
  useAddCustomDomainMutation,
  useVerifyCustomDomainMutation,
  useRemoveCustomDomainMutation,
  useCheckDNSStatusQuery,
} = customDomainsApi;
