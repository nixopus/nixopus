'use client';

import * as React from 'react';
import { Home, Package, SettingsIcon } from 'lucide-react';

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
      url: '/',
      icon: Home,
      isActive: true
    },
    {
      title: 'Self Host',
      url: '#',
      icon: Package
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
        },
        {
          title: 'Billing',
          url: '#'
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
