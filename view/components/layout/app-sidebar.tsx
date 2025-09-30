'use client';

import * as React from 'react';
import {
  Folder,
  Home,
  Package,
  SettingsIcon,
  Container,
  LucideProps
} from 'lucide-react';
import { NavMain } from '@/components/layout/nav-main';
import { NavUser } from '@/components/layout/nav-user';
import { TeamSwitcher } from '@/components/ui/team-switcher';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail
} from '@/components/ui/sidebar';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import { useGetUserOrganizationsQuery } from '@/redux/services/users/userApi';
import { useNavigationState } from '@/hooks/use_navigation_state';
import { setActiveOrganization } from '@/redux/features/users/userSlice';
import { useTranslation } from '@/hooks/use-translation';
import { useRBAC } from '@/lib/rbac';

// Add proper TypeScript interfaces

// 1. Defined specific types for User and Organization
interface User {
  id: string;
  name?: string;
  email?: string;
  // Add other user properties as needed
}

interface Organization {
  id: string;
  name?: string;
  // Add other organization properties as needed
}

// 2. Updated AppState to use the specific types instead of 'any'
interface AppState {
  auth: {
    user: User | null;
  };
  user: {
    organizations: { organization: Organization }[];
    activeOrganization: Organization | null;
  };
}

// 3. Created a strict type for resource strings for type safety
type Resource =
  | 'dashboard'
  | 'deploy'
  | 'container'
  | 'file-manager'
  | 'settings'
  | 'notification'
  | 'organization'
  | 'domain';

interface NavSubItem {
  title: string;
  url: string;
  resource: Resource;
}

// 4. Updated NavItem to use the Resource type and a more specific icon type
interface NavItem {
  title: string;
  url: string;
  icon: React.ComponentType<LucideProps>;
  resource: Resource;
  items?: NavSubItem[];
}

const data: { navMain: NavItem[] } = {
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

// 5. Removed the unused 'toggleAddTeamModal' prop
export function AppSidebar(props: React.ComponentProps<typeof Sidebar>) {
  const { t } = useTranslation();
  const user = useAppSelector((state: AppState) => state.auth.user);
  const { refetch } = useGetUserOrganizationsQuery();
  const organizations = useAppSelector((state: AppState) => state.user.organizations);
  const { activeNav, setActiveNav } = useNavigationState();
  const activeOrg = useAppSelector((state: AppState) => state.user.activeOrganization);
  const dispatch = useAppDispatch();
  const { canAccessResource } = useRBAC();

  const hasAnyPermission = React.useMemo(() => {
    const allowedResources: Resource[] = ['dashboard', 'settings'];

    return (resource: Resource) => {
      if (!user || !activeOrg) return false;

      if (allowedResources.includes(resource)) {
        return true;
      }

      return (
        canAccessResource(resource, 'read') ||
        canAccessResource(resource, 'create') ||
        canAccessResource(resource, 'update') ||
        canAccessResource(resource, 'delete')
      );
    };
  }, [user, activeOrg, canAccessResource]);

  // 6. Corrected and optimized the logic for filtering and mapping navigation items
  const filteredNavItems = React.useMemo(() => {
    return data.navMain.reduce<NavItem[]>((acc, item) => {
      let visibleSubItems: NavSubItem[] | undefined;

      if (item.items) {
        visibleSubItems = item.items
          .filter(subItem => hasAnyPermission(subItem.resource))
          .map(subItem => ({
            ...subItem,
            title: t(subItem.title)
          }));
      }

      // An item is visible if it has direct permission OR it has visible sub-items
      if (hasAnyPermission(item.resource) || (visibleSubItems && visibleSubItems.length > 0)) {
        acc.push({
          ...item,
          title: t(item.title),
          items: visibleSubItems
        });
      }

      return acc;
    }, []);
  }, [hasAnyPermission, t]);

  React.useEffect(() => {
    if (organizations && organizations.length > 0 && !activeOrg) {
      dispatch(setActiveOrganization(organizations[0].organization));
    }
  }, [organizations, activeOrg, dispatch]);

  React.useEffect(() => {
    if (activeOrg?.id) {
      refetch();
    }
  }, [activeOrg?.id, refetch]);

  const handleSponsorClick = React.useCallback((): void => {
    window.open('https://github.com/sponsors/raghavyuva', '_blank');
  }, []);

  const handleHelpClick = React.useCallback((): void => {
    window.open('https://docs.nixopus.com', '_blank');
  }, []);

  const handleReportIssueClick = React.useCallback((): void => {
    const userAgent = typeof navigator !== 'undefined' ? navigator.userAgent : '';
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
      : userAgent.includes('iPhone') || userAgent.includes('iPad')
      ? 'iOS'
      : 'Unknown';

    const screenResolution =
      typeof screen !== 'undefined' ? `${screen.width}x${screen.height}` : 'Unknown';
    const language = typeof navigator !== 'undefined' ? navigator.language : 'Unknown';
    const timezone =
      typeof Intl !== 'undefined' && Intl.DateTimeFormat
        ? Intl.DateTimeFormat().resolvedOptions().timeZone
        : 'Unknown';

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
- Browser: ${browser}
- Operating System: ${os}
- Screen Resolution: ${screenResolution}
- Language: ${language}
- Timezone: ${timezone}
- User Agent: ${userAgent}

Add any other context about the problem here.`;

    const encodedBody = encodeURIComponent(issueBody);
    const url = `https://github.com/raghavyuva/nixopus/issues/new?template=bug_report.md&body=${encodedBody}`;
    window.open(url, '_blank');
  }, []);

  if (!user || !activeOrg) {
    return null;
  }

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher refetch={refetch} />
      </SidebarHeader>
      <SidebarContent>
        {/* 7. Simplified NavMain props, as filtering is now handled correctly in useMemo */}
        <NavMain
          items={filteredNavItems.map((item) => ({
            ...item,
            isActive: item.url === activeNav
          }))}
          onItemClick={(url: string) => setActiveNav(url)}
        />
      </SidebarContent>
      <SidebarFooter>
        <div className="w-full">
          <NavUser user={user} />
        </div>

        <div className="w-full px-3 py-3">
          <div className="flex flex-col gap-2">
            <button
              type="button"
              onClick={handleSponsorClick}
              className="w-full py-2 rounded-lg border"
            >
              {t('user.menu.sponsor') || 'Sponsor'}
            </button>

            <button
              type="button"
              onClick={handleHelpClick}
              className="w-full py-2 rounded-lg border"
            >
              {t('user.menu.help') || 'Help'}
            </button>

            <button
              type="button"
              onClick={handleReportIssueClick}
              className="w-full py-2 rounded-lg border"
            >
              {t('user.menu.reportIssue') || 'Report Issue'}
            </button>
          </div>
        </div>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}