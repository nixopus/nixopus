import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';

// ---------------------------------------------------------------------------
// Mocks – must be declared before the module under test is imported so that
// vitest's hoisting mechanism replaces the real modules.
// NOTE: vi.mock() factories are hoisted before any variable declarations, so
// no helper variables may be referenced inside them.
// ---------------------------------------------------------------------------

vi.mock('next/navigation', () => ({
  usePathname: () => '/'
}));

const mockDispatch = vi.fn();
vi.mock('@/redux/hooks', () => ({
  useAppSelector: (selector: (s: any) => any) =>
    selector({
      auth: { user: { id: 'user-1', name: 'Test User' } },
      user: { activeOrganization: { id: 'org-1', name: 'Test Org' } }
    }),
  useAppDispatch: () => mockDispatch
}));

vi.mock('@/packages/hooks/auth/use-better-auth-orgs', () => ({
  useUserOrganizations: () => ({
    data: [{ organization: { id: 'org-1', name: 'Test Org' } }],
    isLoading: false,
    refetch: vi.fn()
  })
}));

vi.mock('@/packages/hooks/shared/use_navigation_state', () => ({
  useNavigationState: () => ({
    activeNav: '/apps',
    setActiveNav: vi.fn()
  })
}));

vi.mock('@/packages/hooks/shared/use-translation', () => ({
  useTranslation: () => ({
    // Return the key unchanged so tests can match on raw translation keys.
    t: (key: string) => key
  })
}));

vi.mock('@/packages/utils/rbac', () => ({
  useRBAC: () => ({
    canAccessResource: () => true
  })
}));

// Redux slices
vi.mock('@/redux/features/users/userSlice', () => ({
  setActiveOrganization: (org: any) => ({ type: 'user/setActiveOrganization', payload: org })
}));

vi.mock('@/redux/features/users/authSlice', () => ({
  logout: () => ({ type: 'auth/logout' }),
  logoutUser: () => () => Promise.resolve()
}));

// Redux API services – each needs a minimal { util: { resetApiState } } shape.
vi.mock('@/redux/services/users/authApi', () => ({
  authApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/users/userApi', () => ({
  userApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/settings/notificationApi', () => ({
  notificationApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/settings/domainsApi', () => ({
  domainsApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/connector/githubConnectorApi', () => ({
  GithubConnectorApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/deploy/applicationsApi', () => ({
  deployApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/audit', () => ({
  auditApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/feature-flags/featureFlagsApi', () => ({
  FeatureFlagsApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/api-keys/apiKeysApi', () => ({
  apiKeysApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/machine/machineBackupApi', () => ({
  machineBackupApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));
vi.mock('@/redux/services/domains/customDomainsApi', () => ({
  customDomainsApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));

// Plugin registry – return an empty list by default so that only core items
// are exercised in these tests.
vi.mock('@/plugins/registry', () => ({
  getPluginNavItems: vi.fn(() => [])
}));

// ---------------------------------------------------------------------------
// The module under test
// ---------------------------------------------------------------------------
import { useAppSidebar } from '../use-app-sidebar';
import { getPluginNavItems } from '@/plugins/registry';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function renderSidebar() {
  const { result } = renderHook(() => useAppSidebar());
  return result.current;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useAppSidebar – coreNavItems after backups removal', () => {
  beforeEach(() => {
    mockDispatch.mockClear();
    vi.mocked(getPluginNavItems).mockReturnValue([]);
  });

  // ── Removal assertions ──────────────────────────────────────────────────

  it('does not include a nav item with url "/backups"', () => {
    const { filteredNavItems } = renderSidebar();
    const urls = filteredNavItems.map((item) => item.url);
    expect(urls).not.toContain('/backups');
  });

  it('does not include a nav item with title key "navigation.backups"', () => {
    const { filteredNavItems } = renderSidebar();
    // useTranslation mock returns the key unchanged, so the raw key is visible.
    const titles = filteredNavItems.map((item) => item.title);
    expect(titles).not.toContain('navigation.backups');
  });

  it('does not include any item whose resource is "backup"', () => {
    const { filteredNavItems } = renderSidebar();
    const resources = filteredNavItems.map((item) => (item as any).resource);
    expect(resources).not.toContain('backup');
  });

  // ── Presence assertions – remaining core items ───────────────────────────

  it('still includes the selfHost nav item (/apps)', () => {
    const { filteredNavItems } = renderSidebar();
    const urls = filteredNavItems.map((item) => item.url);
    expect(urls).toContain('/apps');
  });

  it('still includes the machines nav item (/machines)', () => {
    const { filteredNavItems } = renderSidebar();
    const urls = filteredNavItems.map((item) => item.url);
    expect(urls).toContain('/machines');
  });

  it('still includes the integrations nav item (/integrations)', () => {
    const { filteredNavItems } = renderSidebar();
    const urls = filteredNavItems.map((item) => item.url);
    expect(urls).toContain('/integrations');
  });

  it('still includes the domains nav item (/domains)', () => {
    const { filteredNavItems } = renderSidebar();
    const urls = filteredNavItems.map((item) => item.url);
    expect(urls).toContain('/domains');
  });

  // ── Settings group presence ──────────────────────────────────────────────

  it('includes a grouped Settings item containing apiKeys sub-item', () => {
    const { filteredNavItems } = renderSidebar();
    const settingsGroup = filteredNavItems.find((item) => 'items' in item && item.items) as any;
    expect(settingsGroup).toBeDefined();
    const subUrls = settingsGroup.items.map((s: any) => s.url);
    expect(subUrls).toContain('/api-keys');
  });

  it('includes a grouped Settings item containing security sub-item', () => {
    const { filteredNavItems } = renderSidebar();
    const settingsGroup = filteredNavItems.find((item) => 'items' in item && item.items) as any;
    expect(settingsGroup).toBeDefined();
    const subUrls = settingsGroup.items.map((s: any) => s.url);
    expect(subUrls).toContain('/security');
  });

  // ── Order validation – no gap where order:50 used to be ─────────────────

  it('top-level items are sorted in ascending order without a gap at order 50', () => {
    const { filteredNavItems } = renderSidebar();
    // Exclude grouped items (Settings group has sub-items).
    const topLevel = filteredNavItems.filter((item) => !('items' in item && item.items));
    const orders = topLevel.map((item) => (item as any).order ?? 50);
    const sorted = [...orders].sort((a, b) => a - b);
    expect(orders).toEqual(sorted);
    // order 50 was exclusively the removed backup item; no remaining item should use it.
    expect(orders).not.toContain(50);
  });

  // ── Regression: backups absent even when plugin registry returns items ───

  it('does not add a backups item even when plugin registry returns unrelated items', () => {
    vi.mocked(getPluginNavItems).mockReturnValueOnce([
      {
        title: 'Plugin Item',
        url: '/plugin-feature',
        icon: {} as any,
        resource: 'plugin',
        order: 60
      }
    ]);

    const { filteredNavItems } = renderSidebar();
    const urls = filteredNavItems.map((item) => item.url);
    expect(urls).not.toContain('/backups');
  });

  // ── Boundary: BROWSE_HIDDEN_URLS blocks /backups from plugins ────────────

  it('does not expose /backups even if a plugin tries to register it', () => {
    vi.mocked(getPluginNavItems).mockReturnValueOnce([
      {
        title: 'Backups Plugin',
        url: '/backups',
        icon: {} as any,
        resource: 'backup',
        order: 50
      }
    ]);

    const { filteredNavItems } = renderSidebar();
    const urls = filteredNavItems.map((item) => item.url);
    // /backups is in BROWSE_HIDDEN_URLS so buildNavItems filters it out.
    expect(urls).not.toContain('/backups');
  });

  // ── Basic shape assertions ────────────────────────────────────────────────

  it('returns filteredNavItems as an array', () => {
    const { filteredNavItems } = renderSidebar();
    expect(Array.isArray(filteredNavItems)).toBe(true);
  });

  it('returns at least five nav entries (four top-level + the Settings group)', () => {
    const { filteredNavItems } = renderSidebar();
    expect(filteredNavItems.length).toBeGreaterThanOrEqual(5);
  });
});
