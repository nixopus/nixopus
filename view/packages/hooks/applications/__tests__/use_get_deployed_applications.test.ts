import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';

// ---------------------------------------------------------------------------
// Mocks – must be declared before the module under test is imported so that
// vitest's hoisting mechanism replaces the real modules.
// ---------------------------------------------------------------------------

// Simulates the REAL backend: server-side pagination, `limit` apps per page,
// `total_count` is the full count across all pages. This mirrors
// DeployService.GetApplicationsByOrganization / the `/deploy/applications`
// endpoint's actual page/limit/total_count contract.
const ALL_APPS = Array.from({ length: 25 }, (_, i) => ({
  id: `app-${i + 1}`,
  name: `app-${String(i + 1).padStart(2, '0')}`,
  domains: [],
  environment: 'production',
  updated_at: '2026-09-30T00:00:00Z',
  build_pack: 'dockerfile',
  port: 3000,
  labels: []
}));

vi.mock('@/redux/services/deploy/applicationsApi', () => ({
  useGetApplicationsQuery: ({ page, limit }: { page: number; limit: number }) => {
    const start = (page - 1) * limit;
    const pageApps = ALL_APPS.slice(start, start + limit);
    return {
      data: { applications: pageApps, total_count: ALL_APPS.length },
      refetch: vi.fn(),
      isLoading: false,
      isFetching: false
    };
  },
  deployApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));

vi.mock('@/redux/services/connector/githubConnectorApi', () => ({
  useGetAllGithubConnectorQuery: () => ({ data: [], refetch: vi.fn(), isLoading: false }),
  useUpdateGithubConnectorMutation: () => [vi.fn(), { isLoading: false }],
  useCreateGithubConnectorMutation: () => [vi.fn()],
  GithubConnectorApi: { util: { resetApiState: vi.fn(() => ({ type: 'reset' })) } }
}));

interface MockAppState {
  user: { activeOrganization: { id: string } };
  githubConnector: { activeConnectorId: string | null };
}

vi.mock('@/redux/hooks', () => ({
  useAppSelector: <T>(selector: (s: MockAppState) => T): T =>
    selector({
      user: { activeOrganization: { id: 'org-1' } },
      githubConnector: { activeConnectorId: null }
    }),
  useAppDispatch: () => vi.fn()
}));

vi.mock('@/redux/features/github-connector/githubConnectorSlice', () => ({
  setActiveConnectorId: (id: string | null) => ({
    type: 'githubConnector/setActiveConnectorId',
    payload: id
  })
}));

vi.mock('@/redux/conf', () => ({
  getSelfHosted: () => Promise.resolve(true)
}));

vi.mock('@/packages/contexts/machine-context', () => ({
  useMachineId: () => null
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => '/apps'
}));

// use-searchable and use_label_filter are left UNMOCKED on purpose: they are
// plain pass-through array utilities when there's no active search term or
// selected labels, so running them for real exercises the actual fix path
// instead of hiding it behind a mock.

// ---------------------------------------------------------------------------
// The module under test
// ---------------------------------------------------------------------------
import useGetDeployedApplications from '../use_get_deployed_applications';

// Regression test for the double-pagination bug: the server already
// paginates (page/limit -> an applications array of at most `limit` items +
// total_count across the whole org). The hook used to re-slice that
// already-paginated array a second time with a different, hardcoded page
// size (9), as if it were the full dataset — silently dropping almost every
// app past page 1. These tests assert the CORRECT behavior: the client
// trusts and renders the server's page as-is.
describe('useGetDeployedApplications – server pagination is trusted as-is (25 apps, server limit=10)', () => {
  it('page 1: server returns 10 apps, and all 10 are rendered', () => {
    const { result } = renderHook(() => useGetDeployedApplications());

    expect(result.current.currentPage).toBe(1);
    expect(result.current.applications.length).toBe(10);
    expect(result.current.applications.map((a) => a.id)).toContain('app-10');
  });

  it('page 2: server returns apps 11-20, and all 10 are rendered', () => {
    const { result } = renderHook(() => useGetDeployedApplications());

    act(() => {
      result.current.handlePageChange(2);
    });

    expect(result.current.currentPage).toBe(2);
    expect(result.current.applications.length).toBe(10);
    expect(result.current.applications.map((a) => a.id)).toEqual([
      'app-11',
      'app-12',
      'app-13',
      'app-14',
      'app-15',
      'app-16',
      'app-17',
      'app-18',
      'app-19',
      'app-20'
    ]);
  });

  it('page 3: server returns the last 5 apps (21-25), and all 5 are rendered', () => {
    const { result } = renderHook(() => useGetDeployedApplications());

    act(() => {
      result.current.handlePageChange(3);
    });

    expect(result.current.currentPage).toBe(3);
    expect(result.current.applications.length).toBe(5);
    expect(result.current.showApplications).toBe(true);
  });

  it('totalPages matches the server page size (10), consistent with what is actually rendered per page', () => {
    const { result } = renderHook(() => useGetDeployedApplications());
    // Math.ceil(25 / 10) = 3 server-side pages, and each page now renders
    // its full ≤10-item content (see tests above) — no more mismatch between
    // the pagination controls and what's actually shown.
    expect(result.current.totalPages).toBe(3);
  });
});
