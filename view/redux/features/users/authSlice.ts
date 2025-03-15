import { authApi } from '@/redux/services/users/authApi';
import { userApi } from '@/redux/services/users/userApi';
import { User } from '@/redux/types/user';
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { createAsyncThunk } from '@reduxjs/toolkit';
import {
  getToken,
  getRefreshToken,
  isTokenExpired,
  setAuthTokens,
  clearAuthTokens,
} from '@/lib/auth';

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isInitialized: boolean;
}

interface AuthPayload {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
}

export const initializeAuth = createAsyncThunk<AuthPayload | null>(
  'auth/initialize',
  async (_, { dispatch, rejectWithValue }) => {
    try {
      const token = getToken();
      const refreshToken = getRefreshToken();

      if (!token || !refreshToken) {
        return null;
      }

      if (isTokenExpired(token)) {
        try {
          const refreshResult = await dispatch(
            authApi.endpoints.refreshToken.initiate({ refresh_token: refreshToken })
          ).unwrap();

          if (refreshResult?.access_token) {
            setAuthTokens({
              access_token: refreshResult.access_token,
              refresh_token: refreshResult.refresh_token,
              expires_in: refreshResult.expires_in
            });

            const userResult = await dispatch(
              authApi.endpoints.getUserDetails.initiate(undefined)
            ).unwrap();

            return {
              user: userResult,
              token: refreshResult.access_token,
              refreshToken: refreshResult.refresh_token || refreshToken,
            };
          } else {
            return null;
          }
        } catch (refreshError) {
          clearAuthTokens();
          return rejectWithValue('Token refresh failed');
        }
      } else {
        try {
          const userResult = await dispatch(
            authApi.endpoints.getUserDetails.initiate(undefined)
          ).unwrap();

          return {
            user: userResult,
            token,
            refreshToken,
          };
        } catch (error) {
          return rejectWithValue('Failed to fetch user details');
        }
      }
    } catch (error) {
      return rejectWithValue('Auth initialization failed');
    }
  }
);

const initialState: AuthState = {
  user: null,
  token: null,
  refreshToken: null,
  isAuthenticated: false,
  isInitialized: false
};

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{ user: User | null; token: string; refreshToken?: string; expiresIn?: number }>
    ) => {
      const { user, token, refreshToken, expiresIn } = action.payload;

      if (token) {
        state.token = token;
        state.isAuthenticated = true;

        setAuthTokens({
          access_token: token,
          refresh_token: refreshToken,
          expires_in: expiresIn
        });
      }

      if (refreshToken) {
        state.refreshToken = refreshToken;
      }

      state.user = user;
    },
    logout: (state) => {
      state.user = null;
      state.token = null;
      state.refreshToken = null;
      state.isAuthenticated = false;
      clearAuthTokens();
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(initializeAuth.fulfilled, (state, action) => {
        if (action.payload) {
          state.user = action.payload.user;
          state.token = action.payload.token;
          state.refreshToken = action.payload.refreshToken;
          state.isAuthenticated = true;
        }
        state.isInitialized = true;
      })
      .addCase(initializeAuth.rejected, (state) => {
        state.isInitialized = true;
      })
      .addMatcher(authApi.endpoints.loginUser.matchFulfilled, (state, { payload }) => {
        if (payload?.access_token) {
          state.user = payload.user;
          state.token = payload.access_token;
          state.refreshToken = payload.refresh_token || null;
          state.isAuthenticated = true;

          setAuthTokens({
            access_token: payload.access_token,
            refresh_token: payload.refresh_token,
            expires_in: payload.expires_in
          });
        }
      })
      .addMatcher(authApi.endpoints.refreshToken.matchFulfilled, (state, { payload }) => {
        if (payload?.access_token) {
          state.token = payload.access_token;

          if (payload.refresh_token) {
            state.refreshToken = payload.refresh_token;
          }

          setAuthTokens({
            access_token: payload.access_token,
            refresh_token: payload.refresh_token,
            expires_in: payload.expires_in
          });
        }
      })
      .addMatcher(authApi.endpoints.getUserDetails.matchFulfilled, (state, { payload }) => {
        if (payload) {
          state.user = payload;
        }
      })
      .addMatcher(userApi.endpoints.updateUserName.matchFulfilled, (state, { payload }) => {
        if (payload && state.user) {
          state.user.username = payload;
        }
      });
  }
});

export const { setCredentials, logout } = authSlice.actions;
export default authSlice.reducer;