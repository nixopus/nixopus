import { USERURLS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from './authApi';
import { AddUserToOrganizationRequest, CreateOrganizationRequest, Organization, UserOrganization } from '@/redux/types/orgs';

export const userApi = createApi({
    reducerPath: 'userApi',
    baseQuery: baseQueryWithReauth,
    tagTypes: ['User'],
    endpoints: (builder) => ({
        getUserOrganizations: builder.query<UserOrganization[], void>({
            query: () => ({
                url: USERURLS.USER_ORGANIZATIONS,
                method: 'GET',
            }),
            providesTags: [{ type: 'User', id: 'LIST' }],
            transformResponse: (response: { data: UserOrganization[] }) => {
                return response.data;
            },
        }),
        createOrganization: builder.mutation<Organization, CreateOrganizationRequest>({
            query(organization) {
                return {
                    url: USERURLS.CREATE_ORGANIZATION,
                    method: 'POST',
                    body: organization,
                };
            },
            invalidatesTags: [{ type: 'User', id: 'LIST' }],
            transformResponse: (response: { data: Organization }) => {
                return response.data;
            },
        }),
        addUserToOrganization: builder.mutation<void, AddUserToOrganizationRequest>({
            query(payload) {
                return {
                    url: USERURLS.ADD_USER_TO_ORGANIZATION,
                    method: 'POST',
                    body: payload,
                };
            },
            invalidatesTags: [{ type: 'User', id: 'LIST' }],
            transformResponse: (response: { data: void }) => {
                return response.data;
            },
        }),
    }),
});

export const {
    useGetUserOrganizationsQuery,
    useCreateOrganizationMutation,
    useAddUserToOrganizationMutation
} = userApi;