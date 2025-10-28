'use client';

import React from 'react';
import {
  Server,
  HardDrive,
  Activity,
  Terminal,
  Box,
  CpuIcon,
  ScreenShare,
  ServerCog
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '@/redux/types/monitor';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { SystemInfoCardSkeleton } from './skeletons/system-info';

interface SystemInfoCardProps {
  systemStats: SystemStatsType | null;
}

interface SystemInfoItemProps {
  icon: React.ReactNode;
  label: string;
  value: string;
}

const SystemInfoItem: React.FC<SystemInfoItemProps> = ({ icon, label, value }) => {
  return (
    <div className="flex items-start gap-3 p-2 rounded-lg hover:bg-muted/50 transition-colors">
      <div className="mt-0.5">{icon}</div>
      <div className="flex-1 min-w-0">
        <TypographyMuted className="text-xs font-medium">{label}</TypographyMuted>
        <TypographySmall className="text-xs font-semibold truncate">{value}</TypographySmall>
      </div>
    </div>
  );
};

const SystemInfoCard: React.FC<SystemInfoCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();

  if (!systemStats) {
    return <SystemInfoCardSkeleton />;
  }

  const { load, memory, os_type, cpu_info, cpu_cores, hostname, kernel_version, architecture } =
    systemStats;

  const memoryDisplay = `${memory.used.toFixed(1)} / ${memory.total.toFixed(1)} GB (${memory.percentage.toFixed(1)}%)`;

  const systemInfoItems = [
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

  return (
    <Card className="overflow-hidden h-full flex flex-col w-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-bold flex items-center">
          <Server className="h-4 w-4 mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.system.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {systemInfoItems.map((item, index) => (
            <SystemInfoItem key={index} icon={item.icon} label={item.label} value={item.value} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
};

export default SystemInfoCard;
