import React from 'react';
import { Activity, BarChart, Box, CpuIcon, HardDrive, ServerCog, Terminal } from 'lucide-react';
import { ScreenShare, Server } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { SystemStatsType } from '@/redux/types/monitor';

export interface SystemInfoItem {
  icon: React.ReactNode;
  label: string;
  value: string;
}

export function useSystemInfoItems(systemStats: SystemStatsType): SystemInfoItem[] {
  const { t } = useTranslation();

  const { load, memory, os_type, cpu_info, cpu_cores, hostname, kernel_version, architecture } =
    systemStats;

  const memoryDisplay = `${memory.used.toFixed(1)} / ${memory.total.toFixed(1)} GB (${memory.percentage.toFixed(1)}%)`;

  return [
    {
      icon: <Server className="h-4 w-4 text-blue-500" />,
      label: t('dashboard.system.labels.operatingSystem'),
      value: os_type || 'N/A'
    },
    {
      icon: <ScreenShare className="h-4 w-4 text-purple-500" />,
      label: t('dashboard.system.labels.hostname'),
      value: hostname || 'N/A'
    },
    {
      icon: <ServerCog className="h-4 w-4 text-green-500" />,
      label: t('dashboard.system.labels.cpu'),
      value: cpu_info || 'N/A'
    },
    {
      icon: <CpuIcon className="h-4 w-4 text-teal-500" />,
      label: t('dashboard.system.labels.cpuCores'),
      value: cpu_cores > 0 ? `${cpu_cores} cores` : 'N/A'
    },
    {
      icon: <HardDrive className="h-4 w-4 text-orange-500" />,
      label: t('dashboard.system.labels.memory'),
      value: memoryDisplay
    },
    {
      icon: <Terminal className="h-4 w-4 text-sky-500" />,
      label: t('dashboard.system.labels.kernelVersion'),
      value: kernel_version || 'N/A'
    },
    {
      icon: <Activity className="h-4 w-4 text-emerald-500" />,
      label: t('dashboard.system.labels.uptime'),
      value: load.uptime?.replaceAll(/([hms])(\d)/g, '$1 $2')
    },
    {
      icon: <Box className="h-4 w-4 text-red-500" />,
      label: t('dashboard.system.labels.architecture'),
      value: architecture || 'N/A'
    }
  ];
}
