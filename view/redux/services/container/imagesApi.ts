import { IMAGEURLS } from '@/redux/api-conf';
import { baseQueryWithReauth } from '@/redux/base-query';
import { createApi } from '@reduxjs/toolkit/query/react';
import { toast } from 'sonner';

export const imagesApi = createApi({
    reducerPath: 'imagesApi',
    baseQuery: baseQueryWithReauth,
    tagTypes: ['Images'],
    endpoints: (builder) => ({
        getImages: builder.query<any[], { containerId: string, imagePrefix: string }>({
            query: ({ containerId, imagePrefix }) => ({
                url: IMAGEURLS.GET_IMAGES,
                method: 'POST',
                body: {
                    all: true,
                    container_id: containerId,
                    image_prefix: imagePrefix,
                },
            }),
            providesTags: ['Images'],
            transformResponse: (response: any) => response.data,
        }),
        pruneImages: builder.mutation<any, { dangling?: boolean }>({
            query: (body) => ({
                url: IMAGEURLS.PRUNE_IMAGES,
                method: 'POST',
                body,
            }),
            invalidatesTags: ['Images'],
            onQueryStarted: async (_, { queryFulfilled }) => {
                try {
                    await queryFulfilled;
                    toast.success('Images pruned successfully');
                } catch (error) {
                    toast.error('Failed to prune images');
                }
            },
        }),
        pruneBuildCache: builder.mutation<any, { all?: boolean }>({
            query: (body) => ({
                url: IMAGEURLS.PRUNE_BUILD_CACHE,
                method: 'POST',
                body,
            }),
            onQueryStarted: async (_, { queryFulfilled }) => {
                try {
                    await queryFulfilled;
                    toast.success('Build cache pruned successfully');
                } catch (error) {
                    toast.error('Failed to prune build cache');
                }
            },
        }),
    }),
});

export const {
    useGetImagesQuery,
    usePruneImagesMutation,
    usePruneBuildCacheMutation,
} = imagesApi; 