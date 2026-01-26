import { useMemo } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { useGetActiveMemberQuery } from '@/redux/services/users/userApi';

export type Resource =
  | 'organization'
  | 'user'
  | 'role'
  | 'permission'
  | 'deploy'
  | 'file-manager'
  | 'dashboard'
  | 'settings'
  | 'audit'
  | 'notification'
  | 'domain'
  | 'feature-flags'
  | 'github-connector'
  | 'container'
  | 'terminal'
  | 'update'
  | 'extensions';

export type Action = 'create' | 'read' | 'update' | 'delete';

export type Permission = `${Resource}:${Action}`;

export const useRBAC = () => {
  const { isAuthenticated, isInitialized, user } = useAppSelector((state) => state.auth);
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  
  // Get role from Better Auth for the current organization
  const { data: activeMemberData, isLoading: isMemberLoading, error } = useGetActiveMemberQuery(
    {
      organizationId: activeOrganization?.id || '',
      userId: user?.id || ''
    },
    {
      skip: !isAuthenticated || !activeOrganization?.id || !user?.id,
      refetchOnMountOrArgChange: true
    }
  );

  // Extract roles from Better Auth member data
  const roles = useMemo(() => {
    if (activeMemberData?.role) {
      return Array.isArray(activeMemberData.role) ? activeMemberData.role : [activeMemberData.role];
    }
    // Fallback: if no member data but we have an active org, assume member role
    if (activeOrganization?.id && user?.id && !error) {
      return ['member'];
    }
    return undefined;
  }, [activeMemberData?.role, activeOrganization?.id, user?.id, error]);

  // Check if user is admin/owner
  const isAdmin = useMemo(() => {
    if (!Array.isArray(roles)) return false;
    return roles.some((role) => {
      const roleStr = typeof role === 'string' ? role.toLowerCase() : '';
      // Strip orgid_ prefix to get base role name
      if (roleStr.startsWith('orgid_')) {
        const lastUnderscore = roleStr.lastIndexOf('_');
        if (lastUnderscore !== -1 && lastUnderscore < roleStr.length - 1) {
          const baseRole = roleStr.substring(lastUnderscore + 1);
          return baseRole === 'admin' || baseRole === 'owner';
        }
      }
      return roleStr === 'admin' || roleStr === 'owner';
    });
  }, [roles]);

  const isLoading = !isInitialized || (isMemberLoading && activeOrganization?.id);

  const canAccessResource = (resource: Resource, action: Action): boolean => {
    // Admins have access to everything - no permission check needed
    if (isAdmin) return true;
    // For non-admin users, allow access for now (no permission checks)
    return true;
  };

  const hasPermissionCheck = (permission: Permission): boolean => {
    // Admins have all permissions - no check needed
    if (isAdmin) return true;
    // For non-admin users, allow access for now (no permission checks)
    return true;
  };

  const hasAnyPermission = (permissions: Permission[]): boolean => {
    // Admins have all permissions
    if (isAdmin) return true;
    // For non-admin users, allow access for now (no permission checks)
    return true;
  };

  const hasAllPermissions = (permissions: Permission[]): boolean => {
    // Admins have all permissions
    if (isAdmin) return true;
    // For non-admin users, allow access for now (no permission checks)
    return true;
  };

  return {
    canAccessResource,
    hasPermission: hasPermissionCheck,
    hasAnyPermission,
    hasAllPermissions,
    isLoading,
    roles,
    permissions: undefined, // No API permissions needed
    isAdmin
  };
};
