import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { HelpCircle, Settings, TerminalIcon } from 'lucide-react';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger
} from '@/components/ui/sidebar';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator
} from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { ModeToggler } from '@/components/ui/theme-toggler';
import { Separator } from '@/components/ui/separator';
import { TeamSwitcher } from '@/components/ui/team-switcher';
import { AnyPermissionGuard } from '@/packages/components/rbac';
import { CreateTeam } from '@/packages/components/team-settings';
import { NavMain } from '@/packages/components/nav-main';
import { RBACGuard } from '@/packages/components/rbac';
import { Terminal } from '@/packages/components/terminal';
import { TopbarWidgets } from '@/packages/components/topbar-widgets';
import { useSettingsModal } from '@/packages/hooks/shared/use-settings-modal';
import {
  AppSidebarProps,
  AppTopBarProps,
  BreadCrumbsProps,
  TERMINAL_POSITION
} from '@/packages/types/layout';
import { cn } from '@/lib/utils';
import { useLayout } from '@/packages/hooks/use-layout';
import { Tour } from './Tour';

interface LayoutProps {
  children: React.ReactNode;
}

function Layout({ children }: LayoutProps) {
  const {
    user,
    activeOrg,
    hasAnyPermission,
    activeNav,
    refetch,
    t,
    filteredNavItems,
    setActiveNav,
    addTeamModalOpen,
    setAddTeamModalOpen,
    createTeam,
    teamName,
    teamDescription,
    isLoading,
    handleTeamNameChange,
    handleTeamDescriptionChange,
    breadcrumbs,
    isTerminalOpen,
    toggleTerminal,
    TerminalPosition,
    setTerminalPosition,
    fitAddonRef,
    setFitAddonRef,
    startTour
  } = useLayout();

  return (
    <SidebarProvider defaultOpen={false}>
      <AppSidebar
        user={user}
        activeOrg={activeOrg}
        hasAnyPermission={hasAnyPermission}
        activeNav={activeNav}
        refetch={refetch}
        t={t}
        filteredNavItems={filteredNavItems}
        setActiveNav={setActiveNav}
      />
      <SidebarInset>
        <AppTopBar
          breadcrumbs={breadcrumbs}
          isTerminalOpen={isTerminalOpen}
          toggleTerminal={toggleTerminal}
          t={t}
          startTour={startTour}
        />
        <Tour>
          <CreateTeam
            open={addTeamModalOpen}
            setOpen={setAddTeamModalOpen}
            createTeam={createTeam}
            teamName={teamName}
            teamDescription={teamDescription}
            handleTeamNameChange={handleTeamNameChange}
            handleTeamDescriptionChange={handleTeamDescriptionChange}
            isLoading={isLoading}
          />
          <div className="w-full h-full flex-1">
            <ResizablePanelLayout
              TerminalPosition={TerminalPosition}
              isTerminalOpen={isTerminalOpen}
              fitAddonRef={fitAddonRef}
              toggleTerminal={toggleTerminal}
              setFitAddonRef={setFitAddonRef}
              setTerminalPosition={setTerminalPosition}
              children={children}
            />
          </div>
        </Tour>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default Layout;

interface ResizablePanelLayoutProps {
  children: React.ReactNode;
  TerminalPosition: TERMINAL_POSITION;
  isTerminalOpen: boolean;
  fitAddonRef: any;
  toggleTerminal: () => void;
  setFitAddonRef: (ref: any) => void;
  setTerminalPosition: React.Dispatch<React.SetStateAction<TERMINAL_POSITION>>;
}

export const ResizablePanelLayout = ({
  children,
  TerminalPosition,
  isTerminalOpen,
  fitAddonRef,
  toggleTerminal,
  setFitAddonRef,
  setTerminalPosition
}: ResizablePanelLayoutProps) => {
  return (
    <ResizablePanelGroup
      direction={TERMINAL_POSITION.BOTTOM === TerminalPosition ? 'vertical' : 'horizontal'}
      className="flex-grow"
    >
      <ResizablePanel defaultSize={65} minSize={30}>
        <div className="h-full w-full overflow-y-auto no-scrollbar">{children}</div>
      </ResizablePanel>
      {isTerminalOpen && <ResizableHandle draggable withHandle />}
      <ResizablePanel
        defaultSize={35}
        minSize={15}
        maxSize={60}
        hidden={!isTerminalOpen}
        onResize={() => {
          if (fitAddonRef?.current) {
            requestAnimationFrame(() => {
              fitAddonRef.current.fit();
            });
          }
        }}
        className="min-h-[200px] flex flex-col"
      >
        <AnyPermissionGuard
          permissions={['terminal:create', 'terminal:read', 'terminal:update']}
          loadingFallback={null}
        >
          <Terminal
            isOpen={isTerminalOpen}
            toggleTerminal={toggleTerminal}
            isTerminalOpen={isTerminalOpen}
            setFitAddonRef={setFitAddonRef}
            terminalPosition={TerminalPosition}
            onTogglePosition={() => {
              setTerminalPosition((prevPosition) => {
                const newPosition =
                  prevPosition === TERMINAL_POSITION.BOTTOM
                    ? TERMINAL_POSITION.RIGHT
                    : TERMINAL_POSITION.BOTTOM;
                localStorage.setItem('terminalPosition', newPosition);
                return newPosition;
              });
            }}
          />
        </AnyPermissionGuard>
      </ResizablePanel>
    </ResizablePanelGroup>
  );
};

function BreadCrumbs({ breadcrumbs }: BreadCrumbsProps) {
  const router = useRouter();
  return (
    <Breadcrumb>
      <BreadcrumbList>
        {' '}
        {breadcrumbs?.length > 0 &&
          breadcrumbs?.map((breadcrumb, idx) => (
            <React.Fragment key={idx}>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink
                  onClick={() => router.push(breadcrumb.href)}
                  className="cursor-pointer"
                >
                  {breadcrumb.label}{' '}
                </BreadcrumbLink>
              </BreadcrumbItem>
              {idx < breadcrumbs.length - 1 && (
                <BreadcrumbSeparator className="hidden md:block" />
              )}{' '}
            </React.Fragment>
          ))}{' '}
      </BreadcrumbList>
    </Breadcrumb>
  );
}

function AppTopBar({ breadcrumbs, isTerminalOpen, toggleTerminal, t, startTour }: AppTopBarProps) {
  return (
    <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
      <div className="flex items-center gap-2 px-4 justify-between w-full">
        <div className="flex items-center gap-2">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />{' '}
          {breadcrumbs?.length > 0 && <BreadCrumbs breadcrumbs={breadcrumbs} />}{' '}
        </div>
        <div className="flex items-center gap-4">
          <TopbarWidgets />
          <AnyPermissionGuard
            permissions={['terminal:create', 'terminal:read', 'terminal:update']}
            loadingFallback={null}
          >
            <Button
              variant={isTerminalOpen ? 'secondary' : 'outline'}
              size="icon"
              onClick={toggleTerminal}
              title={`${isTerminalOpen ? t('terminal.close') : t('terminal.title')} (${t(
                'terminal.shortcut'
              )})`}
              className={cn(
                'transition-all duration-200',
                isTerminalOpen && 'bg-primary/10 text-primary hover:bg-primary/20'
              )}
            >
              <TerminalIcon className="h-5 w-5" />
            </Button>
          </AnyPermissionGuard>
          <Button
            variant="outline"
            size="icon"
            className="ml-auto"
            onClick={startTour}
            data-slot="tour-trigger"
          >
            <HelpCircle className="h-5 w-5" />
          </Button>
          <RBACGuard resource="user" action="update">
            <ModeToggler />
          </RBACGuard>
        </div>
      </div>
    </header>
  );
}

export function AppSidebar({
  toggleAddTeamModal,
  addTeamModalOpen,
  user,
  activeOrg,
  hasAnyPermission,
  activeNav,
  refetch,
  t,
  filteredNavItems,
  setActiveNav,
  ...props
}: AppSidebarProps) {
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
