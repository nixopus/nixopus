import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { Extension, ExtensionListParams, ExtensionListResponse } from '@/redux/types/extension';
import { EXTENSIONURLS } from '@/redux/api-conf';

export const extensionsApi = createApi({
  reducerPath: 'extensionsApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Extensions', 'Extension'],
  endpoints: (builder) => ({
    getExtensions: builder.query<ExtensionListResponse, ExtensionListParams>({
      query: (params) => {
        const searchParams = new URLSearchParams();
        
        if (params.category) {
          searchParams.append('category', params.category);
        }
        if (params.search) {
          searchParams.append('search', params.search);
        }
        if (params.sort_by) {
          searchParams.append('sort_by', params.sort_by);
        }
        if (params.sort_dir) {
          searchParams.append('sort_dir', params.sort_dir);
        }
        if (params.page) {
          searchParams.append('page', params.page.toString());
        }
        if (params.page_size) {
          searchParams.append('page_size', params.page_size.toString());
        }
        
        const queryString = searchParams.toString();
        return {
          url: `${EXTENSIONURLS.GET_EXTENSIONS}${queryString ? `?${queryString}` : ''}`,
          method: 'GET'
        };
      },
      providesTags: ['Extensions'],
      transformResponse: (response: ExtensionListResponse) => response
    }),
    getExtension: builder.query<Extension, { id: string }>({
      query: ({ id }) => ({
        url: EXTENSIONURLS.GET_EXTENSION.replace('{id}', id),
        method: 'GET'
      }),
      providesTags: (result, error, { id }) => [{ type: 'Extension', id }],
      transformResponse: (response: Extension) => response
    }),
    getExtensionByExtensionId: builder.query<Extension, { extensionId: string }>({
      query: ({ extensionId }) => ({
        url: EXTENSIONURLS.GET_EXTENSION_BY_ID.replace('{extension_id}', extensionId),
        method: 'GET'
      }),
      providesTags: (result, error, { extensionId }) => [{ type: 'Extension', id: extensionId }],
      transformResponse: (response: Extension) => response
    })
  })
});

export const {
  useGetExtensionsQuery,
  useGetExtensionQuery,
  useGetExtensionByExtensionIdQuery
} = extensionsApi;
