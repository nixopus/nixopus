import { AUTHURLS } from '@/redux/api-conf';
import { BASE_URL } from '@/redux/conf';
import { RootState } from '@/redux/store';
import { AuthResponse, LoginPayload, RefreshTokenPayload, User } from '@/redux/types/user';
import { createApi, fetchBaseQuery, BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query/react';
import { setCredentials, logout } from '../../features/users/authSlice';

export const baseQueryWithReauth: BaseQueryFn<
    string | FetchArgs,
    unknown,
    FetchBaseQueryError
> = async (args, api, extraOptions) => {
    const baseQuery = fetchBaseQuery({
        baseUrl: BASE_URL,
        prepareHeaders: (headers, { getState }) => {
            const token = (getState() as RootState).auth.token;
            if (token) {
                headers.set('authorization', `Bearer ${token}`);
            }
            return headers;
        },
    });

    let result = await baseQuery(args, api, extraOptions);

    if (result.error && result.error.status === 401) {
        console.log('Token expired, attempting refresh');

        const refreshToken = (api.getState() as RootState).auth.refreshToken;

        if (!refreshToken) {
            api.dispatch(logout());
            return result;
        }

        const refreshResult = await baseQuery({
            url: AUTHURLS.REFRESH_TOKEN,
            method: 'POST',
            body: { refresh_token: refreshToken },
        }, api, extraOptions);

        if (refreshResult.data) {
            const refreshData = refreshResult.data as AuthResponse;

            api.dispatch(setCredentials({
                user: null,
                token: refreshData.access_token,
                refreshToken: refreshData.refresh_token,
                expiresIn: refreshData.expires_in
            }));

            result = await baseQuery(args, api, extraOptions);
        } else {
            api.dispatch(logout());
        }
    }

    return result;
};

export const authApi = createApi({
    reducerPath: 'authApi',
    baseQuery: baseQueryWithReauth,
    tagTypes: ['Authentication'],
    endpoints: (builder) => ({
        loginUser: builder.mutation<AuthResponse, LoginPayload>({
            query(credentials) {
                return {
                    url: AUTHURLS.USER_LOGIN,
                    method: 'POST',
                    body: credentials,
                };
            },
            transformResponse: (response: { data: AuthResponse }) => {
                return { ...response.data };
            },
            invalidatesTags: [{ type: 'Authentication', id: 'LIST' }],
        }),
        getUserDetails: builder.query<User, void>({
            query: () => ({
                url: AUTHURLS.USER_DETAILS,
                method: 'GET',
            }),
            providesTags: [{ type: 'Authentication', id: 'LIST' }],
            transformResponse: (response: { data: User }) => {
                return { ...response.data };
            },
        }),
        refreshToken: builder.mutation<AuthResponse, RefreshTokenPayload>({
            query: (payload) => ({
                url: AUTHURLS.REFRESH_TOKEN,
                method: 'POST',
                body: payload
            }),
            transformResponse: (response: { data: AuthResponse }) => {
                return { ...response.data };
            },
        })
    }),
});

export const {
    useLoginUserMutation,
    useGetUserDetailsQuery,
    useRefreshTokenMutation
} = authApi;