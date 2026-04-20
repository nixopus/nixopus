import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';

// ---------------------------------------------------------------------------
// Module-level mocks – declared before any import of the module under test
// ---------------------------------------------------------------------------

// next/navigation
vi.mock('next/navigation', () => ({
  usePathname: vi.fn(() => '/apps')
}));

// lucide-react icons – lightweight stubs so the module doesn't need the full
// icon library at test time
vi.mock('lucide-react', () => ({
  Globe: 'Globe',
  HardDrive: 'HardDrive',
  KeyRound: 'KeyRound',
  Layers: 'Layers',
  Plug: 'Plug',
  Server: 'Server',
  Settings: 'Settings',
  Shield: 'Shield'
}));

// Redux hooks
const mockDispatch = vi.fn();
vi.mock('@/redux/hooks', () => ({
  useAppSelector: vi.fn((selector: (state: unknown) => unknown) =>
    selector({
      auth: { user: { id: 'user-1', email: 'test@example.com' } },
      user: { activeOrganization: { id: 'org-1', name: 'Test Org' } }
    })
  ),
  useAppDispatch: vi.fn(() => mockDispatch)
}));

// Auth organizations hook
vi.mock('@/packages/hooks/auth/use-better-auth-orgs', () => ({
  useUserOrganizations: vi.fn(() => ({
    data: [{ organization: { id: 'org-1', name: 'Test Org' } }],
    isLoading: false,
    refetch: vi.fn()
  }))
}));

// Navigation state hook
const mockSetActiveNav = vi.fn();
vi.mock('@/packages/hooks/shared/use_navigation_state', () => ({
  useNavigationState: vi.fn(() => ({
    activeNav: '/apps',
    setActiveNav: mockSetActiveNav
  }))
}));

// Translation hook – returns the key itself so assertions can use translation keys
vi.mock('@/packages/hooks/shared/use-translation', () => ({
  useTranslation: vi.fn(() => ({
    t: (key: string) => key
  }))
}));

// RBAC hook – grants access to all resources by default
const mockCanAccessResource = vi.fn(() => true);
vi.mock('@/packages/utils/rbac', () => ({
  useRBAC: vi.fn(() => ({
    canAccessResource: mockCanAccessResource
  }))
}));

// Redux slice actions
vi.mock('@/redux/features/users/userSlice', () => ({
  setActiveOrganization: vi.fn((org) => ({ type: 'user/setActiveOrganization', payload: org }))
}));
vi.mock('@/redux/features/users/authSlice', () => ({
  logout: vi.fn(() => ({ type: 'auth/logout' })),
  logoutUser: vi.fn(() => async () => {})
}));

// API services – each needs a util.resetApiState stub
const makeApiStub = () => ({ util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } });
vi.mock('@/redux/services/users/authApi', () => ({ authApi: makeApiStub() }));
vi.mock('@/redux/services/users/userApi', () => ({ userApi: makeApiStub() }));
vi.mock('@/redux/services/settings/notificationApi', () => ({ notificationApi: makeApiStub() }));
vi.mock('@/redux/services/settings/domainsApi', () => ({ domainsApi: makeApiStub() }));
vi.mock('@/redux/services/connector/githubConnectorApi', () => ({
  GithubConnectorApi: makeApiStub()
}));
vi.mock('@/redux/services/deploy/applicationsApi', () => ({ deployApi: makeApiStub() }));
vi.mock('@/redux/services/file-manager/fileManagersApi', () => ({
  fileManagersApi: makeApiStub()
}));
vi.mock('@/redux/services/audit', () => ({ auditApi: makeApiStub() }));
vi.mock('@/redux/services/feature-flags/featureFlagsApi', () => ({
  FeatureFlagsApi: makeApiStub()
}));
vi.mock('@/redux/services/api-keys/apiKeysApi', () => ({ apiKeysApi: makeApiStub() }));
vi.mock('@/redux/services/machine/machineBackupApi', () => ({
  machineBackupApi: makeApiStub()
}));
vi.mock('@/redux/services/domains/customDomainsApi', () => ({
  customDomainsApi: makeApiStub()
}));

// Plugin registry – return no plugin items by default
const mockGetPluginNavItems = vi.fn(() => []);
vi.mock('@/plugins/registry', () => ({
  getPluginNavItems: mockGetPluginNavItems
}));

// ---------------------------------------------------------------------------
// Import the module under test AFTER all vi.mock() calls
// ---------------------------------------------------------------------------
import { useAppSidebar } from '../use-app-sidebar';
import { usePathname } from 'next/navigation';
import { useAppSelector } from '@/redux/hooks';

// ---------------------------------------------------------------------------
// Helper types
// ---------------------------------------------------------------------------
interface NavItem {
  title: string;
  url: string;
  resource?: string;
  icon?: unknown;
  order?: number;
  items?: Array<{ title: string; url: string; resource?: string; section?: string }>;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useAppSidebar – coreNavItems after backups removal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetPluginNavItems.mockReturnValue([]);
    // Restore default selector behaviour
    (useAppSelector as ReturnType<typeof vi.fn>).mockImplementation(
      (selector: (state: unknown) => unknown) =>
        selector({
          auth: { user: { id: 'user-1', email: 'test@example.com' } },
          user: { activeOrganization: { id: 'org-1', name: 'Test Org' } }
        })
    );
    (usePathname as ReturnType<typeof vi.fn>).mockReturnValue('/apps');
  });

  // -------------------------------------------------------------------------
  // Core: backups item must NOT appear
  // -------------------------------------------------------------------------

  it('does not include a nav item with url "/backups"', () => {
    const { result } = renderHook(() => useAppSidebar());
    const backupsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/backups'
    );
    expect(backupsItem).toBeUndefined();
  });

  it('does not include a nav item with title "navigation.backups"', () => {
    const { result } = renderHook(() => useAppSidebar());
    const backupsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.title === 'navigation.backups'
    );
    expect(backupsItem).toBeUndefined();
  });

  it('does not include a nav item with resource "backup"', () => {
    const { result } = renderHook(() => useAppSidebar());
    const backupsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.resource === 'backup'
    );
    expect(backupsItem).toBeUndefined();
  });

  // -------------------------------------------------------------------------
  // Regression: remaining core items must still be present
  // -------------------------------------------------------------------------

  it('includes the self-host (apps) nav item', () => {
    const { result } = renderHook(() => useAppSidebar());
    const appsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/apps'
    );
    expect(appsItem).toBeDefined();
    expect(appsItem?.title).toBe('navigation.selfHost');
  });

  it('includes the machines nav item', () => {
    const { result } = renderHook(() => useAppSidebar());
    const machinesItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/machines'
    );
    expect(machinesItem).toBeDefined();
    expect(machinesItem?.title).toBe('navigation.machines');
  });

  it('includes the integrations nav item', () => {
    const { result } = renderHook(() => useAppSidebar());
    const integrationsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/integrations'
    );
    expect(integrationsItem).toBeDefined();
    expect(integrationsItem?.title).toBe('navigation.integrations');
  });

  it('includes the domains nav item', () => {
    const { result } = renderHook(() => useAppSidebar());
    const domainsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/domains'
    );
    expect(domainsItem).toBeDefined();
    expect(domainsItem?.title).toBe('navigation.domains');
  });

  it('includes the Settings group containing the api-keys sub-item', () => {
    const { result } = renderHook(() => useAppSidebar());
    const settingsGroup = result.current.filteredNavItems.find(
      (item: NavItem) => item.title === 'Settings'
    );
    expect(settingsGroup).toBeDefined();
    const apiKeysSubItem = settingsGroup?.items?.find(
      (sub: { url: string }) => sub.url === '/api-keys'
    );
    expect(apiKeysSubItem).toBeDefined();
  });

  // -------------------------------------------------------------------------
  // Total count: verify removing backups reduced top-level items by exactly 1
  // compared to what the remaining core items define (no plugins)
  // Top-level items in coreNavItems without group: selfHost, machines, integrations, domains
  // Settings group is built from grouped items (api-keys, security, General, etc.)
  // Expected top-level visible items: 4 top-level + 1 Settings group = 5
  // -------------------------------------------------------------------------

  it('returns the correct total number of top-level nav items without backups', () => {
    const { result } = renderHook(() => useAppSidebar());
    // Top-level (no group): apps, machines, integrations, domains = 4
    // Settings group: 1
    // Total = 5
    expect(result.current.filteredNavItems).toHaveLength(5);
  });

  // -------------------------------------------------------------------------
  // Plugin item with url '/backups' is filtered out via BROWSE_HIDDEN_URLS
  // -------------------------------------------------------------------------

  it('filters out a plugin-provided nav item with url "/backups" via BROWSE_HIDDEN_URLS', () => {
    mockGetPluginNavItems.mockReturnValue([
      { title: 'Plugin Backups', url: '/backups', resource: 'backup', order: 50 }
    ]);

    const { result } = renderHook(() => useAppSidebar());
    const backupsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/backups'
    );
    expect(backupsItem).toBeUndefined();
  });

  it('does not filter out plugin-provided items with urls other than BROWSE_HIDDEN_URLS', () => {
    mockGetPluginNavItems.mockReturnValue([
      { title: 'Custom Plugin', url: '/custom-plugin', resource: 'settings', order: 99 }
    ]);

    const { result } = renderHook(() => useAppSidebar());
    const customItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/custom-plugin'
    );
    expect(customItem).toBeDefined();
  });

  // -------------------------------------------------------------------------
  // Permission-based filtering still works after the change
  // -------------------------------------------------------------------------

  it('hides nav items from filteredNavItems when the user lacks permission', () => {
    // Deny access to 'deploy' resource so the apps item is hidden
    mockCanAccessResource.mockReturnValue(false);

    const { result } = renderHook(() => useAppSidebar());
    const appsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/apps'
    );
    expect(appsItem).toBeUndefined();
  });

  it('returns no filteredNavItems when user is not authenticated', () => {
    (useAppSelector as ReturnType<typeof vi.fn>).mockImplementation(
      (selector: (state: unknown) => unknown) =>
        selector({
          auth: { user: null },
          user: { activeOrganization: null }
        })
    );

    const { result } = renderHook(() => useAppSidebar());
    // hasAnyPermission returns false when user or activeOrg is null,
    // so all items requiring permission are filtered out.
    // The Settings group items also require permission, so the group disappears too.
    expect(result.current.filteredNavItems).toHaveLength(0);
  });

  // -------------------------------------------------------------------------
  // Nav item ordering: backups had order 50, verify surrounding items order
  // is preserved correctly (integrations=41, domains=40, apiKeys=92)
  // -------------------------------------------------------------------------

  it('nav items are sorted in the correct order (domains before integrations)', () => {
    const { result } = renderHook(() => useAppSidebar());
    const items = result.current.filteredNavItems;
    const domainsIndex = items.findIndex((item: NavItem) => item.url === '/domains');
    const integrationsIndex = items.findIndex((item: NavItem) => item.url === '/integrations');

    // domains order=40, integrations order=41, so domains comes first
    expect(domainsIndex).toBeLessThan(integrationsIndex);
  });

  it('machines nav item (order=15) appears before domains (order=40)', () => {
    const { result } = renderHook(() => useAppSidebar());
    const items = result.current.filteredNavItems;
    const machinesIndex = items.findIndex((item: NavItem) => item.url === '/machines');
    const domainsIndex = items.findIndex((item: NavItem) => item.url === '/domains');

    expect(machinesIndex).toBeLessThan(domainsIndex);
  });

  it('Settings group appears after top-level items', () => {
    const { result } = renderHook(() => useAppSidebar());
    const items = result.current.filteredNavItems;
    const settingsIndex = items.findIndex((item: NavItem) => item.title === 'Settings');
    const appsIndex = items.findIndex((item: NavItem) => item.url === '/apps');

    expect(settingsIndex).toBeGreaterThan(appsIndex);
  });

  // -------------------------------------------------------------------------
  // Boundary / regression: '/chats' is also in BROWSE_HIDDEN_URLS
  // -------------------------------------------------------------------------

  it('filters out a plugin-provided nav item with url "/chats" via BROWSE_HIDDEN_URLS', () => {
    mockGetPluginNavItems.mockReturnValue([
      { title: 'Chats', url: '/chats', resource: 'settings', order: 60 }
    ]);

    const { result } = renderHook(() => useAppSidebar());
    const chatsItem = result.current.filteredNavItems.find(
      (item: NavItem) => item.url === '/chats'
    );
    expect(chatsItem).toBeUndefined();
  });
});