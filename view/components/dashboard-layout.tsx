import { AppSidebar } from '@/components/app-sidebar';
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
import { CreateTeam } from './create-team';
import useTeamSwitcher from '@/hooks/use-team-switcher';
import use_bread_crumbs from '@/hooks/use_bread_crumbs';
import React from 'react';

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
  const { getBreadcrumbs } = use_bread_crumbs();
  const breadcrumbs = React.useMemo(() => getBreadcrumbs(), [getBreadcrumbs]);

  return (
    <SidebarProvider>
      <AppSidebar toggleAddTeamModal={toggleAddTeamModal} />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
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
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          {children}
          {addTeamModalOpen && (
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
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
