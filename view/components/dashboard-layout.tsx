import { AppSidebar } from "@/components/app-sidebar"
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
    SidebarInset,
    SidebarProvider,
    SidebarTrigger,
} from "@/components/ui/sidebar"
import { useAppSelector } from "@/redux/hooks"
import { useRouter } from "next/navigation"
import { CreateTeam } from "./create-team"
import useTeamSwitcher from "@/hooks/use-team-switcher"

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
    const user = useAppSelector(state => state.auth.user)
    const router = useRouter()
    const {
        addTeamModalOpen,
        setAddTeamModalOpen,
        toggleAddTeamModal,
        createTeam,
        teamName,
        teamDescription,
        isLoading,
        handleTeamNameChange,
        handleTeamDescriptionChange } = useTeamSwitcher()

    if (!user) {
        router.push("/login")
        return null
    }
    return (
        <SidebarProvider>
            <AppSidebar toggleAddTeamModal={toggleAddTeamModal} />
            <SidebarInset>
                <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
                    <div className="flex items-center gap-2 px-4">
                        <SidebarTrigger className="-ml-1" />
                        <Separator
                            orientation="vertical"
                            className="mr-2 data-[orientation=vertical]:h-4"
                        />
                        <Breadcrumb>
                            <BreadcrumbList>
                                <BreadcrumbItem className="hidden md:block">
                                    <BreadcrumbLink href="#">
                                        Home
                                    </BreadcrumbLink>
                                </BreadcrumbItem>
                                <BreadcrumbSeparator className="hidden md:block" />
                                <BreadcrumbItem>
                                    <BreadcrumbPage>Data Fetching</BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                </header>
                <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
                    {children}
                    {
                        addTeamModalOpen && (
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
                        )
                    }
                </div>
            </SidebarInset>
        </SidebarProvider>
    )
}
