'use client';

import * as React from 'react';
import { Folder, Home, Package, SettingsIcon } from 'lucide-react';
import { NavMain } from '@/components/nav-main';
import { NavUser } from '@/components/nav-user';
import { TeamSwitcher } from '@/components/team-switcher';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail
} from '@/components/ui/sidebar';
import { useAppSelector } from '@/redux/hooks';
import { useGetUserOrganizationsQuery } from '@/redux/services/users/userApi';

const data = {
  navMain: [
    {
      title: 'Dashboard',
      url: '/dashboard',
      icon: Home,
      isActive: true
    },
    {
      title: 'Self Host',
      url: '/self-host',
      icon: Package
    },
    {
      title: 'File Manager',
      url: '/file-manager',
      icon: Folder
    },
    {
      title: 'Settings',
      url: '/settings/general',
      icon: SettingsIcon,
      items: [
        {
          title: 'General',
          url: '/settings/general'
        },
        {
          title: 'Notifications',
          url: '/settings/notifications'
        },
        {
          title: 'Team',
          url: '/settings/teams'
        },
        {
          title: 'Domains',
          url: '/settings/domains'
        }
      ]
    }
  ]
};

export function AppSidebar({
  toggleAddTeamModal,
  ...props
}: React.ComponentProps<typeof Sidebar> & { toggleAddTeamModal?: () => void }) {
  const user = useAppSelector((state) => state.auth.user);
  const { isLoading } = useGetUserOrganizationsQuery();
  const organizations = useAppSelector((state) => state.user.organizations);

  React.useEffect(() => {
    if (user && user.type !== 'admin') {
      delete data.navMain[2];
    }
  }, [user]);

  if (!user) {
    return null;
  }

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher
          teams={isLoading ? [] : organizations}
          toggleAddTeamModal={toggleAddTeamModal}
        />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
