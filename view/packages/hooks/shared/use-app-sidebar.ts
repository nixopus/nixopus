import { useRouter, usePathname } from 'next/navigation';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import { useGetUserOrganizationsQuery } from '@/redux/services/users/userApi';
import { useNavigationState } from '@/packages/hooks/shared/use_navigation_state';
import { setActiveOrganization } from '@/redux/features/users/userSlice';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRBAC } from '@/packages/utils/rbac';
import { logout, logoutUser } from '@/redux/features/users/authSlice';
import { authApi } from '@/redux/services/users/authApi';
import { userApi } from '@/redux/services/users/userApi';
import { notificationApi } from '@/redux/services/settings/notificationApi';
import { domainsApi } from '@/redux/services/settings/domainsApi';
import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { deployApi } from '@/redux/services/deploy/applicationsApi';
import { fileManagersApi } from '@/redux/services/file-manager/fileManagersApi';
import { auditApi } from '@/redux/services/audit';
import { FeatureFlagsApi } from '@/redux/services/feature-flags/featureFlagsApi';
import { useState, useMemo, useEffect } from 'react';
import { Folder, Home, Package, Container, Puzzle } from 'lucide-react';
import { useSettingsModal } from '@/packages/hooks/shared/use-settings-modal';

const data = {
  navMain: [
    {
      title: 'navigation.dashboard',
      url: '/dashboard',
      icon: Home,
      resource: 'dashboard'
    },
    {
      title: 'navigation.selfHost',
      url: '/self-host',
      icon: Package,
      resource: 'deploy'
    },
    {
      title: 'navigation.containers',
      url: '/containers',
      icon: Container,
      resource: 'container'
    },
    {
      title: 'navigation.extensions',
      url: '/extensions',
      icon: Puzzle,
      resource: 'extensions'
    }
  ]
};

export function useAppSidebar() {
  const { t } = useTranslation();
  const user = useAppSelector((state) => state.auth.user);
  const { isLoading, refetch } = useGetUserOrganizationsQuery();
  const organizations = useAppSelector((state) => state.user.organizations);
  const { activeNav, setActiveNav } = useNavigationState();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const dispatch = useAppDispatch();
  const { canAccessResource } = useRBAC();
  const router = useRouter();
  const pathname = usePathname();
  const [showLogoutDialog, setShowLogoutDialog] = useState(false);
  const { closeSettings } = useSettingsModal();

  const hasAnyPermission = useMemo(() => {
    const allowedResources = ['dashboard', 'settings', 'extensions'];

    return (resource: string) => {
      if (!user || !activeOrg) return false;

      if (allowedResources.includes(resource)) {
        return true;
      }

      return (
        canAccessResource(resource as any, 'read') ||
        canAccessResource(resource as any, 'create') ||
        canAccessResource(resource as any, 'update') ||
        canAccessResource(resource as any, 'delete')
      );
    };
  }, [user, activeOrg, canAccessResource]);

  const clearLocalStorage = () => {
    const keys = [
      'COLLAPSIBLE_STATE_KEY',
      'LAST_ACTIVE_NAV_KEY',
      'SIDEBAR_STORAGE_KEY',
      'terminalOpen',
      'persist:root',
      'active_organization'
    ];
    keys.forEach((key) => localStorage.removeItem(key));
  };

  const resetApiStates = () => {
    const apis = [
      authApi,
      userApi,
      notificationApi,
      domainsApi,
      GithubConnectorApi,
      deployApi,
      fileManagersApi,
      auditApi,
      FeatureFlagsApi
    ];
    apis.forEach((api) => dispatch(api.util.resetApiState()));
  };

  const handleSponsor = () => {
    window.open('https://github.com/sponsors/raghavyuva', '_blank');
  };

  const getClientInfo = () => {
    const userAgent = navigator.userAgent;
    const browser = userAgent.includes('Chrome')
      ? 'Chrome'
      : userAgent.includes('Firefox')
        ? 'Firefox'
        : userAgent.includes('Safari')
          ? 'Safari'
          : userAgent.includes('Edge')
            ? 'Edge'
            : 'Unknown';

    const os = userAgent.includes('Windows')
      ? 'Windows'
      : userAgent.includes('Mac')
        ? 'macOS'
        : userAgent.includes('Linux')
          ? 'Linux'
          : userAgent.includes('Android')
            ? 'Android'
            : userAgent.includes('iOS')
              ? 'iOS'
              : 'Unknown';

    return {
      browser,
      os,
      userAgent,
      screenResolution: `${screen.width}x${screen.height}`,
      language: navigator.language,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
    };
  };

  const handleReportIssue = () => {
    const clientInfo = getClientInfo();

    const issueBody = `**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Additional context**
- Browser: ${clientInfo.browser}
- Operating System: ${clientInfo.os}
- Screen Resolution: ${clientInfo.screenResolution}
- Language: ${clientInfo.language}
- Timezone: ${clientInfo.timezone}
- User Agent: ${clientInfo.userAgent}

Add any other context about the problem here.`;

    const encodedBody = encodeURIComponent(issueBody);
    const url = `https://github.com/raghavyuva/nixopus/issues/new?template=bug_report.md&body=${encodedBody}`;
    window.open(url, '_blank');
  };

  const handleHelp = () => {
    window.open('https://docs.nixopus.com', '_blank');
  };

  const handleLogoutClick = () => {
    setShowLogoutDialog(true);
  };

  const handleLogoutConfirm = async () => {
    setShowLogoutDialog(false);
    closeSettings();

    try {
      clearLocalStorage();
      resetApiStates();
      dispatch({ type: 'RESET_STATE' });
      await dispatch(logoutUser() as any);
      router.push('/auth');
    } catch (error) {
      console.error('Logout failed:', error);
      clearLocalStorage();
      resetApiStates();
      dispatch({ type: 'RESET_STATE' });
      dispatch(logout());
      router.push('/auth');
    }
  };

  const handleLogoutCancel = () => {
    setShowLogoutDialog(false);
  };

  const filteredNavItems = useMemo(
    () =>
      data.navMain
        .filter((item) => {
          if (!item.resource) return false;

          if ('items' in item && item.items && Array.isArray(item.items)) {
            const filteredSubItems = item.items.filter(
              (subItem: { resource?: string }) =>
                subItem.resource && hasAnyPermission(subItem.resource)
            );
            return filteredSubItems.length > 0;
          }

          return hasAnyPermission(item.resource);
        })
        .map((item) => {
          const baseItem = {
            ...item,
            title: t(item.title as any)
          };

          if ('items' in item && item.items && Array.isArray(item.items)) {
            return {
              ...baseItem,
              items: item.items.map(
                (subItem: { title: string; url: string; resource?: string }) => ({
                  ...subItem,
                  title: t(subItem.title as any)
                })
              )
            };
          }

          return baseItem;
        }),
    [data.navMain, hasAnyPermission, t]
  );

  useEffect(() => {
    if (organizations && organizations.length > 0 && !activeOrg) {
      dispatch(setActiveOrganization(organizations[0].organization));
    }
  }, [organizations, activeOrg, dispatch]);

  useEffect(() => {
    if (activeOrg?.id) {
      refetch();
    }
  }, [activeOrg?.id, refetch]);

  // Sync activeNav with current pathname to prevent multiple active menu items
  useEffect(() => {
    if (pathname) {
      // Find the matching navigation item URL for the current pathname
      const matchingNavItem = data.navMain.find(
        (item) => pathname === item.url || pathname.startsWith(item.url + '/')
      );

      // Only update activeNav if we found a match and it's different from current activeNav
      if (matchingNavItem && matchingNavItem.url !== activeNav) {
        setActiveNav(matchingNavItem.url);
      }
    }
  }, [pathname, activeNav, setActiveNav]);

  return {
    user,
    isLoading,
    refetch,
    organizations,
    activeNav,
    setActiveNav,
    activeOrg,
    dispatch,
    hasAnyPermission,
    t,
    showLogoutDialog,
    filteredNavItems,
    handleSponsor,
    handleReportIssue,
    handleHelp,
    handleLogoutClick,
    handleLogoutConfirm,
    handleLogoutCancel
  };
}
