'use client';

import { ChevronRight } from 'lucide-react';
import { usePathname, useRouter } from 'next/navigation';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
  SidebarGroup,
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

export function NavMain({ items, onItemClick }: TopNavMainProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { isItemCollapsed, toggleItem } = useCollapsibleState();
  const { state } = useSidebar();

  const handleClick = (url: string) => {
    onItemClick?.(url);
    router.push(url);
  };

  const isItemActive = (url: string) => {
    if (pathname === url) return true;
    if (pathname.startsWith(url + '/')) return true;
    return false;
  };

  return (
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
                      {item.items?.map((subItem) => {
                        const subItemIsActive = isItemActive(subItem.url);
                        return (
                          <SidebarMenuSubItem key={subItem.title}>
                            <SidebarMenuSubButton asChild isActive={subItemIsActive}>
                              <Link href={subItem.url}>
                                <span>{subItem.title}</span>
                              </Link>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
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
  );
}
