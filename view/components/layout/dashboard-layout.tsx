'use client';
import { AppSidebar } from '@/components/layout/app-sidebar';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator
} from '@/components/ui/breadcrumb';
import { Separator } from '@/components/ui/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import { useRouter } from 'next/navigation';
import { CreateTeam } from '@/components/features/create-team';
import { KeyboardShortcuts } from '@/components/features/keyboard-shortcuts';
import useTeamSwitcher from '@/hooks/use-team-switcher';
import useBreadCrumbs from '@/hooks/use-bread-crumbs';
import React, { useEffect } from 'react';
import { Terminal } from '@/app/terminal/terminal';
import { useTerminalState } from '@/app/terminal/utils/useTerminalState';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Tour } from '@/components/Tour';
import { useTour } from '@/hooks/useTour';
import { Button } from '@/components/ui/button';
import { HelpCircle } from 'lucide-react';
import { AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { ModeToggler } from '@/components/ui/theme-toggler';
import { RBACGuard } from '@/components/rbac/RBACGuard';

enum TERMINAL_POSITION {
  BOTTOM = 'bottom',
  RIGHT = 'right'
}

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const {
    addTeamModalOpen,
    setAddTeamModalOpen,
    toggleAddTeamModal,
    createTeam,
    teamName,
    teamDescription,
    isLoading,
    handleTeamNameChange,
    handleTeamDescriptionChange
  } = useTeamSwitcher();

  const { getBreadcrumbs } = useBreadCrumbs();
  const breadcrumbs = React.useMemo(() => getBreadcrumbs(), [getBreadcrumbs]);
  const { isTerminalOpen, toggleTerminal } = useTerminalState();
  const [TerminalPosition, setTerminalPosition] = React.useState(TERMINAL_POSITION.BOTTOM);
  const [fitAddonRef, setFitAddonRef] = React.useState<any | null>(null);
  const { startTour } = useTour();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 't' && e.ctrlKey) {
        e.preventDefault();
        setTerminalPosition((prevPosition) =>
          prevPosition === TERMINAL_POSITION.BOTTOM
            ? TERMINAL_POSITION.RIGHT
            : TERMINAL_POSITION.BOTTOM
        );
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <SidebarProvider defaultOpen={true}>
      <AppSidebar toggleAddTeamModal={toggleAddTeamModal} addTeamModalOpen={addTeamModalOpen} />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4 justify-between w-full">
            <div className="flex items-center gap-2 px-4">
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
              {breadcrumbs.length > 0 && (
                <Breadcrumb>
                  <BreadcrumbList>
                    {breadcrumbs.map((breadcrumb, idx) => (
                      <React.Fragment key={idx}>
                        <BreadcrumbItem className="hidden md:block">
                          <BreadcrumbLink onClick={() => router.push(breadcrumb.href)}>
                            {breadcrumb.label}
                          </BreadcrumbLink>
                        </BreadcrumbItem>
                        {idx < breadcrumbs.length - 1 && (
                          <BreadcrumbSeparator className="hidden md:block" />
                        )}
                      </React.Fragment>
                    ))}
                  </BreadcrumbList>
                </Breadcrumb>
              )}
            </div>
            <div className="flex items-center gap-4">
              <Button
                variant="outline"
                size="icon"
                className="ml-auto"
                onClick={startTour}
                data-slot="tour-trigger"
              >
                <HelpCircle className="h-5 w-5" />
              </Button>
              <KeyboardShortcuts />
              <RBACGuard resource="user" action="update">
                <ModeToggler />
              </RBACGuard>
            </div>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <Tour>
            <div className="flex h-[calc(100vh-5rem)]">
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
              <ResizablePanelGroup
                direction={
                  TERMINAL_POSITION.BOTTOM === TerminalPosition ? 'vertical' : 'horizontal'
                }
                className="flex-grow"
              >
                <ResizablePanel
                  defaultSize={80}
                  minSize={30}
                  className="overflow-auto no-scrollbar"
                >
                  <div className="h-full overflow-y-auto no-scrollbar">{children}</div>
                </ResizablePanel>
                {isTerminalOpen && <ResizableHandle draggable withHandle />}
                <ResizablePanel
                  defaultSize={20}
                  minSize={15}
                  maxSize={50}
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
                    />
                  </AnyPermissionGuard>
                </ResizablePanel>
              </ResizablePanelGroup>
            </div>
          </Tour>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
