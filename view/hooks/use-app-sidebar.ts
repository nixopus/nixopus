import { useRouter } from 'next/navigation';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import { useGetUserOrganizationsQuery } from '@/redux/services/users/userApi';
import { useNavigationState } from '@/hooks/use_navigation_state';
import { setActiveOrganization } from '@/redux/features/users/userSlice';
import { useTranslation } from '@/hooks/use-translation';
import { useRBAC } from '@/lib/rbac';
import { logout, logoutUser } from '@/redux/features/users/authSlice';
import { authApi } from '@/redux/services/users/authApi';
import { userApi } from '@/redux/services/users/userApi';
import { notificationApi } from '@/redux/services/settings/notificationApi';
import { domainsApi } from '@/redux/services/settings/domainsApi';
import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { deployApi } from '@/redux/services/deploy/applicationsApi';
import { fileManagersApi } from '@/redux/services/file-manager/fileManagersApi';
import { auditApi } from '@/redux/services/audit';
import { useState, useMemo, useEffect } from 'react';
import { Folder, Home, Package, SettingsIcon, Container, Puzzle } from 'lucide-react';

const data = {
  navMain: [
    {
      title: 'navigation.dashboard',
      url: '/dashboard',
      icon: Home,
      resource: 'dashboard'
    },
    {
      title: 'navigation.extensions',
      url: '/extensions',
      icon: Puzzle,
      resource: 'extensions'
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
      title: 'navigation.fileManager',
      url: '/file-manager',
      icon: Folder,
      resource: 'file-manager'
    },
    {
      title: 'navigation.settings',
      url: '/settings/general',
      icon: SettingsIcon,
      resource: 'settings',
      items: [
        {
          title: 'navigation.general',
          url: '/settings/general',
          resource: 'settings'
        },
        {
          title: 'navigation.notifications',
          url: '/settings/notifications',
          resource: 'notification'
        },
        {
          title: 'navigation.team',
          url: '/settings/teams',
          resource: 'organization'
        },
        {
          title: 'navigation.domains',
          url: '/settings/domains',
          resource: 'domain'
        }
      ]
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
  const [showLogoutDialog, setShowLogoutDialog] = useState(false);

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
      auditApi
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

          if (item.items) {
            const filteredSubItems = item.items.filter(
              (subItem) => subItem.resource && hasAnyPermission(subItem.resource)
            );
            return filteredSubItems.length > 0;
          }

          return hasAnyPermission(item.resource);
        })
        .map((item) => ({
          ...item,
          title: t(item.title as any),
          items: item.items?.map((subItem) => ({
            ...subItem,
            title: t(subItem.title as any)
          }))
        })),
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
