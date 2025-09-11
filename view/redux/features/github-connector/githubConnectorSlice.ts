import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { GithubRepository } from '@/redux/types/github';
import { createSlice } from '@reduxjs/toolkit';

interface githubConnectorState {
  repositories: GithubRepository[];
}

const initialState: githubConnectorState = {
  repositories: []
};

export const githubConnectorSlice = createSlice({
  name: 'githubConnector',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addMatcher(
      GithubConnectorApi.endpoints.getAllGithubRepositories.matchFulfilled,
      (state, { payload }) => {
        if (payload && Array.isArray(payload.repositories)) {
          state.repositories = payload.repositories;
        }
      }
    );
  }
});

export default githubConnectorSlice.reducer;
