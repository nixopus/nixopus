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

export interface InviteAcceptRequest {
  token: string;
  organization_id?: string;
  role?: string;
  email?: string;
}

import { baseQueryWithReauth } from '@/redux/base-query';
import {
  User,
  UserSettings,
  UpdateFontRequest,
  UpdateThemeRequest,
  UpdateLanguageRequest,
  UpdateAutoUpdateRequest,
  UpdateAvatarRequest,
  UserPreferences,
  UserPreferencesData,
  OrganizationSettings,
  OrganizationSettingsData,
  UpdateCheckResponse
} from '@/redux/types/user';

export const userApi = createApi({
  reducerPath: 'userApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['User'],
  endpoints: (builder) => ({
    getUserOrganizations: builder.query<UserOrganization[], void>({
      query: () => {
        // Better Auth endpoints go through Next.js proxy, use full URL with frontend origin
        const url = typeof window !== 'undefined' 
          ? `${window.location.origin}/api/${USERURLS.USER_ORGANIZATIONS}`
          : USERURLS.USER_ORGANIZATIONS;
        return {
          url,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: any) => {
        // Better Auth returns organizations directly as array, transform to UserOrganization format
        if (!response || !Array.isArray(response)) {
          return [];
        }
        return response.map((org: any) => ({
          id: org.id,
          organization: {
            id: org.id,
            name: org.name,
            description: org.metadata?.description || '',
            created_at: org.createdAt || new Date().toISOString(),
            updated_at: org.updatedAt || new Date().toISOString(),
            deleted_at: null,
          },
          role: {
            id: org.role || 'member',
            name: org.role || 'member',
            description: '',
          },
          created_at: org.createdAt || new Date().toISOString(),
          updated_at: org.updatedAt || new Date().toISOString(),
          deleted_at: null,
        }));
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
    getActiveMember: builder.query<{ role: string | string[]; permissions?: string[] }, { organizationId: string; userId: string }>({
      query: ({ organizationId }) => {
        // Use list-members endpoint with organizationId
        // We'll filter by userId in transformResponse
        const url = typeof window !== 'undefined' 
          ? `${window.location.origin}/api/auth/organization/list-members?organizationId=${organizationId}`
          : `auth/organization/list-members?organizationId=${organizationId}`;
        return {
          url,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'User', id: 'ACTIVE_MEMBER' }],
      transformResponse: (response: any, meta, { userId }) => {
        // Better Auth returns array of members, find the current user's member record
        let members = response;
        
        // If wrapped in data property
        if (response?.data && Array.isArray(response.data)) {
          members = response.data;
        } else if (!Array.isArray(response)) {
          // If not an array, try to extract members
          members = [];
        }
        
        if (!members || members.length === 0) {
          console.warn('getActiveMember: No members found');
          return { role: 'member', permissions: [] };
        }
        
        // Find the member record for the current user
        const member = members.find((m: any) => {
          const memberUserId = m.userId || m.user?.id;
          return memberUserId === userId;
        });
        
        if (!member) {
          console.warn('getActiveMember: Current user not found in members list');
          return { role: 'member', permissions: [] };
        }
        
        const role = member.role || 'member';
        const permissions = member.permissions || [];
        
        return {
          role,
          permissions: Array.isArray(permissions) ? permissions : []
        };
      }
    }),
    getOrganizationUsers: builder.query<OrganizationUsers[], string>({
      query(organizationId) {
        // Better Auth endpoints go through Next.js proxy, use full URL with frontend origin
        const url = typeof window !== 'undefined' 
          ? `${window.location.origin}/api/auth/organization/list-members?organizationId=${organizationId}`
          : `auth/organization/list-members?organizationId=${organizationId}`;
        return {
          url,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: any) => {
        // Better Auth returns members directly as array
        if (!response || !Array.isArray(response)) {
          return [];
        }
        // Transform Better Auth member format to OrganizationUsers format
        return response.map((member: any) => {
          const user = member.user || {};
          const role = member.role || 'member';
          
          return {
            id: member.id,
            user_id: member.userId || user.id,
            organization_id: member.organizationId,
            created_at: member.createdAt || new Date().toISOString(),
            updated_at: member.updatedAt || new Date().toISOString(),
            deleted_at: null,
            user: {
              id: user.id || member.userId,
              email: user.email || '',
              username: user.name || user.username || '',
              avatar: user.image || user.avatar || '',
              type: Array.isArray(role) ? role[0] : role,
              organization_users: [],
              is_verified: user.emailVerified || false,
              is_email_verified: user.emailVerified || false,
              two_factor_enabled: user.twoFactorEnabled || false,
              two_factor_secret: '',
              created_at: user.createdAt || new Date().toISOString(),
              updated_at: user.updatedAt || new Date().toISOString(),
              organizations: []
            } as User,
            roles: Array.isArray(role) ? role : [role],
            permissions: member.permissions || []
          };
        });
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
    acceptInvite: builder.mutation<{ message: string }, InviteAcceptRequest>({
      query(payload) {
        return {
          url: USERURLS.ACCEPT_INVITE,
          method: 'POST',
          body: payload
        };
      },
      transformResponse: (response: { data: { message: string } }) => {
        return response.data;
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }]
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
    checkForUpdates: builder.query<UpdateCheckResponse, void>({
      query: () => ({
        url: USERURLS.CHECK_FOR_UPDATES,
        method: 'GET'
      }),
      transformResponse: (response: { data: UpdateCheckResponse }) => {
        return response.data;
      }
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
    }),

    getUserPreferences: builder.query<UserPreferences, void>({
      query: () => ({
        url: USERURLS.GET_PREFERENCES,
        method: 'GET'
      }),
      transformResponse: (response: { data: UserPreferences }) => {
        return response.data;
      }
    }),
    updateUserPreferences: builder.mutation<UserPreferences, UserPreferencesData>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_PREFERENCES,
          method: 'PUT',
          body: payload
        };
      },
      transformResponse: (response: { data: UserPreferences }) => {
        return response.data;
      }
    }),
    getOrganizationSettings: builder.query<OrganizationSettings, void>({
      query: () => ({
        url: USERURLS.GET_ORGANIZATION_SETTINGS,
        method: 'GET'
      }),
      transformResponse: (response: { data: OrganizationSettings }) => {
        return response.data;
      }
    }),
    updateOrganizationSettings: builder.mutation<OrganizationSettings, OrganizationSettingsData>({
      query(payload) {
        return {
          url: USERURLS.UPDATE_ORGANIZATION_SETTINGS,
          method: 'PUT',
          body: payload
        };
      },
      transformResponse: (response: { data: OrganizationSettings }) => {
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
  useGetActiveMemberQuery,
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
  useResendInviteMutation,
  useAcceptInviteMutation,
  useGetUserPreferencesQuery,
  useUpdateUserPreferencesMutation,
  useGetOrganizationSettingsQuery,
  useUpdateOrganizationSettingsMutation
} = userApi;
