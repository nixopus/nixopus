'use client';

import { AlertCircle, ChevronsUpDown, HelpCircle, Heart, LogOut } from 'lucide-react';
import { useRouter } from 'next/navigation';

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar
} from '@/components/ui/sidebar';
import { User } from '@/redux/types/user';
import { useAppDispatch, useAppSelector } from '@/redux/hooks';
import { logout, logoutUser } from '@/redux/features/users/authSlice';
import { authApi } from '@/redux/services/users/authApi';
import { userApi } from '@/redux/services/users/userApi';
import { notificationApi } from '@/redux/services/settings/notificationApi';
import { domainsApi } from '@/redux/services/settings/domainsApi';
import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { deployApi } from '@/redux/services/deploy/applicationsApi';
import { fileManagersApi } from '@/redux/services/file-manager/fileManagersApi';
import { auditApi } from '@/redux/services/audit';
import { useTranslation } from '@/hooks/use-translation';

export function NavUser({ user }: { user: User }) {
  const { isMobile } = useSidebar();
  const dispatch = useAppDispatch();
  const router = useRouter();
  const { t } = useTranslation();

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

  const handleLogout = async () => {
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

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="h-8 w-8 rounded-lg">
                <AvatarImage src={user.avatar} alt={user.username} />
                <AvatarFallback className="rounded-lg">
                  {user.username
                    .split(' ')
                    .map((n) => n[0])
                    .join('')
                    .toUpperCase()
                    .slice(0, 2)}
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{user.username}</span>
                <span className="truncate text-xs">{user.email}</span>
              </div>
              <ChevronsUpDown className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? 'bottom' : 'right'}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="h-8 w-8 rounded-lg">
                  <AvatarImage src={user.avatar} alt={user.username} />
                  <AvatarFallback className="rounded-lg">
                    {user.username
                      .split(' ')
                      .map((n) => n[0])
                      .join('')
                      .toUpperCase()
                      .slice(0, 2)}
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">{user.username}</span>
                  <span className="truncate text-xs">{user.email}</span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleSponsor}>
              <Heart className="text-red-500" />
              {t('user.menu.sponsor')}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleHelp}>
              <HelpCircle />
              {t('user.menu.help')}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleReportIssue}>
              <AlertCircle />
              {t('user.menu.reportIssue')}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              <LogOut />
              {t('user.menu.logout')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
