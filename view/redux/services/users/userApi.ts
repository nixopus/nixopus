import { USERURLS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import {
  AddUserToOrganizationRequest,
  CreateOrganizationRequest,
  CreateUserRequest,
  Organization,
  OrganizationUsers,
  RemoveUserFromOrganizationRequest,
  UpdateOrganizationDetailsRequest,
  UpdateUserRoleRequest,
  UserOrganization
} from '@/redux/types/orgs';

export interface InviteSendRequest {
  email: string;
  organization_id: string;
  role: string;
}

export interface InviteResendRequest {
  email: string;
  organization_id: string;
  role: string;
}

import { baseQueryWithReauth } from '@/redux/base-query';
import {
  UserSettings,
  UpdateFontRequest,
  UpdateThemeRequest,
  UpdateLanguageRequest,
  UpdateAutoUpdateRequest,
  UpdateAvatarRequest
} from '@/redux/types/user';

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
    removeUserFromOrganization: builder.mutation<void, RemoveUserFromOrganizationRequest>({
      query(payload) {
        return {
          url: USERURLS.REMOVE_USER_FROM_ORGANIZATION,
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
    }),
    createUser: builder.mutation<void, CreateUserRequest>({
      query(body) {
        return {
          url: USERURLS.CREATE_USER,
          method: 'POST',
          body
        };
      }
    }),
    updateUserRole: builder.mutation<void, UpdateUserRoleRequest>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_USER_ROLE,
          method: 'POST',
          body: payload
        };
      }
    }),
    sendInvite: builder.mutation<{ message: string }, InviteSendRequest>({
      query(payload) {
        return {
          url: USERURLS.SEND_INVITE,
          method: 'POST',
          body: payload
        };
      },
      transformResponse: (response: { data: { message: string } }) => {
        return response.data;
      }
    }),
    resendInvite: builder.mutation<{ message: string }, InviteResendRequest>({
      query(payload) {
        return {
          url: USERURLS.RESEND_INVITE,
          method: 'POST',
          body: payload
        };
      },
      transformResponse: (response: { data: { message: string } }) => {
        return response.data;
      }
    }),
    getResources: builder.query<string[], void>({
      query: () => ({
        url: USERURLS.GET_RESOURCES,
        method: 'GET'
      }),
      providesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: { data: string[] }) => {
        return response.data;
      }
    }),
    deleteOrganization: builder.mutation<void, string>({
      query(organizationId) {
        return {
          url: `${USERURLS.CREATE_ORGANIZATION}`,
          method: 'DELETE',
          body: { id: organizationId }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }]
    }),
    getUserSettings: builder.query<UserSettings, void>({
      query: () => ({
        url: USERURLS.GET_SETTINGS,
        method: 'GET'
      }),
      transformResponse: (response: { data: UserSettings }) => {
        return response.data;
      }
    }),
    updateFont: builder.mutation<UserSettings, UpdateFontRequest>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_FONT,
          method: 'PATCH',
          body: payload
        };
      }
    }),
    updateTheme: builder.mutation<UserSettings, UpdateThemeRequest>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_THEME,
          method: 'PATCH',
          body: payload
        };
      }
    }),
    updateLanguage: builder.mutation<UserSettings, UpdateLanguageRequest>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_LANGUAGE,
          method: 'PATCH',
          body: payload
        };
      }
    }),
    updateAutoUpdate: builder.mutation<UserSettings, UpdateAutoUpdateRequest>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_AUTO_UPDATE,
          method: 'PATCH',
          body: payload
        };
      }
    }),
    checkForUpdates: builder.query<void, void>({
      query: () => ({
        url: USERURLS.CHECK_FOR_UPDATES,
        method: 'GET'
      })
    }),
    performUpdate: builder.mutation<void, void>({
      query: () => ({
        url: USERURLS.PERFORM_UPDATE,
        method: 'POST'
      })
    }),
    updateAvatar: builder.mutation<string, UpdateAvatarRequest>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_AVATAR,
          method: 'PATCH',
          body: payload
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: { data: string }) => {
        return response.data;
      }
    })
  })
});

export const {
  useGetUserOrganizationsQuery,
  useCreateOrganizationMutation,
  useAddUserToOrganizationMutation,
  useRemoveUserFromOrganizationMutation,
  useUpdateUserNameMutation,
  useGetOrganizationUsersQuery,
  useUpdateOrganizationDetailsMutation,
  useCreateUserMutation,
  useUpdateUserRoleMutation,
  useGetResourcesQuery,
  useDeleteOrganizationMutation,
  useGetUserSettingsQuery,
  useUpdateFontMutation,
  useUpdateThemeMutation,
  useUpdateLanguageMutation,
  useUpdateAutoUpdateMutation,
  useCheckForUpdatesQuery,
  usePerformUpdateMutation,
  useUpdateAvatarMutation,
  useSendInviteMutation,
  useResendInviteMutation
} = userApi;
