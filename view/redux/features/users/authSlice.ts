import { authApi } from '@/redux/services/users/authApi';
import { userApi } from '@/redux/services/users/userApi';
import { User } from '@/redux/types/user';
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { createAsyncThunk } from '@reduxjs/toolkit';
import { authClient } from '@/packages/lib/auth-client';
import { setActiveOrganization } from './userSlice';
import { fetchUserOrganizations } from './orgSlice';

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

export const initializeAuth = createAsyncThunk<AuthPayload | null, void, { rejectValue: string }>(
  'auth/initialize',
  async (_, { dispatch, rejectWithValue }) => {
    try {
      const sessionResult = await authClient.getSession();

      if (!sessionResult?.data?.session) {
        return null;
      }

      try {
        const userResult = await dispatch(
          authApi.endpoints.getUserDetails.initiate(undefined)
        ).unwrap();

        try {
          // Use new Better Auth organizations service
          const organizationsResult = await dispatch(fetchUserOrganizations()).unwrap();

          if (organizationsResult && organizationsResult.length > 0) {
            const firstOrg = organizationsResult[0];
            dispatch(setActiveOrganization(firstOrg.organization));
          }
        } catch (orgError: any) {
          // Don't fail auth if organizations can't be loaded
        }

        return {
          user: userResult,
          token: sessionResult.data.session.token || '',
          refreshToken: undefined
        };
      } catch (error: any) {
        return rejectWithValue('Failed to fetch user details');
      }
    } catch (error: any) {
      return rejectWithValue('Auth initialization failed');
    }
  }
);

export const logoutUser = createAsyncThunk('auth/logoutUser', async (_, { dispatch }) => {
  try {
    await authClient.signOut();
    dispatch(logout());
  } catch (error) {
    console.error('Better Auth logout failed:', error);
    dispatch(logout());
  }
});

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
      } else if (token) {
        state.token = token;
        state.isAuthenticated = true;
        state.twoFactor.isRequired = false;
        state.twoFactor.tempToken = undefined;
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
        if (payload?.temp_token) {
          state.twoFactor.isRequired = true;
          state.twoFactor.tempToken = payload.temp_token;
          state.token = payload.temp_token;
          state.isAuthenticated = false;
        } else if (payload?.access_token) {
          state.user = payload.user;
          state.token = payload.access_token;
          state.refreshToken = payload.refresh_token || undefined;
          state.isAuthenticated = true;
          state.isInitialized = true;
          state.twoFactor.isRequired = false;
          state.twoFactor.tempToken = undefined;
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
