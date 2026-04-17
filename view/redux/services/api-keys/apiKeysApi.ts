import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';

export interface ApiKey {
  id: string;
  name: string | null;
  start: string | null;
  prefix: string | null;
  userId?: string;
  referenceId?: string;
  configId?: string;
  enabled: boolean;
  expiresAt: string | null;
  createdAt: string;
  updatedAt: string;
  remaining: number | null;
  rateLimitEnabled: boolean;
  rateLimitMax: number | null;
  rateLimitTimeWindow: number | null;
  requestCount: number;
  lastRequest: string | null;
  metadata: Record<string, unknown> | null;
  permissions: string | null;
}

export interface CreateApiKeyRequest {
  name?: string;
  expiresIn?: number;
  prefix?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateApiKeyResponse extends ApiKey {
  key: string;
}

export interface UpdateApiKeyRequest {
  keyId: string;
  name?: string;
}

export interface DeleteApiKeyRequest {
  keyId: string;
}

export const apiKeysApi = createApi({
  reducerPath: 'apiKeysApi',
  baseQuery: baseQueryWithReauth,
  keepUnusedDataFor: 300,
  tagTypes: ['ApiKey'],
  endpoints: (builder) => ({
    listApiKeys: builder.query<ApiKey[], void>({
      query: () => ({ url: '/api/auth/api-key/list', method: 'GET' }),
      providesTags: [{ type: 'ApiKey', id: 'LIST' }],
      transformResponse: (response: unknown): ApiKey[] => {
        if (Array.isArray(response)) return response;
        if (response && typeof response === 'object') {
          const r = response as Record<string, unknown>;
          if (Array.isArray(r.apiKeys)) return r.apiKeys as ApiKey[];
          if (Array.isArray(r.data)) return r.data as ApiKey[];
        }
        return [];
      },
    }),
    createApiKey: builder.mutation<CreateApiKeyResponse, CreateApiKeyRequest>({
      query: (data) => ({ url: '/api/auth/api-key/create', method: 'POST', body: data }),
      invalidatesTags: [{ type: 'ApiKey', id: 'LIST' }],
    }),
    updateApiKey: builder.mutation<ApiKey, UpdateApiKeyRequest>({
      query: (data) => ({ url: '/api/auth/api-key/update', method: 'POST', body: data }),
      invalidatesTags: [{ type: 'ApiKey', id: 'LIST' }],
      async onQueryStarted({ keyId, name }, { dispatch, queryFulfilled }) {
        const patch = dispatch(
          apiKeysApi.util.updateQueryData('listApiKeys', undefined, (draft) => {
            const key = draft.find((k) => k.id === keyId);
            if (key && name !== undefined) key.name = name;
          }),
        );
        try {
          await queryFulfilled;
        } catch {
          patch.undo();
        }
      },
    }),
    deleteApiKey: builder.mutation<{ success: boolean }, DeleteApiKeyRequest>({
      query: (data) => ({ url: '/api/auth/api-key/delete', method: 'POST', body: data }),
      invalidatesTags: [{ type: 'ApiKey', id: 'LIST' }],
      async onQueryStarted({ keyId }, { dispatch, queryFulfilled }) {
        const patch = dispatch(
          apiKeysApi.util.updateQueryData('listApiKeys', undefined, (draft) => {
            const idx = draft.findIndex((k) => k.id === keyId);
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
  }),
});

export const {
  useListApiKeysQuery,
  useCreateApiKeyMutation,
  useUpdateApiKeyMutation,
  useDeleteApiKeyMutation,
} = apiKeysApi;
