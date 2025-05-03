import { IMAGEURLS } from '@/redux/api-conf';
import { baseQueryWithReauth } from '@/redux/base-query';
import { createApi } from '@reduxjs/toolkit/query/react';

export const imagesApi = createApi({
  reducerPath: 'imagesApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Images'],
  endpoints: (builder) => ({
    getImages: builder.query<any[], { containerId: string; imagePrefix: string }>({
      query: ({ containerId, imagePrefix }) => ({
        url: IMAGEURLS.GET_IMAGES,
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
    pruneImages: builder.mutation<any, { dangling?: boolean; until?: string; label?: string }>({
      query: (body) => ({
        url: IMAGEURLS.PRUNE_IMAGES,
        method: 'POST',
        body
      }),
      invalidatesTags: ['Images']
    }),
    pruneBuildCache: builder.mutation<any, { all?: boolean }>({
      query: (body) => ({
        url: IMAGEURLS.PRUNE_BUILD_CACHE,
        method: 'POST',
        body
      }),
      invalidatesTags: ['Images']
    })
  })
});

export const { useGetImagesQuery, usePruneImagesMutation, usePruneBuildCacheMutation } = imagesApi;
