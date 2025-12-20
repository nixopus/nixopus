'use client';

import * as React from 'react';
import { Settings } from 'lucide-react';
import { NavMain } from '@/components/layout/nav-main';
import { TeamSwitcher } from '@/components/ui/team-switcher';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem
} from '@/components/ui/sidebar';
import { useAppSidebar } from '@/hooks/use-app-sidebar';
import { useSettingsModal } from '@/hooks/use-settings-modal';

export function AppSidebar({
  toggleAddTeamModal,
  addTeamModalOpen,
  ...props
}: React.ComponentProps<typeof Sidebar> & {
  toggleAddTeamModal?: () => void;
  addTeamModalOpen?: boolean;
}) {
  const {
    user,
    refetch,
    activeNav,
    setActiveNav,
    activeOrg,
    hasAnyPermission,
    t,
    filteredNavItems
  } = useAppSidebar();
  const { openSettings } = useSettingsModal();

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
            items:
              'items' in item && item.items
                ? item.items.filter(
                    (subItem) => subItem.resource && hasAnyPermission(subItem.resource)
                  )
                : undefined
          }))}
          onItemClick={(url) => setActiveNav(url)}
        />
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={() => openSettings()} className="cursor-pointer">
              <Settings />
              <span>{t('settings.title')}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
