import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import { FileData } from '@/redux/types/files';
import { FILEMANAGERURLS } from '@/redux/api-conf';

export const fileManagersApi = createApi({
    reducerPath: 'fileManagersApi',
    baseQuery: baseQueryWithReauth,
    tagTypes: ['fileManager'],
    endpoints: (builder) => ({
        getFilesInPath: builder.query<FileData[], { path: string }>({
            query: ({ path }) => ({
                url: `${FILEMANAGERURLS.LIST_FILES_AT_PATH}?path=${encodeURIComponent(path)}`,
                method: 'GET',
            }),
            transformResponse: (response: { data: FileData[] }) => response.data,
        }),
        createDirectory: builder.mutation<CreateDirectoryResponse, { path: string; name: string }>({
            query: ({ path, name }) => ({
                url: FILEMANAGERURLS.CREATE_DIRECTORY,
                method: 'POST',
                body: { path, name },
            }),
            transformResponse: (response: CreateDirectoryResponse) => response,
        }),
        deleteDirectory: builder.mutation<CreateDirectoryResponse, { path: string }>({
            query: ({ path }) => ({
                url: FILEMANAGERURLS.DELETE_DIRECTORY,
                method: 'DELETE',
                body: { path },
            }),
            transformResponse: (response: CreateDirectoryResponse) => response,
        }),
        deleteFile: builder.mutation<CreateDirectoryResponse, { path: string }>({
            query: ({ path }) => ({
                url: FILEMANAGERURLS.DELETE_FILE,
                method: 'DELETE',
                body: { path },
            }),
            transformResponse: (response: CreateDirectoryResponse) => response,
        }),
        moveOrRenameDirectory: builder.mutation<
            CreateDirectoryResponse,
            { from_path: string; to_path: string }
        >({
            query: ({ from_path, to_path }) => ({
                url: FILEMANAGERURLS.MOVE_FOLDER_FILES_RECURSIVELY_OR_RENAME,
                method: 'POST',
                body: { from_path, to_path },
            }),
            transformResponse: (response: CreateDirectoryResponse) => response,
        }),
        copyFileOrDirectory: builder.mutation<
            CreateDirectoryResponse,
            { from_path: string; to_path: string }
        >({
            query: ({ from_path, to_path }) => ({
                url: FILEMANAGERURLS.COPY_FOLDER_FILES_RECURSIVELY,
                method: 'POST',
                body: { from_path, to_path },
            }),
            transformResponse: (response: CreateDirectoryResponse) => response,
        }),
        createFile: builder.mutation<CreateDirectoryResponse, { path: string; name: string }>({
            query: ({ path, name }) => ({
                url: FILEMANAGERURLS.CREATE_FILE,
                method: 'POST',
                body: { path, name },
            }),
            transformResponse: (response: CreateDirectoryResponse) => response,
        }),
        calculateDirectorySize: builder.mutation<FileSizeResponse['data'], { path: string }>({
            query: ({ path }) => ({
                url: FILEMANAGERURLS.CALCULATE_DIRECTORY_SIZE,
                method: 'POST',
                body: { path },
            }),
            transformResponse: (response: FileSizeResponse) => response.data,
        }),
        getDiskUsage: builder.query<DiskUsageData, void>({
            query: () => ({
                url: FILEMANAGERURLS.GET_DISK_USAGE,
                method: 'GET',
            }),
            transformResponse: (response: { data: DiskUsageData }) => response.data,
        }),
        getMemoryUsage: builder.query<MemoryUsageData, void>({
            query: () => ({
                url: FILEMANAGERURLS.GET_MEMORY_USAGE,
                method: 'GET',
            }),
            transformResponse: (response: { data: MemoryUsageData }) => response.data,
        }),
    })
});

export const {
    useGetFilesInPathQuery,
    useCreateDirectoryMutation,
    useDeleteDirectoryMutation,
    useDeleteFileMutation,
    useMoveOrRenameDirectoryMutation,
    useCopyFileOrDirectoryMutation,
    useCreateFileMutation,
    useCalculateDirectorySizeMutation,
    useGetDiskUsageQuery,
    useGetMemoryUsageQuery,
} = fileManagersApi;
