import { useAppSelector } from '@/redux/hooks';
import { hasPermission } from './permission';

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
  | 'update';

export type Action = 'create' | 'read' | 'update' | 'delete';

export type Permission = `${Resource}:${Action}`;

export const useRBAC = () => {
  const user = useAppSelector((state) => state.auth.user);
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);

  const canAccessResource = (resource: Resource, action: Action): boolean => {
    if (!user || !activeOrganization) return false;
    return hasPermission(user, resource, action, activeOrganization.id);
  };

  const hasPermissionCheck = (permission: Permission): boolean => {
    if (!user || !activeOrganization) return false;
    const [resource, action] = permission.split(':') as [Resource, Action];
    return hasPermission(user, resource, action, activeOrganization.id);
  };

  const hasAnyPermission = (permissions: Permission[]): boolean => {
    return permissions.some(hasPermissionCheck);
  };

  const hasAllPermissions = (permissions: Permission[]): boolean => {
    return permissions.every(hasPermissionCheck);
  };

  const isLoading = !user || !activeOrganization;

  return {
    canAccessResource,
    hasPermission: hasPermissionCheck,
    hasAnyPermission,
    hasAllPermissions,
    isLoading
  };
}; 