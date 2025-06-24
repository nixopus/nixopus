import { User } from '@/redux/types/user';

/**
 * Check if a user has a specific permission for a resource
 *
 * @param user - The user object from Redux state
 * @param resource - The resource type (e.g., "organization", "user", "role", etc.)
 * @param action - The permission action (e.g., "read", "create", "update", "delete")
 * @param organizationId - Optional organization ID to check permissions for a specific organization
 * @returns boolean indicating if the user has the specified permission
 */
export const hasPermission = (
  user: User | null | undefined,
  resource: string,
  action: string,
  organizationId: string
): boolean => {
  if (!user || !user.organization_users) return false;

  return user.organization_users.some((orgUser) => {
    if (organizationId && orgUser.organization_id !== organizationId) {
      return false;
    }

    return orgUser.role.permissions.some(
      (permission) => permission.resource === resource && permission.name === action
    );
  });
};