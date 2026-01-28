/**
 * React Hooks for Better Auth Organizations
 *
 * Provides React hooks that wrap the Redux layer for organizations.
 *
 * Features:
 * - Loading/error states
 * - Refetch functionality
 * - Automatic data fetching
 * - Cache management
 *
 * Dependencies: Redux layer only (orgSlice)
 */

import { useEffect, useCallback } from 'react';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import type { AppDispatch } from '@/redux/store';
import {
  fetchUserOrganizations,
  fetchOrganizationMembers,
  invalidateOrganizationsCache,
  invalidateMembersCache,
  clearOrganizations,
  clearOrganizationMembers,
  clearAllMembers
} from '@/redux/features/users/orgSlice';
import { UserOrganization, OrganizationUsers } from '@/redux/types/orgs';

/**
 * Hook return type for user organizations
 */
export interface UseUserOrganizationsReturn {
  /** Array of user organizations */
  data: UserOrganization[];
  /** Loading state */
  isLoading: boolean;
  /** Error message if any */
  error: string | null;
  /** Refetch organizations (bypasses cache) */
  refetch: () => Promise<void>;
  /** Refetch organizations (uses cache if valid) */
  refresh: () => Promise<void>;
  /** Invalidate cache (next fetch will bypass cache) */
  invalidateCache: () => void;
  /** Clear organizations data */
  clear: () => void;
}

/**
 * Hook return type for organization members
 */
export interface UseOrganizationMembersReturn {
  /** Array of organization members */
  data: OrganizationUsers[];
  /** Loading state */
  isLoading: boolean;
  /** Error message if any */
  error: string | null;
  /** Refetch members (bypasses cache) */
  refetch: () => Promise<void>;
  /** Refetch members (uses cache if valid) */
  refresh: () => Promise<void>;
  /** Invalidate cache (next fetch will bypass cache) */
  invalidateCache: () => void;
  /** Clear members data */
  clear: () => void;
}

/**
 * Hook options for useUserOrganizations
 */
export interface UseUserOrganizationsOptions {
  /** Automatically fetch on mount (default: true) */
  autoFetch?: boolean;
  /** Skip fetching if condition is true */
  skip?: boolean;
}

/**
 * Hook options for useOrganizationMembers
 */
export interface UseOrganizationMembersOptions {
  /** Automatically fetch on mount (default: true) */
  autoFetch?: boolean;
  /** Skip fetching if condition is true */
  skip?: boolean;
}

/**
 * Hook: Get user organizations
 *
 * @param options - Hook options
 * @returns Organizations data, loading state, error, and refetch functions
 *
 * @example
 * ```tsx
 * const { data, isLoading, error, refetch } = useUserOrganizations();
 *
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error} />;
 *
 * return (
 *   <div>
 *     {data.map(org => <OrgCard key={org.id} org={org} />)}
 *   </div>
 * );
 * ```
 */
export function useUserOrganizations(
  options: UseUserOrganizationsOptions = {}
): UseUserOrganizationsReturn {
  const { autoFetch = true, skip = false } = options;
  const dispatch = useAppDispatch() as AppDispatch;

  // Get state from Redux
  const organizations = useAppSelector((state) => state.orgs.organizations);
  const isLoading = useAppSelector((state) => state.orgs.organizationsLoading);
  const error = useAppSelector((state) => state.orgs.organizationsError);

  // Auto-fetch on mount if enabled
  useEffect(() => {
    if (autoFetch && !skip && !isLoading && organizations.length === 0 && !error) {
      // Dispatch async thunk - fire and forget in useEffect
      dispatch(fetchUserOrganizations());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoFetch, skip]); // Only depend on these to avoid re-fetching

  // Refetch function (bypasses cache)
  const refetch = useCallback(async () => {
    await dispatch(fetchUserOrganizations({ force: true })).unwrap();
  }, [dispatch]);

  // Refresh function (uses cache if valid)
  const refresh = useCallback(async () => {
    await dispatch(fetchUserOrganizations()).unwrap();
  }, [dispatch]);

  // Invalidate cache
  const invalidateCache = useCallback(() => {
    dispatch(invalidateOrganizationsCache());
  }, [dispatch]);

  // Clear data
  const clear = useCallback(() => {
    dispatch(clearOrganizations());
  }, [dispatch]);

  return {
    data: organizations,
    isLoading,
    error,
    refetch,
    refresh,
    invalidateCache,
    clear
  };
}

/**
 * Hook: Get organization members
 *
 * @param organizationId - Organization ID to fetch members for
 * @param options - Hook options
 * @returns Members data, loading state, error, and refetch functions
 *
 * @example
 * ```tsx
 * const { data, isLoading, error, refetch } = useOrganizationMembers(orgId);
 *
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error} />;
 *
 * return (
 *   <div>
 *     {data.map(member => <MemberCard key={member.id} member={member} />)}
 *   </div>
 * );
 * ```
 */
export function useOrganizationMembers(
  organizationId: string | null | undefined,
  options: UseOrganizationMembersOptions = {}
): UseOrganizationMembersReturn {
  const { autoFetch = true, skip = false } = options;
  const dispatch = useAppDispatch() as AppDispatch;

  // Get state from Redux
  const members = useAppSelector((state) =>
    organizationId ? state.orgs.members[organizationId] || [] : []
  );
  const isLoading = useAppSelector((state) =>
    organizationId ? state.orgs.membersLoading[organizationId] || false : false
  );
  const error = useAppSelector((state) =>
    organizationId ? state.orgs.membersError[organizationId] || null : null
  );

  // Auto-fetch on mount if enabled
  useEffect(() => {
    if (autoFetch && !skip && organizationId && !isLoading && members.length === 0 && !error) {
      // Dispatch async thunk - fire and forget in useEffect
      dispatch(fetchOrganizationMembers({ organizationId }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoFetch, skip, organizationId]); // Only depend on these

  // Refetch function (bypasses cache)
  const refetch = useCallback(async () => {
    if (!organizationId) return;
    await dispatch(fetchOrganizationMembers({ organizationId, force: true })).unwrap();
  }, [dispatch, organizationId]);

  // Refresh function (uses cache if valid)
  const refresh = useCallback(async () => {
    if (!organizationId) return;
    await dispatch(fetchOrganizationMembers({ organizationId })).unwrap();
  }, [dispatch, organizationId]);

  // Invalidate cache
  const invalidateCache = useCallback(() => {
    if (!organizationId) return;
    dispatch(invalidateMembersCache(organizationId));
  }, [dispatch, organizationId]);

  // Clear data
  const clear = useCallback(() => {
    if (!organizationId) return;
    dispatch(clearOrganizationMembers(organizationId));
  }, [dispatch, organizationId]);

  return {
    data: members,
    isLoading,
    error,
    refetch,
    refresh,
    invalidateCache,
    clear
  };
}

/**
 * Convenience hook: Get user organizations (shorter name)
 * Alias for useUserOrganizations
 */
export const useOrgs = useUserOrganizations;

/**
 * Convenience hook: Get organization members (shorter name)
 * Alias for useOrganizationMembers
 */
export const useOrgMembers = useOrganizationMembers;
