/**
 * Better Auth Organizations Service Layer
 *
 * Pure API layer for Better Auth organization operations.
 * No React/Redux dependencies - can be used anywhere.
 *
 * This layer handles:
 * - API calls to Better Auth
 * - Data transformation (Better Auth format → App types)
 * - Error handling and type safety
 */

import { authClient } from './auth-client';
import { UserOrganization, OrganizationUsers } from '@/redux/types/orgs';
import { User } from '@/redux/types/user';

/**
 * Better Auth Organization Response Types
 */
interface BetterAuthOrganization {
  id: string;
  name: string;
  slug?: string;
  logo?: string | null;
  metadata?: {
    description?: string;
    [key: string]: any;
  };
  createdAt: string;
  updatedAt: string;
  role?: string;
}

interface BetterAuthMember {
  id: string;
  userId: string;
  organizationId: string;
  role: string | string[];
  permissions?: string[];
  createdAt: string;
  updatedAt: string;
  user?: {
    id: string;
    email: string;
    name?: string;
    username?: string;
    image?: string;
    avatar?: string;
    emailVerified?: boolean;
    twoFactorEnabled?: boolean;
    createdAt?: string;
    updatedAt?: string;
  };
}

/**
 * Service Error Types
 */
export class BetterAuthOrgError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public originalError?: unknown
  ) {
    super(message);
    this.name = 'BetterAuthOrgError';
  }
}

/**
 * Transform Better Auth organization to UserOrganization format
 */
function transformOrganization(org: BetterAuthOrganization): UserOrganization {
  return {
    id: org.id,
    organization: {
      id: org.id,
      name: org.name,
      description: org.metadata?.description || '',
      created_at: org.createdAt || new Date().toISOString(),
      updated_at: org.updatedAt || new Date().toISOString(),
      deleted_at: null
    },
    role: {
      id: org.role || 'member',
      name: org.role || 'member',
      description: ''
    },
    created_at: org.createdAt || new Date().toISOString(),
    updated_at: org.updatedAt || new Date().toISOString(),
    deleted_at: null
  };
}

/**
 * Transform Better Auth member to OrganizationUsers format
 */
function transformMember(member: BetterAuthMember): OrganizationUsers {
  const user = member.user || ({} as any);
  const role = member.role || 'member';

  return {
    id: member.id,
    user_id: member.userId || user?.id || '',
    organization_id: member.organizationId,
    created_at: member.createdAt || new Date().toISOString(),
    updated_at: member.updatedAt || new Date().toISOString(),
    deleted_at: null,
    user: {
      id: user?.id || member.userId || '',
      email: user?.email || '',
      username: user?.name || user?.username || '',
      avatar: user?.image || user?.avatar || '',
      type: Array.isArray(role) ? role[0] : role,
      organization_users: [],
      is_verified: user.emailVerified || false,
      is_email_verified: user.emailVerified || false,
      two_factor_enabled: user.twoFactorEnabled || false,
      two_factor_secret: '',
      created_at: user.createdAt || new Date().toISOString(),
      updated_at: user.updatedAt || new Date().toISOString(),
      organizations: []
    } as User,
    roles: Array.isArray(role) ? role : [role],
    permissions: member.permissions || []
  };
}

/**
 * Get user's organizations
 *
 * Uses Better Auth REST API through the Next.js proxy.
 * Falls back to direct API call if proxy is not available.
 *
 * @returns Array of user organizations
 * @throws BetterAuthOrgError if request fails
 */
export async function getUserOrganizations(): Promise<UserOrganization[]> {
  try {
    // Check if we're in browser environment
    const isBrowser = typeof window !== 'undefined';

    if (isBrowser) {
      // Use Next.js proxy route
      const baseUrl = window.location.origin;
      const response = await fetch(`${baseUrl}/api/auth/organization/list`, {
        method: 'GET',
        credentials: 'include', // Include cookies for authentication
        headers: {
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error');
        throw new BetterAuthOrgError(
          `Failed to fetch organizations: ${errorText}`,
          response.status
        );
      }

      const data: BetterAuthOrganization[] = await response.json();

      // Handle both array and object with data property
      const organizations = Array.isArray(data) ? data : (data as any).data || [];

      return organizations.map(transformOrganization);
    } else {
      // Server-side: use auth service URL directly
      const authServiceUrl = process.env.AUTH_SERVICE_URL || 'http://localhost:9090';
      const response = await fetch(`${authServiceUrl}/api/auth/organization/list`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error');
        throw new BetterAuthOrgError(
          `Failed to fetch organizations: ${errorText}`,
          response.status
        );
      }

      const data: BetterAuthOrganization[] = await response.json();
      const organizations = Array.isArray(data) ? data : (data as any).data || [];

      return organizations.map(transformOrganization);
    }
  } catch (error) {
    if (error instanceof BetterAuthOrgError) {
      throw error;
    }

    throw new BetterAuthOrgError(
      `Unexpected error fetching organizations: ${error instanceof Error ? error.message : String(error)}`,
      undefined,
      error
    );
  }
}

/**
 * Get organization members
 *
 * @param organizationId - Organization ID to fetch members for
 * @returns Array of organization members
 * @throws BetterAuthOrgError if request fails
 */
export async function getOrganizationMembers(organizationId: string): Promise<OrganizationUsers[]> {
  if (!organizationId) {
    throw new BetterAuthOrgError('Organization ID is required');
  }

  try {
    const isBrowser = typeof window !== 'undefined';

    if (isBrowser) {
      // Use Next.js proxy route
      const baseUrl = window.location.origin;
      const response = await fetch(
        `${baseUrl}/api/auth/organization/list-members?organizationId=${organizationId}`,
        {
          method: 'GET',
          credentials: 'include', // Include cookies for authentication
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error');
        throw new BetterAuthOrgError(
          `Failed to fetch organization members: ${errorText}`,
          response.status
        );
      }

      const data: BetterAuthMember[] = await response.json();

      // Handle both array and object with data property
      const members = Array.isArray(data) ? data : (data as any).data || [];

      return members.map(transformMember);
    } else {
      // Server-side: use auth service URL directly
      const authServiceUrl = process.env.AUTH_SERVICE_URL || 'http://localhost:9090';
      const response = await fetch(
        `${authServiceUrl}/api/auth/organization/list-members?organizationId=${organizationId}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error');
        throw new BetterAuthOrgError(
          `Failed to fetch organization members: ${errorText}`,
          response.status
        );
      }

      const data: BetterAuthMember[] = await response.json();
      const members = Array.isArray(data) ? data : (data as any).data || [];

      return members.map(transformMember);
    }
  } catch (error) {
    if (error instanceof BetterAuthOrgError) {
      throw error;
    }

    throw new BetterAuthOrgError(
      `Unexpected error fetching organization members: ${error instanceof Error ? error.message : String(error)}`,
      undefined,
      error
    );
  }
}

/**
 * Service layer exports
 */
export const betterAuthOrgsService = {
  getUserOrganizations,
  getOrganizationMembers
};
