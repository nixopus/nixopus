'use client';

import { useMemo } from 'react';
import { Edit, Terminal, FolderOpen, Trash2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { MenuItemConfig, UseServerCardProps } from '../types';

export function useServerCard({ server, onEdit, onDelete }: UseServerCardProps) {
  const { t } = useTranslation();

  const menuItems: MenuItemConfig[] = useMemo(
    () => [
      {
        key: 'edit',
        icon: Edit,
        label: t('common.edit'),
        onClick: () => onEdit(server)
      },
      {
        key: 'terminal',
        icon: Terminal,
        label: t('servers.actions.openTerminal'),
        disabled: server.status !== 'online'
      },
      {
        key: 'files',
        icon: FolderOpen,
        label: t('servers.actions.fileManager'),
        disabled: server.status !== 'online'
      },
      {
        key: 'delete',
        icon: Trash2,
        label: t('common.delete'),
        onClick: () => onDelete(server),
        destructive: true,
        separator: true
      }
    ],
    [t, server, onEdit, onDelete]
  );

  const visibleLabels = server.labels.slice(0, 2);
  const remainingLabelsCount = server.labels.length > 2 ? server.labels.length - 2 : 0;

  const getMetricColor = (value: number): string => {
    if (value < 50) return 'text-emerald-500';
    if (value < 75) return 'text-amber-500';
    return 'text-red-500';
  };

  return {
    t,
    menuItems,
    visibleLabels,
    remainingLabelsCount,
    getMetricColor
  };
}
