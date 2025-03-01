"use client"

import * as React from "react"
import { ChevronsUpDown, GroupIcon, Plus } from "lucide-react"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { UserOrganization } from "@/redux/types/orgs"
import { useAppDispatch, useAppSelector } from "@/redux/hooks"
import { setActiveOrganization } from "@/redux/features/users/userSlice"

export function TeamSwitcher({
  teams,
  toggleAddTeamModal
}: {
  teams?: UserOrganization[] | [],
  toggleAddTeamModal?: () => void
}) {
  const { isMobile } = useSidebar()
  const user = useAppSelector(state => state.auth.user)
  const isAdmin = React.useMemo(() => user.type === "admin", [user])
  const activeTeam = useAppSelector(state => state.user.activeOrganization)
  const dispatch = useAppDispatch()

  React.useEffect(() => {
    dispatch(setActiveOrganization(teams?.[0]))
  }, [teams])

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                <GroupIcon className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{activeTeam?.organization.name}</span>
                <span className="truncate text-xs">{activeTeam?.organization.description}</span>
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            align="start"
            side={isMobile ? "bottom" : "right"}
            sideOffset={4}
          >
            <DropdownMenuLabel className="text-muted-foreground text-xs">
              Teams
            </DropdownMenuLabel>
            {(teams || []).map((team, index) => (
              <DropdownMenuItem
                key={team.organization.id}
                onClick={() => dispatch(setActiveOrganization(team))}
                className="gap-2 p-2"
              >
                <div className="flex size-6 items-center justify-center rounded-xs border">
                  <GroupIcon className="size-4 shrink-0" />
                </div>
                {team.organization.name}
                <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            {
              isAdmin && <DropdownMenuItem className="gap-2 p-2" onClick={toggleAddTeamModal}>
                <div className="bg-background flex size-6 items-center justify-center rounded-md border">
                  <Plus className="size-4" />
                </div>
                <div className="text-muted-foreground font-medium">Add team</div>
              </DropdownMenuItem>
            }
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
