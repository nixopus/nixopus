import { USERURLS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import {
  AddUserToOrganizationRequest,
  CreateOrganizationRequest,
  Organization,
  OrganizationUsers,
  UpdateOrganizationDetailsRequest,
  UserOrganization
} from '@/redux/types/orgs';
import { baseQueryWithReauth } from '@/redux/base-query';

export const userApi = createApi({
  reducerPath: 'userApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['User'],
  endpoints: (builder) => ({
    getUserOrganizations: builder.query<UserOrganization[], void>({
      query: () => ({
        url: USERURLS.USER_ORGANIZATIONS,
        method: 'GET'
      }),
      providesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: { data: UserOrganization[] }) => {
        return response.data;
      }
    }),
    createOrganization: builder.mutation<Organization, CreateOrganizationRequest>({
      query(organization) {
        return {
          url: USERURLS.CREATE_ORGANIZATION,
          method: 'POST',
          body: organization
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: { data: Organization }) => {
        return response.data;
      }
    }),
    addUserToOrganization: builder.mutation<void, AddUserToOrganizationRequest>({
      query(payload) {
        return {
          url: USERURLS.ADD_USER_TO_ORGANIZATION,
          method: 'POST',
          body: payload
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: { data: void }) => {
        return response.data;
      }
    }),
    updateUserName: builder.mutation<string, string>({
      query(name) {
        return {
          url: USERURLS.UPDATE_USER_NAME,
          method: 'PATCH',
          body: { name }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: { data: string }) => {
        return response.data;
      }
    }),
    requestPasswordResetLink: builder.mutation<void, void>({
      query() {
        return {
          url: USERURLS.REQUEST_PASSWORD_RESET_LINK,
          method: 'POST'
        };
      }
    }),
    getOrganizationUsers: builder.query<OrganizationUsers[], string>({
      query(organizationId) {
        return {
          url: `${USERURLS.ORGANIZATION_USERS}?id=${organizationId}`,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: { data: OrganizationUsers[] }) => {
        return response.data;
      }
    }),
    updateOrganizationDetails: builder.mutation<Organization, UpdateOrganizationDetailsRequest>({
      query(payload) {
        return {
          url: `${USERURLS.CREATE_ORGANIZATION}?id=${payload.id}`,
          method: 'PUT',
          body: payload
        };
      }
    })
  })
});

export const {
  useGetUserOrganizationsQuery,
  useCreateOrganizationMutation,
  useAddUserToOrganizationMutation,
  useUpdateUserNameMutation,
  useRequestPasswordResetLinkMutation,
  useGetOrganizationUsersQuery,
  useUpdateOrganizationDetailsMutation
} = userApi;
