import { GITHUB_CONNECTOR } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from '@/redux/base-query';
import {
  CreateGithubConnectorRequest,
  GitHubAppCredentials,
  GithubConnector,
  GithubRepository,
  GithubRepositoryBranch,
  UpdateGithubConnectorRequest
} from '@/redux/types/github';

export const GithubConnectorApi = createApi({
  reducerPath: 'GithubConnectorApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['GithubConnector'],
  endpoints: (builder) => ({
    getAllGithubConnector: builder.query<GithubConnector[], void>({
      query: () => ({
        url: GITHUB_CONNECTOR.GET_GITHUB_CONNECTORS,
        method: 'GET'
      }),
      providesTags: [{ type: 'GithubConnector', id: 'LIST' }],
      transformResponse: (response: { data: GithubConnector[] }) => {
        return response.data;
      }
    }),
    createGithubConnector: builder.mutation<null, CreateGithubConnectorRequest>({
      query: (data) => ({
        url: GITHUB_CONNECTOR.ADD_GITHUB_CONNECTOR,
        method: 'POST',
        body: data
      }),
      invalidatesTags: [{ type: 'GithubConnector', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    updateGithubConnector: builder.mutation<null, UpdateGithubConnectorRequest>({
      query: (data) => ({
        url: GITHUB_CONNECTOR.UPDATE_GITHUB_CONNECTOR,
        method: 'PUT',
        body: data
      }),
      invalidatesTags: [{ type: 'GithubConnector', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    deleteGithubConnector: builder.mutation<null, string>({
      query: (id) => ({
        url: GITHUB_CONNECTOR.DELETE_GITHUB_CONNECTOR,
        method: 'DELETE',
        body: { id }
      }),
      invalidatesTags: [{ type: 'GithubConnector', id: 'LIST' }],
      transformResponse: (response: { data: null }) => {
        return response.data;
      }
    }),
    getAllGithubRepositories: builder.query<GithubRepository[], void>({
      query: () => ({
        url: GITHUB_CONNECTOR.ALL_REPOSITORIES,
        method: 'GET'
      }),
      providesTags: [{ type: 'GithubConnector', id: 'LIST' }],
      transformResponse: (response: { data: any[] }) => {
        return response.data;
      }
    }),
    getGithubRepositoryBranches: builder.mutation<GithubRepositoryBranch[], string>({
      query: (repository_name) => ({
        url: GITHUB_CONNECTOR.GET_REPOSITORY_BRANCHES,
        method: 'POST',
        body: { repository_name }
      }),
      invalidatesTags: [{ type: 'GithubConnector', id: 'LIST' }],
      transformResponse: (response: { data: any[] }) => {
        return response.data;
      }
    })
  })
});

export const {
  useCreateGithubConnectorMutation,
  useUpdateGithubConnectorMutation,
  useDeleteGithubConnectorMutation,
  useGetAllGithubConnectorQuery,
  useGetAllGithubRepositoriesQuery,
  useGetGithubRepositoryBranchesMutation
} = GithubConnectorApi;
