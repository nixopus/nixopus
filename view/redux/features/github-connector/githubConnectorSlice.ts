import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { GithubRepository } from '@/redux/types/github';
import { createSlice, PayloadAction } from '@reduxjs/toolkit';

const ACTIVE_CONNECTOR_KEY = 'active_github_connector';

interface githubConnectorState {
  repositories: GithubRepository[];
  activeConnectorId: string | null;
}

const getInitialActiveConnectorId = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(ACTIVE_CONNECTOR_KEY);
};

const initialState: githubConnectorState = {
  repositories: [],
  activeConnectorId: getInitialActiveConnectorId()
};

export const githubConnectorSlice = createSlice({
  name: 'githubConnector',
  initialState,
  reducers: {
    setActiveConnectorId: (state, action: PayloadAction<string | null>) => {
      state.activeConnectorId = action.payload;
      if (typeof window !== 'undefined') {
        if (action.payload) {
          localStorage.setItem(ACTIVE_CONNECTOR_KEY, action.payload);
        } else {
          localStorage.removeItem(ACTIVE_CONNECTOR_KEY);
        }
      }
    }
  },
  extraReducers: (builder) => {
    builder.addMatcher(
      GithubConnectorApi.endpoints.getAllGithubRepositories.matchFulfilled,
      (state, { payload }) => {
        if (payload && Array.isArray(payload.repositories)) {
          state.repositories = payload.repositories;
        }
      }
    );
    // Initialize active connector from connectors list if not set
    builder.addMatcher(
      GithubConnectorApi.endpoints.getAllGithubConnector.matchFulfilled,
      (state, { payload }) => {
        if (!state.activeConnectorId && payload && payload.length > 0) {
          const storedId = getInitialActiveConnectorId();
          const connectorExists = payload.some((c) => c.id === storedId);
          if (connectorExists) {
            state.activeConnectorId = storedId;
          } else {
            state.activeConnectorId = payload[0].id;
            if (typeof window !== 'undefined') {
              localStorage.setItem(ACTIVE_CONNECTOR_KEY, payload[0].id);
            }
          }
        }
      }
    );
  }
});

export const { setActiveConnectorId } = githubConnectorSlice.actions;
export default githubConnectorSlice.reducer;
