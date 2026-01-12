'use client';

import React from 'react';
import { Activity, BarChart, Clock, HardDrive, Plus } from 'lucide-react';
import { ArrowDownCircle, ArrowUpCircle, Network, Server } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { TypographyMuted } from '@/components/ui/typography';
import useSmtpBanner from '@/packages/hooks/dashboard/use-smtp-banner';
import { X } from 'lucide-react';
import { ArrowRight, Package } from 'lucide-react';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { DataTable } from '@/components/ui/data-table';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRouter } from 'next/navigation';
import { ContainersWidgetProps } from '@/packages/types/containers';
import useClock from '@/packages/hooks/dashboard/use-clock';
import {
  createCPUChartData,
  createCPUChartConfig,
  createLoadAverageChartData,
  createLoadAverageChartConfig,
  createMemoryChartData,
  createMemoryChartConfig
} from '@/packages/utils/dashboard';
import { formatGB, formatPercentage } from '@/packages/utils/dashboard';
import { CHART_COLORS } from '@/packages/utils/dashboard';
import { Cpu } from 'lucide-react';
import { DEFAULT_METRICS } from '@/packages/utils/dashboard';
import { useSystemMetric } from '@/packages/hooks/dashboard/use-system-metric';
import { BarChartComponent } from '@/components/ui/bar-chart-component';
import { TypographySmall } from '@/components/ui/typography';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { CardTitle } from '@/components/ui/card';
import { DraggableGrid } from '@/components/ui/draggable-grid';
import { DoughnutChartComponent } from '@/components/ui/doughnut-chart-component';
import { useNetwork } from '@/packages/hooks/dashboard/use-network';
import { useDiskMountsColumns } from '@/packages/hooks/dashboard/use-disk-mounts-columns';
import { useSystemInfoItems } from '@/packages/hooks/dashboard/use-system-info-items';
import { useSystemStatsItems } from '@/packages/hooks/dashboard/use-system-stats-items';
import { useTopCores } from '@/packages/hooks/dashboard/use-top-cores';
import { useLoadAverageItems } from '@/packages/hooks/dashboard/use-load-average-items';
import {
  WidgetSelectorProps,
  CPUUsageCardProps,
  CPUUsageHeaderProps,
  CPUUsageChartProps,
  TopCoresListProps,
  CoreItemProps,
  CPU_COLORS,
  DiskUsageCardProps,
  MountData,
  LoadAverageCardProps,
  MemoryUsageCardProps,
  NetworkWidgetProps,
  SystemInfoCardProps,
  SystemInfoItemProps,
  SystemMetricCardProps,
  SystemStatsProps
} from '@/packages/types/dashboard';
import {
  ClockCardSkeletonContent,
  CPUUsageCardSkeletonContent,
  DiskUsageCardSkeletonContent,
  LoadAverageCardSkeleton,
  LoadAverageCardSkeletonContent,
  MemoryUsageCardSkeleton,
  MemoryUsageCardSkeletonContent,
  SystemInfoCardSkeleton
} from './dashboard-skeletons';
import { NetworkCardSkeletonContent } from './dashboard-skeletons';
import { CPUUsageCardSkeleton } from './dashboard-skeletons';

export function WidgetSelector({ availableWidgets, onAddWidget }: WidgetSelectorProps) {
  if (availableWidgets.length === 0) {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Plus className="h-4 w-4" />
          Add Widget
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        {availableWidgets.map((widget) => (
          <DropdownMenuItem
            key={widget.id}
            onClick={() => onAddWidget(widget.id)}
            className="cursor-pointer"
          >
            {widget.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function SMTPBanner() {
  const { handleDismiss, handleConfigure, t, isVisible } = useSmtpBanner();

  if (!isVisible) return null;

  return (
    <Alert className="mb-4">
      <AlertDescription className="flex items-center justify-between">
        <TypographyMuted>{t('dashboard.smtpBanner.message')}</TypographyMuted>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleConfigure}>
            {t('dashboard.smtpBanner.configure')}
          </Button>
          <Button variant="ghost" size="sm" onClick={handleDismiss}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  );
}

export const ContainersWidget: React.FC<ContainersWidgetProps> = ({ containersData, columns }) => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <CardWrapper
      title={t('dashboard.containers.title')}
      icon={Package}
      compact
      actions={
        <Button variant="outline" size="sm" onClick={() => router.push('/containers')}>
          <ArrowRight className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          {t('dashboard.containers.viewAll')}
        </Button>
      }
    >
      <DataTable
        data={containersData}
        columns={columns}
        emptyMessage={t('dashboard.containers.table.noContainers')}
        showBorder={false}
        hoverable={false}
      />
    </CardWrapper>
  );
};

export const ClockWidget: React.FC = () => {
  const { formattedTime, formattedDate } = useClock();

  return (
    <SystemMetricCard
      title="Clock"
      icon={Clock}
      isLoading={false}
      skeletonContent={<ClockCardSkeletonContent />}
    >
      <div className="flex flex-col items-center justify-center h-full space-y-3">
        <div className="text-5xl font-bold text-primary tabular-nums">{formattedTime}</div>
        <div className="text-sm text-muted-foreground">{formattedDate}</div>
      </div>
    </SystemMetricCard>
  );
};

export const CPUUsageHeader: React.FC<CPUUsageHeaderProps> = ({ overallUsage, label }) => {
  return (
    <div className="text-center">
      <TypographyMuted className="text-xs">{label}</TypographyMuted>
      <div className="text-3xl font-bold text-primary mt-1">{formatPercentage(overallUsage)}%</div>
    </div>
  );
};

export const CPUUsageChart: React.FC<CPUUsageChartProps> = ({
  chartData,
  chartConfig,
  yAxisLabel,
  xAxisLabel
}) => {
  return (
    <div>
      <BarChartComponent
        data={chartData}
        chartConfig={chartConfig}
        height="h-[180px]"
        yAxisLabel={yAxisLabel}
        xAxisLabel={xAxisLabel}
        showAxisLabels={true}
      />
    </div>
  );
};

export const CoreItem: React.FC<CoreItemProps> = ({ coreId, usage, color }) => {
  return (
    <div className="flex flex-col items-center gap-1">
      <div className="flex items-center gap-1">
        <div className="h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
        <TypographyMuted className="text-xs">Core {coreId}</TypographyMuted>
      </div>
      <TypographySmall className="text-sm font-bold">{formatPercentage(usage)}%</TypographySmall>
    </div>
  );
};

export const TopCoresList: React.FC<TopCoresListProps> = ({ cores }) => {
  return (
    <div className="grid grid-cols-3 gap-2 text-center">
      {cores.map((core) => {
        const color = CPU_COLORS[core.core_id % CPU_COLORS.length];
        return (
          <CoreItem key={core.core_id} coreId={core.core_id} usage={core.usage} color={color} />
        );
      })}
    </div>
  );
};

export const CPUUsageCard: React.FC<CPUUsageCardProps> = ({ systemStats }) => {
  const {
    data: cpu,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.cpu,
    defaultData: DEFAULT_METRICS.cpu
  });

  const perCoreData = cpu.per_core;
  const chartData = createCPUChartData(perCoreData);
  const chartConfig = createCPUChartConfig(perCoreData.length);
  const topCores = useTopCores(perCoreData, 3);

  return (
    <SystemMetricCard
      title={t('dashboard.cpu.title')}
      icon={Cpu}
      isLoading={isLoading}
      skeletonContent={<CPUUsageCardSkeletonContent />}
    >
      <div className="space-y-4">
        <CPUUsageHeader overallUsage={cpu.overall} label={t('dashboard.cpu.overall')} />

        <CPUUsageChart
          chartData={chartData}
          chartConfig={chartConfig}
          yAxisLabel={t('dashboard.cpu.usage')}
          xAxisLabel={t('dashboard.cpu.cores')}
        />

        <TopCoresList cores={topCores} />
      </div>
    </SystemMetricCard>
  );
};

export const DiskUsageCard: React.FC<DiskUsageCardProps> = ({ systemStats }) => {
  const {
    data: disk,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.disk,
    defaultData: DEFAULT_METRICS.disk
  });

  return (
    <SystemMetricCard
      title={t('dashboard.disk.title')}
      icon={HardDrive}
      isLoading={isLoading}
      skeletonContent={<DiskUsageCardSkeletonContent />}
    >
      <div className="space-y-2 sm:space-y-3">
        <div className="w-full h-2 bg-gray-200 rounded-full">
          <div className={`h-2 rounded-full bg-primary`} style={{ width: `${disk.percentage}%` }} />
        </div>
        <div className="flex justify-between">
          <TypographyMuted className="text-xs truncate max-w-[80px] sm:max-w-[100px]">
            {t('dashboard.disk.used').replace('{value}', disk.used.toFixed(2))}
          </TypographyMuted>
          <TypographyMuted className="text-xs truncate max-w-[60px] sm:max-w-[80px]">
            {t('dashboard.disk.percentage').replace('{value}', formatPercentage(disk.percentage))}
          </TypographyMuted>
          <TypographyMuted className="text-xs truncate max-w-[80px] sm:max-w-[100px]">
            {t('dashboard.disk.total').replace('{value}', disk.total.toFixed(2))}
          </TypographyMuted>
        </div>
        <div className="text-xs font-mono mt-1 sm:mt-2">
          <DiskMountsTable mounts={disk.allMounts} />
        </div>
      </div>
    </SystemMetricCard>
  );
};

export function DiskMountsTable({ mounts }: { mounts: MountData[] }) {
  const { t } = useTranslation();
  const columns = useDiskMountsColumns();

  return (
    <div
      className="max-h-[300px] overflow-y-auto overflow-x-hidden scrollbar-accessible"
      role="region"
      aria-label={`${t('dashboard.disk.table.headers.mount')} table with ${mounts.length} ${mounts.length === 1 ? 'mount point' : 'mount points'}`}
      aria-live="polite"
      tabIndex={0}
    >
      <DataTable
        data={mounts}
        columns={columns}
        tableClassName="min-w-full"
        containerClassName="overflow-x-hidden"
        showBorder={false}
        hoverable={false}
        striped={false}
      />
    </div>
  );
}

export const LoadAverageCard: React.FC<LoadAverageCardProps> = ({ systemStats }) => {
  const {
    data: load,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.load,
    defaultData: DEFAULT_METRICS.load
  });

  const chartData = createLoadAverageChartData(load);
  const chartConfig = createLoadAverageChartConfig();
  const loadAverageItems = useLoadAverageItems(load);

  return (
    <SystemMetricCard
      title={t('dashboard.load.title')}
      icon={Activity}
      isLoading={isLoading}
      skeletonContent={<LoadAverageCardSkeletonContent />}
    >
      <br />
      <br />
      <br />
      <div className="space-y-4">
        <div>
          <BarChartComponent
            data={chartData}
            chartConfig={chartConfig}
            height="h-[180px]"
            yAxisLabel="Load"
            xAxisLabel="Time Period"
            showAxisLabels={true}
          />
        </div>

        <div className="grid grid-cols-3 gap-2 text-center">
          {loadAverageItems.map((item) => (
            <div key={item.label} className="flex flex-col items-center gap-1">
              <div className="flex items-center gap-1">
                <div className="h-2 w-2 rounded-full" style={{ backgroundColor: item.color }} />
                <TypographyMuted className="text-xs">{item.label}</TypographyMuted>
              </div>
              <TypographySmall className="text-sm font-bold">
                {item.value.toFixed(2)}
              </TypographySmall>
            </div>
          ))}
        </div>
      </div>
    </SystemMetricCard>
  );
};

export const MemoryUsageCard: React.FC<MemoryUsageCardProps> = ({ systemStats }) => {
  const {
    data: memory,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.memory,
    defaultData: DEFAULT_METRICS.memory
  });

  const freeMemory = memory.total - memory.used;

  const chartData = createMemoryChartData(memory.used, freeMemory);
  const chartConfig = createMemoryChartConfig();

  return (
    <SystemMetricCard
      title={t('dashboard.memory.title')}
      icon={BarChart}
      isLoading={isLoading}
      skeletonContent={<MemoryUsageCardSkeletonContent />}
    >
      <div className="space-y-4">
        <div className="flex items-center justify-center h-[200px]">
          <DoughnutChartComponent
            data={chartData}
            chartConfig={chartConfig}
            centerLabel={{
              value: `${memory.percentage.toFixed(1)}%`,
              subLabel: 'Used'
            }}
            innerRadius={60}
            outerRadius={80}
            maxHeight="max-h-[200px]"
          />
        </div>

        <div className="space-y-2">
          <div className="flex justify-between text-xs">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: CHART_COLORS.blue }} />
              <TypographyMuted>Used: {formatGB(memory.used)} GB</TypographyMuted>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: CHART_COLORS.green }} />
              <TypographyMuted>Free: {formatGB(freeMemory)} GB</TypographyMuted>
            </div>
          </div>

          <TypographyMuted className="text-xs text-center">
            Total: {formatGB(memory.total)} GB
          </TypographyMuted>
        </div>
      </div>
    </SystemMetricCard>
  );
};

export const NetworkWidget: React.FC<NetworkWidgetProps> = ({ systemStats }) => {
  const { networkData } = useNetwork({ systemStats });

  const isLoading = !systemStats || !systemStats.network;

  return (
    <SystemMetricCard
      title="Network Traffic"
      icon={Network}
      isLoading={isLoading}
      skeletonContent={<NetworkCardSkeletonContent />}
    >
      <div className="flex flex-col items-center justify-center h-full space-y-4">
        <div className="grid grid-cols-2 gap-4 w-full">
          <div className="flex flex-col items-center text-center">
            <ArrowDownCircle className="h-8 w-8 text-blue-500 mb-2" />
            <div className="text-xs text-muted-foreground mb-1">Download</div>
            <div className="text-2xl font-bold text-primary tabular-nums">
              {networkData.downloadSpeed}
            </div>
          </div>
          <div className="flex flex-col items-center text-center">
            <ArrowUpCircle className="h-8 w-8 text-green-500 mb-2" />
            <div className="text-xs text-muted-foreground mb-1">Upload</div>
            <div className="text-2xl font-bold text-primary tabular-nums">
              {networkData.uploadSpeed}
            </div>
          </div>
        </div>
        <div className="flex gap-4 text-xs text-muted-foreground">
          <span>↓ {networkData.totalDownload}</span>
          <span>↑ {networkData.totalUpload}</span>
        </div>
      </div>
    </SystemMetricCard>
  );
};

export const SystemInfoItem: React.FC<SystemInfoItemProps> = ({ icon, label, value }) => {
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

export const SystemInfoCard: React.FC<SystemInfoCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();

  if (!systemStats) {
    return <SystemInfoCardSkeleton />;
  }

  const systemInfoItems = useSystemInfoItems(systemStats);

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

export const SystemMetricCard: React.FC<SystemMetricCardProps> = ({
  title,
  icon: Icon,
  isLoading = false,
  children,
  skeletonContent
}) => {
  const content = isLoading && skeletonContent ? skeletonContent : children;

  return (
    <Card className="overflow-hidden h-full flex flex-col">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Icon className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{title}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1">{content}</CardContent>
    </Card>
  );
};

export const SystemStats: React.FC<SystemStatsProps> = ({ systemStats }) => {
  if (!systemStats) {
    return (
      <div className="space-y-4">
        <SystemInfoCardSkeleton />
        <LoadAverageCardSkeleton />
        <CPUUsageCardSkeleton />
        <MemoryUsageCardSkeleton />
      </div>
    );
  }

  const systemStatsItems = useSystemStatsItems(systemStats, {
    SystemInfoCard,
    LoadAverageCard,
    CPUUsageCard,
    MemoryUsageCard
  });

  return <DraggableGrid items={systemStatsItems} storageKey="system-stats-card-order" />;
};
