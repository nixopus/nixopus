import React from 'react';
import { Sidebar } from '@/components/ui/sidebar';
import { User } from '@/redux/types/user';
import { Organization } from '@/redux/types/orgs';
import type { ComponentType } from 'react';
import { LucideIcon } from 'lucide-react';
import { translationKey } from '@/packages/hooks/shared/use-translation';

export interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  user: User | null;
  activeOrg: Organization | null;
  hasAnyPermission: (resource: string) => boolean;
  activeNav: string;
  refetch: () => void;
  t: (key: translationKey) => string;
  filteredNavItems: SideNav[];
  setActiveNav: (url: string) => void;
}

interface SideNav {
  title: string;
  url: string;
  icon: LucideIcon | ComponentType<{ className?: string }>;
  resource: string;
  items?: SideNavItem[];
}

interface SideNavItem {
  title: string;
  url: string;
  resource?: string;
}

// Topbar
export interface AppTopBarProps {
  breadcrumbs: BreadCrumbType[];
  isTerminalOpen: boolean;
  toggleTerminal: () => void;
  t: (key: translationKey) => string;
  startTour: () => void;
  user: User | null;
  onLogout: () => void;
}

export interface BreadCrumbsProps {
  breadcrumbs: BreadCrumbType[];
}

interface BreadCrumbType {
  label: string;
  href: string;
  external?: boolean;
}

// Sidebar
export interface TopNavMainProps {
  items: TopNavItem[];
  onItemClick?: (url: string) => void;
}

interface TopNavItem {
  title: string;
  url: string;
  icon?: LucideIcon | ComponentType<{ className?: string }>;
  isActive?: boolean;
  items?: {
    title: string;
    url: string;
  }[];
}

export enum TERMINAL_POSITION {
  BOTTOM = 'bottom',
  RIGHT = 'right'
}

// Dashboard
export interface DashboardItem {
  id: string;
  component: React.JSX.Element;
  className?: string;
  isDefault: boolean;
}
