import { IMAGEURLS } from '@/redux/api-conf';
import { baseQueryWithReauth } from '@/redux/base-query';
import { createApi } from '@reduxjs/toolkit/query/react';

export const imagesApi = createApi({
  reducerPath: 'imagesApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Images'],
  endpoints: (builder) => ({
    getImages: builder.query<
      any[],
      { containerId: string; imagePrefix: string; server_id?: string }
    >({
      query: ({ containerId, imagePrefix, server_id }) => ({
        url: `${IMAGEURLS.GET_IMAGES}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
        method: 'POST',
        body: {
          all: true,
          container_id: containerId,
          image_prefix: imagePrefix
        }
      }),
      providesTags: ['Images'],
      transformResponse: (response: any) => response.data
    }),
    pruneImages: builder.mutation<
      any,
      { dangling?: boolean; until?: string; label?: string; server_id?: string }
    >({
      query: ({ server_id, ...body }) => ({
        url: `${IMAGEURLS.PRUNE_IMAGES}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
        method: 'POST',
        body
      }),
      invalidatesTags: ['Images']
    }),
    pruneBuildCache: builder.mutation<any, { all?: boolean; server_id?: string }>({
      query: ({ server_id, ...body }) => ({
        url: `${IMAGEURLS.PRUNE_BUILD_CACHE}${server_id ? `?server_id=${encodeURIComponent(server_id)}` : ''}`,
        method: 'POST',
        body
      }),
      invalidatesTags: ['Images']
    })
  })
});

export const { useGetImagesQuery, usePruneImagesMutation, usePruneBuildCacheMutation } = imagesApi;
