import { UserOrganization } from '@/redux/types/orgs';
import { createSlice } from '@reduxjs/toolkit';
import { fetchUserOrganizations } from './orgSlice';

interface UserState {
  organizations: UserOrganization[];
  activeOrganization: UserOrganization | null;
}

const initialState: UserState = {
  organizations: [],
  activeOrganization: null
};

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    setActiveOrganization: (state, action) => {
      state.activeOrganization = action.payload;
    }
  },
  extraReducers: (builder) => {
    // Sync organizations from orgSlice when fetched
    builder.addCase(fetchUserOrganizations.fulfilled, (state, { payload }) => {
      if (payload.length > 0) {
        state.organizations = payload;
      }
    });
  }
});

export const { setActiveOrganization } = userSlice.actions;

export default userSlice.reducer;
