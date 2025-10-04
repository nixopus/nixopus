'use client';

import * as React from 'react';
import { Folder, Home, Package, SettingsIcon, Container } from 'lucide-react';
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

export function AppSidebar({
  toggleAddTeamModal,
  addTeamModalOpen,
  ...props
}: React.ComponentProps<typeof Sidebar> & { 
  toggleAddTeamModal?: () => void;
  addTeamModalOpen?: boolean;
}) {
  const { t } = useTranslation();
  const user = useAppSelector((state) => state.auth.user);
  const { isLoading, refetch } = useGetUserOrganizationsQuery();
  const organizations = useAppSelector((state) => state.user.organizations);
  const { activeNav, setActiveNav } = useNavigationState();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const dispatch = useAppDispatch();
  const { canAccessResource } = useRBAC();

  const hasAnyPermission = React.useMemo(() => {
    const allowedResources = ['dashboard', 'settings'];

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

  const filteredNavItems = React.useMemo(
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
          title: t(item.title),
          items: item.items?.map((subItem) => ({
            ...subItem,
            title: t(subItem.title)
          }))
        })),
    [data.navMain, hasAnyPermission, t]
  );

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

  if (!user || !activeOrg) {
    return null;
  }

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher 
          refetch={refetch} 
          toggleAddTeamModal={toggleAddTeamModal}
          addTeamModalOpen={addTeamModalOpen}
        />
      </SidebarHeader>
      <SidebarContent>
        <NavMain
          items={filteredNavItems.map((item) => ({
            ...item,
            isActive: item.url === activeNav,
            items: item.items?.filter(
              (subItem) => subItem.resource && hasAnyPermission(subItem.resource)
            )
          }))}
          onItemClick={(url) => setActiveNav(url)}
        />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
