import { userApi } from '@/redux/services/users/userApi';
import { UserOrganization } from '@/redux/types/orgs';
import { createSlice } from '@reduxjs/toolkit';

interface UserState {
    organizations: UserOrganization[]
}

const initialState: UserState = {
    organizations: [],
};

export const userSlice = createSlice({
    name: 'user',
    initialState,
    reducers: {},
    extraReducers: (builder) => {
        builder
            .addMatcher(userApi.endpoints.getUserOrganizations.matchFulfilled, (state, { payload }) => {
                console.log("payload", payload)
                if (payload.length > 0) {
                    state.organizations = payload;
                }
            });
    },
});

export default userSlice.reducer;