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
  clearAuthTokens
} from '@/lib/auth';

interface AuthState {
  user: User | null;
  token: string | undefined;
  refreshToken: string | undefined;
  isAuthenticated: boolean;
  isInitialized: boolean;
  isLoading: boolean;
  twoFactor: {
    isRequired: boolean;
    tempToken: string | undefined;
  };
}

interface AuthPayload {
  user: User | null;
  token: string | undefined;
  refreshToken: string | undefined;
}

export const initializeAuth = createAsyncThunk<AuthPayload | null>(
  'auth/initialize',
  async (_, { dispatch, rejectWithValue }) => {
    try {
      console.log('Starting auth initialization');
      const token = getToken();
      const refreshToken = getRefreshToken();

      if (!token || !refreshToken) {
        console.log('No tokens found, returning null');
        return null;
      }

      if (isTokenExpired(token)) {
        console.log('Token expired, attempting refresh');
        try {
          const refreshResult = await dispatch(
            authApi.endpoints.refreshToken.initiate({ refresh_token: refreshToken })
          ).unwrap();

          if (refreshResult?.access_token) {
            console.log('Token refresh successful');
            setAuthTokens({
              access_token: refreshResult.access_token,
              refresh_token: refreshResult.refresh_token,
              expires_in: refreshResult.expires_in
            });

            const userResult = await dispatch(
              authApi.endpoints.getUserDetails.initiate(undefined)
            ).unwrap();

            console.log('User details fetched successfully after refresh');
            return {
              user: userResult,
              token: refreshResult.access_token,
              refreshToken: refreshResult.refresh_token || refreshToken
            };
          } else {
            console.log('Token refresh failed - no access token received');
            return null;
          }
        } catch (refreshError) {
          console.error('Token refresh error:', refreshError);
          clearAuthTokens();
          return rejectWithValue('Token refresh failed');
        }
      } else {
        console.log('Token valid, fetching user details');
        try {
          const userResult = await dispatch(
            authApi.endpoints.getUserDetails.initiate(undefined)
          ).unwrap();

          console.log('User details fetched successfully');
          return {
            user: userResult,
            token,
            refreshToken
          };
        } catch (error) {
          console.error('Failed to fetch user details:', error);
          return rejectWithValue('Failed to fetch user details');
        }
      }
    } catch (error) {
      console.error('Auth initialization error:', error);
      return rejectWithValue('Auth initialization failed');
    }
  }
);

const initialState: AuthState = {
  user: null,
  token: undefined,
  refreshToken: undefined,
  isAuthenticated: false,
  isInitialized: false,
  isLoading: false,
  twoFactor: {
    isRequired: false,
    tempToken: undefined
  }
};

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{
        user: User | null;
        token: string;
        refreshToken?: string;
        expiresIn?: number;
        tempToken?: string;
      }>
    ) => {
      const { user, token, refreshToken, expiresIn, tempToken } = action.payload;

      if (tempToken) {
        state.twoFactor.tempToken = tempToken;
        state.twoFactor.isRequired = true;
        state.token = tempToken;
        state.isAuthenticated = false;

        setAuthTokens({
          access_token: tempToken,
          refresh_token: undefined,
          expires_in: expiresIn
        });
      } else if (token) {
        state.token = token;
        state.isAuthenticated = true;
        state.twoFactor.isRequired = false;
        state.twoFactor.tempToken = undefined;

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
      state.token = undefined;
      state.refreshToken = undefined;
      state.isAuthenticated = false;
      state.twoFactor.isRequired = false;
      state.twoFactor.tempToken = undefined;
      clearAuthTokens();
    },
    clearTwoFactor: (state) => {
      state.twoFactor.isRequired = false;
      state.twoFactor.tempToken = undefined;
    },
    setTwoFactorEnabled: (state, action: PayloadAction<boolean>) => {
      if (state.user) {
        state.user.two_factor_enabled = action.payload;
      }
    }
  },
  extraReducers: (builder) => {
    builder
      .addCase(initializeAuth.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(initializeAuth.fulfilled, (state, action) => {
        if (action.payload) {
          state.user = action.payload.user;
          state.token = action.payload.token;
          state.refreshToken = action.payload.refreshToken;
          state.isAuthenticated = true;
          state.twoFactor.isRequired = false;
          state.twoFactor.tempToken = undefined;
        }
        state.isInitialized = true;
        state.isLoading = false;
      })
      .addCase(initializeAuth.rejected, (state) => {
        state.isInitialized = true;
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.loginUser.matchPending, (state) => {
        state.isLoading = true;
      })
      .addMatcher(authApi.endpoints.loginUser.matchFulfilled, (state, { payload }) => {
        console.log('Login successful, payload:', payload);
        if (payload?.temp_token) {
          console.log('2FA required, setting temp token');
          state.twoFactor.isRequired = true;
          state.twoFactor.tempToken = payload.temp_token;
          state.token = payload.temp_token;
          state.isAuthenticated = false;

          setAuthTokens({
            access_token: payload.temp_token,
            refresh_token: undefined,
            expires_in: payload.expires_in
          });
        } else if (payload?.access_token) {
          console.log('Setting auth state with access token');
          state.user = payload.user;
          state.token = payload.access_token;
          state.refreshToken = payload.refresh_token || undefined;
          state.isAuthenticated = true;
          state.isInitialized = true;
          state.twoFactor.isRequired = false;
          state.twoFactor.tempToken = undefined;

          setAuthTokens({
            access_token: payload.access_token,
            refresh_token: payload.refresh_token || undefined,
            expires_in: payload.expires_in
          });
        } else {
          console.error('Login payload missing required tokens:', payload);
        }
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.loginUser.matchRejected, (state) => {
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.twoFactorLogin.matchPending, (state) => {
        state.isLoading = true;
      })
      .addMatcher(authApi.endpoints.twoFactorLogin.matchFulfilled, (state, { payload }) => {
        if (payload?.access_token) {
          state.user = payload.user;
          state.token = payload.access_token;
          state.refreshToken = payload.refresh_token || undefined;
          state.isAuthenticated = true;
          state.isInitialized = true;
          state.twoFactor.isRequired = false;
          state.twoFactor.tempToken = undefined;

          setAuthTokens({
            access_token: payload.access_token,
            refresh_token: payload.refresh_token || undefined,
            expires_in: payload.expires_in
          });
        }
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.twoFactorLogin.matchRejected, (state) => {
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.refreshToken.matchPending, (state) => {
        state.isLoading = true;
      })
      .addMatcher(authApi.endpoints.refreshToken.matchFulfilled, (state, { payload }) => {
        if (payload?.access_token) {
          state.token = payload.access_token;
          state.refreshToken = payload.refresh_token || undefined;
          state.isAuthenticated = true;

          setAuthTokens({
            access_token: payload.access_token,
            refresh_token: payload.refresh_token || undefined,
            expires_in: payload.expires_in
          });
        }
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.refreshToken.matchRejected, (state) => {
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.getUserDetails.matchPending, (state) => {
        state.isLoading = true;
      })
      .addMatcher(authApi.endpoints.getUserDetails.matchFulfilled, (state, { payload }) => {
        if (payload) {
          state.user = payload;
        }
        state.isLoading = false;
      })
      .addMatcher(authApi.endpoints.getUserDetails.matchRejected, (state) => {
        state.isLoading = false;
      })
      .addMatcher(userApi.endpoints.updateUserName.matchFulfilled, (state, { payload }) => {
        if (payload && state.user) {
          state.user.username = payload;
        }
      });
  }
});

export const { setCredentials, logout, clearTwoFactor, setTwoFactorEnabled } = authSlice.actions;
export default authSlice.reducer;
