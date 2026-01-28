/**
 * Verification Tests for Better Auth Organizations Service
 *
 * These tests verify Phase 1 implementation:
 * - Service functions can be called independently
 * - Data transformation works correctly
 * - Errors are handled and typed properly
 * - No React/Redux imports in service layer
 *
 * To run: These are manual verification tests - run in browser console or Node.js
 */

import {
  getUserOrganizations,
  getOrganizationMembers,
  BetterAuthOrgError
} from '../better-auth-orgs';

/**
 * Manual Verification Test 1: Service functions can be called independently
 *
 * Run in browser console after logging in:
 *
 * ```javascript
 * import { getUserOrganizations } from '@/packages/lib/better-auth-orgs';
 * const orgs = await getUserOrganizations();
 * console.log('Organizations:', orgs);
 * ```
 *
 * Expected: Array of UserOrganization objects
 */
export async function testGetUserOrganizations() {
  try {
    console.log('Testing getUserOrganizations...');
    const organizations = await getUserOrganizations();
    console.log('✅ Success! Organizations:', organizations);
    return organizations;
  } catch (error) {
    console.error('❌ Error:', error);
    if (error instanceof BetterAuthOrgError) {
      console.error('Status:', error.statusCode);
      console.error('Original:', error.originalError);
    }
    throw error;
  }
}

/**
 * Manual Verification Test 2: Data transformation works correctly
 *
 * Run in browser console:
 *
 * ```javascript
 * import { getUserOrganizations } from '@/packages/lib/better-auth-orgs';
 * const orgs = await getUserOrganizations();
 *
 * // Verify structure
 * console.log('First org structure:', {
 *   hasId: !!orgs[0]?.id,
 *   hasOrganization: !!orgs[0]?.organization,
 *   hasRole: !!orgs[0]?.role,
 *   orgId: orgs[0]?.organization?.id,
 *   orgName: orgs[0]?.organization?.name,
 * });
 * ```
 */
export function verifyOrganizationStructure(org: any) {
  const checks = {
    hasId: typeof org?.id === 'string',
    hasOrganization: typeof org?.organization === 'object',
    hasRole: typeof org?.role === 'object',
    hasCreatedAt: typeof org?.created_at === 'string',
    hasUpdatedAt: typeof org?.updated_at === 'string',
    orgHasId: typeof org?.organization?.id === 'string',
    orgHasName: typeof org?.organization?.name === 'string',
    roleHasId: typeof org?.role?.id === 'string',
    roleHasName: typeof org?.role?.name === 'string'
  };

  const allPassed = Object.values(checks).every(Boolean);
  console.log('Structure checks:', checks);
  console.log(allPassed ? '✅ All checks passed!' : '❌ Some checks failed');
  return allPassed;
}

/**
 * Manual Verification Test 3: Error handling
 *
 * Run in browser console (without being logged in):
 *
 * ```javascript
 * import { getUserOrganizations, BetterAuthOrgError } from '@/packages/lib/better-auth-orgs';
 * try {
 *   await getUserOrganizations();
 * } catch (error) {
 *   console.log('Error type:', error instanceof BetterAuthOrgError);
 *   console.log('Error message:', error.message);
 *   console.log('Status code:', error.statusCode);
 * }
 * ```
 */
export function verifyErrorHandling(error: unknown) {
  const isBetterAuthError = error instanceof BetterAuthOrgError;
  const hasMessage = error instanceof Error && error.message.length > 0;
  const hasStatusCode = error instanceof BetterAuthOrgError && error.statusCode !== undefined;

  console.log('Error handling checks:', {
    isBetterAuthError,
    hasMessage,
    hasStatusCode
  });

  return isBetterAuthError && hasMessage;
}

/**
 * Manual Verification Test 4: Get organization members
 *
 * Run in browser console:
 *
 * ```javascript
 * import { getOrganizationMembers } from '@/packages/lib/better-auth-orgs';
 * const orgs = await getUserOrganizations();
 * const members = await getOrganizationMembers(orgs[0].organization.id);
 * console.log('Members:', members);
 * ```
 */
export async function testGetOrganizationMembers(organizationId: string) {
  try {
    console.log('Testing getOrganizationMembers for org:', organizationId);
    const members = await getOrganizationMembers(organizationId);
    console.log('✅ Success! Members:', members);
    return members;
  } catch (error) {
    console.error('❌ Error:', error);
    if (error instanceof BetterAuthOrgError) {
      console.error('Status:', error.statusCode);
    }
    throw error;
  }
}

/**
 * Verification Checklist
 *
 * Run these in order:
 *
 * 1. ✅ Service functions can be called independently
 *    - Import and call getUserOrganizations()
 *    - Should return array without errors
 *
 * 2. ✅ Data transformation works correctly
 *    - Check structure of returned organizations
 *    - Verify all required fields are present
 *    - Verify types match UserOrganization interface
 *
 * 3. ✅ Errors are handled and typed properly
 *    - Test with invalid session (logout first)
 *    - Verify BetterAuthOrgError is thrown
 *    - Verify error has message and statusCode
 *
 * 4. ✅ No React/Redux imports in service layer
 *    - Check better-auth-orgs.ts imports
 *    - Should only import from:
 *      - './auth-client'
 *      - '@/redux/types/orgs' (types only)
 *      - No React, Redux, or hooks
 *
 * 5. ✅ Organization members work
 *    - Call getOrganizationMembers() with valid org ID
 *    - Verify members array structure
 *    - Verify transformation to OrganizationUsers format
 */
