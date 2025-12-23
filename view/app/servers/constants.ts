import React from 'react';
import { Crown, HardDrive, Server } from 'lucide-react';
import type { ServerStatus, ServerRole, ServerHealth } from '@/redux/types/servers';

export const STATUS_COLORS: Record<ServerStatus, string> = {
  online: 'bg-emerald-500',
  offline: 'bg-zinc-500',
  connecting: 'bg-amber-500 animate-pulse',
  error: 'bg-red-500'
};

export const STATUS_TEXT_COLORS: Record<ServerStatus, string> = {
  online: 'text-emerald-500',
  offline: 'text-zinc-500',
  connecting: 'text-amber-500',
  error: 'text-red-500'
};

export const HEALTH_COLORS: Record<ServerHealth, string> = {
  healthy: 'text-emerald-500',
  degraded: 'text-amber-500',
  critical: 'text-red-500',
  unknown: 'text-zinc-400'
};

export const ROLE_ICONS: Record<ServerRole, React.ReactNode> = {
  manager: React.createElement(Crown, { className: 'h-4 w-4 text-amber-400' }),
  worker: React.createElement(HardDrive, { className: 'h-4 w-4 text-blue-400' }),
  standalone: React.createElement(Server, { className: 'h-4 w-4 text-zinc-400' })
};

export const ROLE_ICONS_SMALL: Record<ServerRole, React.ReactNode> = {
  manager: React.createElement(Crown, { className: 'h-3.5 w-3.5 text-amber-400' }),
  worker: React.createElement(HardDrive, { className: 'h-3.5 w-3.5 text-blue-400' }),
  standalone: React.createElement(Server, { className: 'h-3.5 w-3.5 text-zinc-400' })
};

export const ROLE_BADGE_COLORS: Record<ServerRole, string> = {
  manager: 'bg-amber-500/10 text-amber-400 border-amber-500/30',
  worker: 'bg-blue-500/10 text-blue-400 border-blue-500/30',
  standalone: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/30'
};

export const ROLE_BG_COLORS: Record<ServerRole, string> = {
  manager: 'bg-amber-500/10',
  worker: 'bg-blue-500/10',
  standalone: 'bg-zinc-500/10'
};

export const NODE_DIMENSIONS = {
  WIDTH: 260,
  HEIGHT: 200,
  HORIZONTAL_GAP: 80,
  VERTICAL_GAP: 120
} as const;
