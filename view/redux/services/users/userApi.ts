import { USERURLS } from '@/redux/api-conf';
import { createApi } from '@reduxjs/toolkit/query/react';
import { baseQueryWithReauth } from './authApi';
import { UserOrganization } from '@/redux/types/orgs';

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
    }),
});

export const {
    useGetUserOrganizationsQuery
} = userApi;