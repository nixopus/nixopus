'use client';

import * as React from 'react';
import { AlertCircle, HelpCircle, Heart, LogOut } from 'lucide-react';
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
import { LogoutDialog } from '@/components/ui/logout-dialog';

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
    showLogoutDialog,
    filteredNavItems,
    handleSponsor,
    handleReportIssue,
    handleHelp,
    handleLogoutClick,
    handleLogoutConfirm,
    handleLogoutCancel
  } = useAppSidebar();

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
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={handleSponsor} className="cursor-pointer">
              <Heart className="text-red-500" />
              <span>{t('user.menu.sponsor')}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={handleHelp} className="cursor-pointer">
              <HelpCircle />
              <span>{t('user.menu.help')}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={handleReportIssue} className="cursor-pointer">
              <AlertCircle />
              <span>{t('user.menu.reportIssue')}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={handleLogoutClick} className="cursor-pointer">
              <LogOut />
              <span>{t('user.menu.logout')}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
      <LogoutDialog
        open={showLogoutDialog}
        onConfirm={handleLogoutConfirm}
        onCancel={handleLogoutCancel}
      />
    </Sidebar>
  );
}
