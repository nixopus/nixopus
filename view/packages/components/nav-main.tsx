'use client';

import React from 'react';
import { ChevronRight, ChartColumnDecreasing, DatabaseBackup, Layers, Server } from 'lucide-react';
import { usePathname, useRouter } from 'next/navigation';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@nixopus/ui';
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  useSidebar
} from '@/components/ui/sidebar';
import { SidebarHoverMenu } from '@/components/ui/sidebar-hover-menu';
import Link from 'next/link';
import { useCollapsibleState } from '@/packages/hooks/shared/use-collapsible-state';
import { TopNavMainProps } from '@/packages/types/layout';
import { useGetServersQuery } from '@/redux/services/servers/serversApi';
import { cn } from '@/lib/utils';

export function NavMain({ items, onItemClick }: TopNavMainProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { isItemCollapsed, toggleItem } = useCollapsibleState();
  const { state } = useSidebar();

  const handleClick = (url: string) => {
    onItemClick?.(url);
    router.push(url);
  };

  const machineMatch = pathname.match(/^\/machines\/[^/]+/);
  const effectivePath = machineMatch ? pathname.replace(machineMatch[0], '') || '/' : pathname;

  const isItemActive = (url: string) => {
    const checkPath = url.startsWith('/machines/') ? pathname : effectivePath;
    return checkPath === url || checkPath.startsWith(url + '/');
  };

  return (
    <>
      <SidebarGroup>
        <SidebarMenu>
          {items.map((item) => {
            const hasNestedItems = (item.items?.length || 0) > 0;
            const isCollapsed = state === 'collapsed';
            const itemIsActive = item.isActive || isItemActive(item.url);

            const hasActiveSubItem =
              hasNestedItems && item.items?.some((subItem) => isItemActive(subItem.url));

            if (hasNestedItems && isCollapsed) {
              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarHoverMenu items={item.items || []}>
                    <SidebarMenuButton
                      className="cursor-pointer"
                      tooltip={item.title}
                      isActive={itemIsActive || hasActiveSubItem}
                      onClick={() => handleClick(item.url)}
                    >
                      {item.icon && <item.icon />}
                      <span>{item.title}</span>
                    </SidebarMenuButton>
                  </SidebarHoverMenu>
                </SidebarMenuItem>
              );
            }

            return (
              <Collapsible
                key={item.title}
                asChild
                open={!isItemCollapsed(item.title)}
                onOpenChange={() => toggleItem(item.title)}
                className="group/collapsible"
              >
                <SidebarMenuItem>
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton
                      className="cursor-pointer"
                      tooltip={item.title}
                      isActive={itemIsActive || hasActiveSubItem}
                      onClick={() => handleClick(item.url)}
                    >
                      {item.icon && <item.icon />}
                      <span>{item.title}</span>
                      {hasNestedItems && (
                        <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                      )}
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                  {hasNestedItems && (
                    <CollapsibleContent>
                      <SidebarMenuSub>
                        {item.items?.map((subItem, index) => {
                          const subItemIsActive = isItemActive(subItem.url);
                          const prevSection =
                            index > 0 ? item.items?.[index - 1]?.section : undefined;
                          const showSectionLabel =
                            subItem.section && subItem.section !== prevSection;

                          return (
                            <React.Fragment key={subItem.title}>
                              {showSectionLabel && (
                                <li className="pt-3 pb-1 first:pt-1">
                                  <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/70">
                                    {subItem.section}
                                  </span>
                                </li>
                              )}
                              <SidebarMenuSubItem>
                                <SidebarMenuSubButton asChild isActive={subItemIsActive}>
                                  <Link href={subItem.url}>
                                    <span>{subItem.title}</span>
                                  </Link>
                                </SidebarMenuSubButton>
                              </SidebarMenuSubItem>
                            </React.Fragment>
                          );
                        })}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  )}
                </SidebarMenuItem>
              </Collapsible>
            );
          })}
        </SidebarMenu>
      </SidebarGroup>
      <MachineNav />
    </>
  );
}

function MachineNav() {
  const router = useRouter();
  const { data } = useGetServersQuery({ page: 1, page_size: 100 });
  const servers = data?.servers ?? [];

  if (servers.length === 0) return null;

  const machineOptions = [
    { label: 'Charts', icon: ChartColumnDecreasing, path: 'charts' },
    { label: 'Backups', icon: DatabaseBackup, path: 'backups' },
    { label: 'Apps', icon: Layers, path: 'apps' }
  ];

  return (
    <SidebarGroup>
      <SidebarGroupLabel>Machines</SidebarGroupLabel>
      <SidebarMenu>
        {servers.map((server) => (
          <SidebarMenuItem key={server.id}>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton className="cursor-pointer">
                  <Server className="size-4" />
                  <span className="truncate">{server.name || server.host}</span>
                  <span
                    className={cn(
                      'ml-auto size-2 rounded-full shrink-0',
                      server.is_active ? 'bg-green-500' : 'bg-red-500'
                    )}
                  />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="right" align="start" className="w-48">
                {machineOptions.map((opt) => (
                  <DropdownMenuItem
                    key={opt.path}
                    onClick={() => router.push(`/machines/${server.id}/${opt.path}`)}
                    className="cursor-pointer"
                  >
                    <opt.icon className="mr-2 size-4" />
                    {opt.label}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
