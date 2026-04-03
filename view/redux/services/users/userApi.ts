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

export interface PendingInviteResponse {
  id: string;
  email: string;
  role: string;
  status: string;
  invitedBy?: { name?: string; email?: string };
  invitedAt: string;
  expiresAt?: string;
  organizationId: string;
}

export interface CancelInviteRequest {
  invitationId: string;
  organization_id: string;
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
    // Note: getUserOrganizations removed - now using Better Auth service layer via useUserOrganizations hook
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
          url: '/api/auth/organization/remove-member',
          method: 'POST',
          body: {
            memberIdOrEmail: payload.member_id,
            organizationId: payload.organization_id
          }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }]
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
    getActiveMember: builder.query<
      { role: string | string[]; permissions?: string[] },
      { organizationId: string; userId: string }
    >({
      query: ({ organizationId }) => {
        // Use list-members endpoint with organizationId
        // We'll filter by userId in transformResponse
        const url =
          typeof window !== 'undefined'
            ? `${window.location.origin}/api/auth/organization/list-members?organizationId=${organizationId}`
            : `auth/organization/list-members?organizationId=${organizationId}`;
        return {
          url,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'User', id: 'ACTIVE_MEMBER' }],
      keepUnusedDataFor: 120,
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
    // Note: getOrganizationUsers removed - now using Better Auth service layer via useOrganizationMembers hook
    getOrganizationUsers: builder.query<OrganizationUsers[], string>({
      query(organizationId) {
        return {
          url: `/api/auth/organization/list-members?organizationId=${organizationId}`,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: unknown): OrganizationUsers[] => {
        let members: any[] = [];
        if (Array.isArray(response)) {
          members = response;
        } else if (response && typeof response === 'object') {
          const r = response as Record<string, unknown>;
          if (Array.isArray(r.members)) members = r.members;
          else if (Array.isArray(r.data)) members = r.data;
        }
        if (!members.length) return [];
        return members.map(
          (member: any): OrganizationUsers => ({
            id: member.id,
            user_id: member.userId,
            organization_id: member.organizationId,
            created_at: member.createdAt || new Date().toISOString(),
            updated_at: member.updatedAt || new Date().toISOString(),
            deleted_at: null,
            user: {
              id: member.user?.id || member.userId,
              username: member.user?.name || member.user?.email || '',
              email: member.user?.email || '',
              avatar: member.user?.image || undefined,
              type: Array.isArray(member.role)
                ? member.role[0] || 'member'
                : member.role || 'member',
              is_verified: member.user?.emailVerified || false,
              is_email_verified: member.user?.emailVerified || false,
              two_factor_enabled: false,
              two_factor_secret: '',
              created_at: member.user?.createdAt || new Date().toISOString(),
              updated_at: member.user?.updatedAt || new Date().toISOString(),
              organization_users: [],
              organizations: []
            },
            roles: Array.isArray(member.role) ? member.role : [member.role || 'member'],
            permissions: []
          })
        );
      }
    }),
    getPendingInvites: builder.query<PendingInviteResponse[], string>({
      query(organizationId) {
        return {
          url: `/api/auth/organization/list-invitations?organizationId=${organizationId}`,
          method: 'GET'
        };
      },
      providesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: unknown): PendingInviteResponse[] => {
        if (!response || !Array.isArray(response)) {
          return [];
        }
        return response.map(
          (invite: any): PendingInviteResponse => ({
            id: invite.id,
            email: invite.email,
            role: invite.role || 'member',
            status: invite.status || 'pending',
            organizationId: invite.organizationId,
            invitedBy: invite.invitedBy,
            invitedAt: invite.createdAt || invite.invitedAt || new Date().toISOString(),
            expiresAt: invite.expiresAt
          })
        );
      }
    }),
    cancelInvite: builder.mutation<{ message: string }, CancelInviteRequest>({
      query(payload) {
        return {
          url: '/api/auth/organization/cancel-invitation',
          method: 'POST',
          body: {
            invitationId: payload.invitationId
          }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: () => {
        return { message: 'Invitation canceled successfully' };
      }
    }),
    updateOrganizationDetails: builder.mutation<Organization, UpdateOrganizationDetailsRequest>({
      query(payload) {
        return {
          url: '/api/auth/organization/update',
          method: 'POST',
          body: {
            organizationId: payload.id,
            data: {
              name: payload.name,
              metadata: { description: payload.description }
            }
          }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: (response: any) => {
        return {
          id: response?.id || '',
          name: response?.name || '',
          description: response?.metadata?.description || '',
          created_at: response?.createdAt || new Date().toISOString(),
          updated_at: response?.updatedAt || new Date().toISOString(),
          deleted_at: null
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
          url: '/api/auth/organization/update-member-role',
          method: 'POST',
          body: {
            memberId: payload.member_id,
            role: payload.role,
            organizationId: payload.organization_id
          }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }]
    }),
    sendInvite: builder.mutation<{ message: string }, InviteSendRequest>({
      query(payload) {
        return {
          url: '/api/auth/organization/invite-member',
          method: 'POST',
          body: {
            email: payload.email,
            role: payload.role,
            organizationId: payload.organization_id,
            resend: false
          }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: () => {
        return { message: 'Invitation sent successfully' };
      }
    }),
    resendInvite: builder.mutation<{ message: string }, InviteResendRequest>({
      query(payload) {
        return {
          url: '/api/auth/organization/invite-member',
          method: 'POST',
          body: {
            email: payload.email,
            role: payload.role,
            organizationId: payload.organization_id,
            resend: true
          }
        };
      },
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
      transformResponse: () => {
        return { message: 'Invitation resent successfully' };
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
      keepUnusedDataFor: 300,
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
  // useGetUserOrganizationsQuery - Removed: Use useUserOrganizations from @/packages/hooks/auth/use-better-auth-orgs
  useCreateOrganizationMutation,
  useAddUserToOrganizationMutation,
  useRemoveUserFromOrganizationMutation,
  useUpdateUserNameMutation,
  useGetActiveMemberQuery,
  useGetOrganizationUsersQuery,
  useGetPendingInvitesQuery,
  useCancelInviteMutation,
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
