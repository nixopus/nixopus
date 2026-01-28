/**
 * Redux Slice for Better Auth Organizations
 *
 * Manages:
 * - User organizations state
 * - Organization members state
 * - Loading/error states
 * - Caching (5-minute cache)
 * - Cache invalidation
 *
 * Dependencies: Only better-auth-orgs service layer
 */

import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import {
  getUserOrganizations,
  getOrganizationMembers,
  BetterAuthOrgError
} from '@/packages/lib/better-auth-orgs';
import { UserOrganization, OrganizationUsers } from '@/redux/types/orgs';

/**
 * Cache duration: 5 minutes (in milliseconds)
 */
const CACHE_DURATION = 5 * 60 * 1000;

/**
 * State interface
 */
interface OrgState {
  // Organizations
  organizations: UserOrganization[];
  organizationsLoading: boolean;
  organizationsError: string | null;
  organizationsLastFetched: number | null;

  // Members (keyed by organization ID)
  members: Record<string, OrganizationUsers[]>;
  membersLoading: Record<string, boolean>;
  membersError: Record<string, string | null>;
  membersLastFetched: Record<string, number | null>;
}

const initialState: OrgState = {
  organizations: [],
  organizationsLoading: false,
  organizationsError: null,
  organizationsLastFetched: null,
  members: {},
  membersLoading: {},
  membersError: {},
  membersLastFetched: {}
};

/**
 * Check if cache is still valid
 */
function isCacheValid(lastFetched: number | null): boolean {
  if (!lastFetched) return false;
  return Date.now() - lastFetched < CACHE_DURATION;
}

/**
 * Async thunk: Fetch user organizations
 *
 * Uses caching - won't refetch if cache is still valid unless force=true
 */
export const fetchUserOrganizations = createAsyncThunk<
  UserOrganization[],
  { force?: boolean } | void,
  { rejectValue: string }
>('orgs/fetchUserOrganizations', async (options, { getState, rejectWithValue }) => {
  try {
    const state = getState() as { orgs: OrgState };
    const orgState = state.orgs;

    // Check cache unless force refresh
    const force = typeof options === 'object' ? options.force : false;
    if (!force && isCacheValid(orgState.organizationsLastFetched)) {
      // Return cached data
      return orgState.organizations;
    }

    // Fetch from service layer
    const organizations = await getUserOrganizations();
    return organizations;
  } catch (error) {
    const errorMessage =
      error instanceof BetterAuthOrgError
        ? error.message
        : error instanceof Error
          ? error.message
          : 'Failed to fetch organizations';
    return rejectWithValue(errorMessage);
  }
});

/**
 * Async thunk: Fetch organization members
 *
 * Uses caching - won't refetch if cache is still valid unless force=true
 */
export const fetchOrganizationMembers = createAsyncThunk<
  OrganizationUsers[],
  { organizationId: string; force?: boolean },
  { rejectValue: string }
>(
  'orgs/fetchOrganizationMembers',
  async ({ organizationId, force = false }, { getState, rejectWithValue }) => {
    try {
      if (!organizationId) {
        return rejectWithValue('Organization ID is required');
      }

      const state = getState() as { orgs: OrgState };
      const orgState = state.orgs;

      // Check cache unless force refresh
      const lastFetched = orgState.membersLastFetched[organizationId];
      if (!force && isCacheValid(lastFetched)) {
        // Return cached data
        return orgState.members[organizationId] || [];
      }

      // Fetch from service layer
      const members = await getOrganizationMembers(organizationId);
      return members;
    } catch (error) {
      const errorMessage =
        error instanceof BetterAuthOrgError
          ? error.message
          : error instanceof Error
            ? error.message
            : 'Failed to fetch organization members';
      return rejectWithValue(errorMessage);
    }
  }
);

/**
 * Redux slice
 */
export const orgSlice = createSlice({
  name: 'orgs',
  initialState,
  reducers: {
    /**
     * Clear organizations cache and data
     */
    clearOrganizations: (state) => {
      state.organizations = [];
      state.organizationsLastFetched = null;
      state.organizationsError = null;
    },

    /**
     * Clear members cache for a specific organization
     */
    clearOrganizationMembers: (state, action: PayloadAction<string>) => {
      const orgId = action.payload;
      delete state.members[orgId];
      delete state.membersLastFetched[orgId];
      delete state.membersError[orgId];
      delete state.membersLoading[orgId];
    },

    /**
     * Clear all members cache
     */
    clearAllMembers: (state) => {
      state.members = {};
      state.membersLastFetched = {};
      state.membersError = {};
      state.membersLoading = {};
    },

    /**
     * Invalidate organizations cache (force next fetch)
     */
    invalidateOrganizationsCache: (state) => {
      state.organizationsLastFetched = null;
    },

    /**
     * Invalidate members cache for a specific organization
     */
    invalidateMembersCache: (state, action: PayloadAction<string>) => {
      const orgId = action.payload;
      delete state.membersLastFetched[orgId];
    },

    /**
     * Set organizations directly (useful for optimistic updates)
     */
    setOrganizations: (state, action: PayloadAction<UserOrganization[]>) => {
      state.organizations = action.payload;
      state.organizationsLastFetched = Date.now();
    }
  },
  extraReducers: (builder) => {
    // Fetch user organizations
    builder
      .addCase(fetchUserOrganizations.pending, (state) => {
        state.organizationsLoading = true;
        state.organizationsError = null;
      })
      .addCase(fetchUserOrganizations.fulfilled, (state, action) => {
        state.organizationsLoading = false;
        state.organizations = action.payload;
        state.organizationsLastFetched = Date.now();
        state.organizationsError = null;
      })
      .addCase(fetchUserOrganizations.rejected, (state, action) => {
        state.organizationsLoading = false;
        state.organizationsError = action.payload || 'Failed to fetch organizations';
      });

    // Fetch organization members
    builder
      .addCase(fetchOrganizationMembers.pending, (state, action) => {
        const orgId = action.meta.arg.organizationId;
        state.membersLoading[orgId] = true;
        state.membersError[orgId] = null;
      })
      .addCase(fetchOrganizationMembers.fulfilled, (state, action) => {
        const orgId = action.meta.arg.organizationId;
        state.membersLoading[orgId] = false;
        state.members[orgId] = action.payload;
        state.membersLastFetched[orgId] = Date.now();
        state.membersError[orgId] = null;
      })
      .addCase(fetchOrganizationMembers.rejected, (state, action) => {
        const orgId = action.meta.arg.organizationId;
        state.membersLoading[orgId] = false;
        state.membersError[orgId] = action.payload || 'Failed to fetch organization members';
      });
  }
});

export const {
  clearOrganizations,
  clearOrganizationMembers,
  clearAllMembers,
  invalidateOrganizationsCache,
  invalidateMembersCache,
  setOrganizations
} = orgSlice.actions;

export default orgSlice.reducer;
