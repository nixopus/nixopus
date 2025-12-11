import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Server } from '@/redux/types/server';
import { serversApi } from '@/redux/services/settings/serversApi';

interface ServerState {
  activeServer: Server | null;
  activeServerId: string | null;
}

const initialState: ServerState = {
  activeServer: null,
  activeServerId: null
};

export const serverSlice = createSlice({
  name: 'server',
  initialState,
  reducers: {
    setActiveServer: (state, action: PayloadAction<Server | null>) => {
      state.activeServer = action.payload;
      state.activeServerId = action.payload?.id || null;
    },
    setActiveServerId: (state, action: PayloadAction<string | null>) => {
      state.activeServerId = action.payload;
      if (action.payload === null) {
        state.activeServer = null;
      }
    },
    clearActiveServer: (state) => {
      state.activeServer = null;
      state.activeServerId = null;
    }
  },
  extraReducers: (builder) => {
    builder
      .addMatcher(serversApi.endpoints.getActiveServer.matchFulfilled, (state, action) => {
        // Only update if we got a server from the API, or if we don't have a persisted server
        // This preserves persisted server state if API returns null but we have a persisted server
        if (action.payload) {
          state.activeServer = action.payload;
          state.activeServerId = action.payload.id;
        } else if (!state.activeServer) {
          // Only clear if we don't have a persisted server
          state.activeServer = null;
          state.activeServerId = null;
        }
        // If action.payload is null but state.activeServer exists, keep the persisted server
      })
      .addMatcher(serversApi.endpoints.updateServerStatus.matchFulfilled, (state, action) => {
        if (action.payload.status === 'active') {
          state.activeServer = action.payload;
          state.activeServerId = action.payload.id;
        } else if (state.activeServerId === action.payload.id) {
          state.activeServer = null;
          state.activeServerId = null;
        }
      });
  }
});

export const { setActiveServer, setActiveServerId, clearActiveServer } = serverSlice.actions;

export default serverSlice.reducer;
