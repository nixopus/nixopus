import { ApplicationDeployment } from '@/redux/types/applications';
import React, { useState } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import PaginationWrapper from '@/components/ui/pagination';
import { DataTable } from '@/components/ui/data-table';
import { useDeploymentsList } from '@/packages/hooks/applications/use_deployments_list';
import { useDeploymentStatusChart } from '@/packages/hooks/applications/use_deployment_status_chart';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { BarChart, CartesianGrid, XAxis, YAxis } from 'recharts';
import { Bar } from 'recharts';
import { Application } from '@/redux/types/applications';
import { useDuplicateProject } from '@/packages/hooks/applications/use_duplicate_project';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import { Label } from '@/components/ui/label';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { Input } from '@/components/ui/input';
import { ResourceGuard } from '@/packages/components/rbac';
import SubPageHeader from '@/components/ui/sub-page-header';
import { useApplicationHeader } from '@/packages/hooks/applications/use_application_header';
import DeploymentLogsTable from '@/packages/components/deployment-logs';
import { Skeleton } from '@/components/ui/skeleton';
import { useProjectFamilySwitcher } from '@/packages/hooks/applications/use_project_family_switcher';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { useDeploymentOverview } from '@/packages/hooks/applications/use_deployment_overview';
import { useDeploymentHealthChart } from '@/packages/hooks/applications/use_deployment_health_chart';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';
import { Check, Copy } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useStatusIndicator } from '@/packages/hooks/applications/use_status_indicator';
import { useLatestDeployment } from '@/packages/hooks/applications/use_latest_deployment';
import { DraggableGrid } from '@/components/ui/draggable-grid';
import { useMonitoring } from '@/packages/hooks/applications/use-monitoring';
import { DeploymentsListProps, UseMonitoringReturn } from '../types/application';
import { MonitorProps } from '../types/application';
import { DeploymentOverviewProps } from '../types/application';
import { DeploymentHealthChartProps } from '../types/application';
import { LatestDeploymentProps } from '../types/application';
import { InfoLineProps } from '../types/application';
import { SectionLabelProps } from '../types/application';
import { StatBlockProps } from '../types/application';
import { StatusIndicatorProps } from '../types/application';
import { DuplicateProjectDialogProps } from '../types/application';
import { ApplicationLogsProps } from '../types/application';
import { ProjectFamilySwitcherProps } from '../types/application';

function DeploymentsList({
  deployments,
  currentPage,
  totalPages,
  onPageChange
}: DeploymentsListProps) {
  const { t } = useTranslation();
  const { columns, handleRowClick } = useDeploymentsList({ deployments });

  return (
    <div className="space-y-6">
      {deployments && deployments.length > 0 ? (
        <>
          <DataTable
            data={deployments}
            columns={columns}
            onRowClick={handleRowClick}
            showBorder={true}
            hoverable={true}
          />
          {totalPages > 1 && (
            <div className="mt-8 flex justify-center">
              <PaginationWrapper
                currentPage={currentPage}
                totalPages={totalPages}
                onPageChange={onPageChange}
              />
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-12 rounded-lg border">
          <p className="text-muted-foreground">{t('selfHost.deployment.list.noDeployments')}</p>
        </div>
      )}
    </div>
  );
}

export default DeploymentsList;

export const DeploymentStatusChart = ({
  deployments = []
}: {
  deployments: ApplicationDeployment[];
}) => {
  const { t } = useTranslation();
  const { statusCounts, statusChartConfig } = useDeploymentStatusChart({ deployments });

  return (
    <CardWrapper
      title={t('selfHost.deployments.chart.title')}
      description={t('selfHost.deployments.chart.description')}
      footer={t('selfHost.deployments.chart.total').replace(
        '{count}',
        deployments.length.toString()
      )}
      footerClassName="text-sm text-muted-foreground"
    >
      <ChartContainer config={statusChartConfig}>
        <BarChart accessibilityLayer data={statusCounts}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="status" type="category" tickLine={false} axisLine={false} width={100} />
          <YAxis type="number" tickLine={false} axisLine={false} />
          <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="dashed" />} />
          <Bar dataKey="value" className="fill-secondary" radius={4} />
        </BarChart>
      </ChartContainer>
    </CardWrapper>
  );
};

export function DuplicateProjectDialog({ application }: DuplicateProjectDialogProps) {
  const {
    open,
    setOpen,
    formFields,
    dialogActions,
    trigger,
    dialogTitle,
    dialogDescription,
    isLoading,
    isDisabled
  } = useDuplicateProject({ application });

  if (isDisabled) {
    return trigger;
  }

  return (
    <DialogWrapper
      open={open}
      onOpenChange={setOpen}
      title={dialogTitle}
      description={dialogDescription}
      trigger={trigger}
      actions={dialogActions}
      size="sm"
      loading={isLoading}
    >
      <div className="grid gap-4 py-4">
        {formFields.map((field) => (
          <div key={field.id} className="grid gap-2">
            <Label htmlFor={field.id}>{field.label}</Label>
            {field.type === 'select' ? (
              <SelectWrapper
                value={field.value}
                onValueChange={field.onChange}
                options={field.options}
                placeholder={field.placeholder}
                disabled={field.disabled}
                loading={field.loading}
              />
            ) : (
              <Input
                id={field.id}
                value={field.value}
                onChange={field.onChange}
                placeholder={field.placeholder}
              />
            )}
            {field.hint && <p className="text-xs text-muted-foreground">{field.hint}</p>}
          </div>
        ))}
      </div>
    </DialogWrapper>
  );
}

export const ApplicationDetailsHeader = ({ application }: { application?: Application }) => {
  const { icon, title, metadata, actions } = useApplicationHeader({ application });

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={null}>
      <SubPageHeader icon={icon} title={title} metadata={metadata} actions={actions} />
    </ResourceGuard>
  );
};

export const ApplicationLogs = ({ id }: ApplicationLogsProps) => {
  return <DeploymentLogsTable id={id} isDeployment={false} />;
};

export function Monitor({ application }: MonitorProps) {
  const monitoringHookResult = useMonitoring(application) as UseMonitoringReturn;
  const {
    showDragHint,
    mounted,
    layoutResetKey,
    hasCustomLayout,
    dismissHint,
    handleResetLayout,
    handleLayoutChange,
    monitoringItems
  } = monitoringHookResult;

  if (!application) {
    return null;
  }

  if (!mounted) {
    return (
      <ResourceGuard
        resource="deploy"
        action="read"
        loadingFallback={<Skeleton className="h-96" />}
      >
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Skeleton className="h-48 lg:col-span-2" />
          <Skeleton className="h-48" />
          <Skeleton className="h-48 lg:col-span-7" />
          <Skeleton className="h-48 lg:col-span-3" />
        </div>
      </ResourceGuard>
    );
  }

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <DraggableGrid
        items={monitoringItems}
        storageKey="monitoring-card-order"
        gridCols="grid-cols-1 lg:grid-cols-2"
        resetKey={layoutResetKey}
        onReorder={handleLayoutChange}
      />
    </ResourceGuard>
  );
}

export function ProjectFamilySwitcher({ application }: ProjectFamilySwitcherProps) {
  const { shouldShow, isLoading, trigger, dropdownContent } = useProjectFamilySwitcher({
    application
  });

  if (!shouldShow) {
    if (isLoading) {
      return <Skeleton className="h-8 w-8 rounded-md" />;
    }
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>{trigger}</DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        {dropdownContent}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function DeploymentOverview({
  totalDeployments,
  successfulDeployments,
  failedDeployments,
  currentStatus
}: DeploymentOverviewProps) {
  const { title, statBlocks } = useDeploymentOverview({
    totalDeployments,
    successfulDeployments,
    failedDeployments,
    currentStatus
  });

  return (
    <CardWrapper header={title}>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
        {statBlocks.map((block) => (
          <StatBlock
            key={block.key}
            value={block.value}
            label={block.label}
            sublabel={block.sublabel}
            color={block.color}
            pulse={block.pulse}
          />
        ))}
      </div>
    </CardWrapper>
  );
}

export function DeploymentHealthChart({
  deploymentsByStatus,
  totalDeployments,
  successRate
}: DeploymentHealthChartProps) {
  const { t } = useTranslation();
  const {
    chartData,
    customHeader,
    emptyStateContent,
    statusLegendItems,
    tooltipContent,
    hasData,
    successRate: chartSuccessRate
  } = useDeploymentHealthChart({
    deploymentsByStatus,
    totalDeployments,
    successRate
  });

  if (!hasData) {
    return (
      <CardWrapper title={t('selfHost.monitoring.chart.title')} className="h-full">
        {emptyStateContent}
      </CardWrapper>
    );
  }

  return (
    <CardWrapper header={customHeader} className="h-full">
      <div className="flex flex-col lg:flex-row items-center gap-8">
        <div className="relative w-64 h-64">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={chartData}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={100}
                paddingAngle={2}
                dataKey="value"
                strokeWidth={0}
              >
                {chartData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip content={tooltipContent} />
            </PieChart>
          </ResponsiveContainer>
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <span className="text-4xl font-bold">{chartSuccessRate}%</span>
            <span className="text-sm text-muted-foreground">
              {t('selfHost.monitoring.chart.successRate')}
            </span>
          </div>
        </div>

        <div className="flex-1 grid grid-cols-2 gap-4 w-full lg:w-auto">
          {statusLegendItems.map((item) => (
            <div key={item.status} className="flex items-center gap-3 p-3 rounded-lg bg-muted/50">
              <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
              <div>
                <p className="text-sm font-medium capitalize">{item.label}</p>
                <p className="text-2xl font-bold">{item.count}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </CardWrapper>
  );
}

export function LatestDeployment({ deployment }: LatestDeploymentProps) {
  const { t } = useTranslation();
  const { emptyStateContent, headerContent, infoLines, hasDeployment } = useLatestDeployment({
    deployment
  });

  if (!hasDeployment) {
    return (
      <CardWrapper
        className="h-full border-dashed"
        contentClassName="flex flex-col items-center justify-center h-full py-8 text-muted-foreground"
      >
        {emptyStateContent}
      </CardWrapper>
    );
  }

  return (
    <CardWrapper className="h-full" contentClassName="pt-6 h-full flex flex-col">
      {headerContent}
      <SectionLabel>{t('selfHost.monitoring.latestDeployment.title')}</SectionLabel>
      <div className="flex flex-col gap-y-2 mt-4 flex-1">
        {infoLines.map((line) => (
          <InfoLine
            key={line.key}
            icon={line.icon}
            label={line.label}
            value={line.value}
            displayValue={line.displayValue}
            sublabel={line.sublabel}
            mono={line.mono}
            copyable={line.copyable}
          />
        ))}
      </div>
    </CardWrapper>
  );
}

export function InfoLine({
  icon: Icon,
  label,
  value,
  displayValue,
  sublabel,
  mono,
  copyable
}: InfoLineProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex items-start gap-3 py-2">
      <Icon className="h-4 w-4 mt-1 text-muted-foreground flex-shrink-0" />
      <div className="flex-1 min-w-0">
        <p className="text-xs text-muted-foreground uppercase tracking-wide mb-0.5">{label}</p>
        <div className="flex items-center gap-2">
          <span className={cn('text-sm truncate', mono && 'font-mono')} title={value}>
            {displayValue || value}
          </span>
          {copyable && (
            <button
              onClick={handleCopy}
              className="text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
            >
              {copied ? (
                <Check className="h-3 w-3 text-emerald-500" />
              ) : (
                <Copy className="h-3 w-3" />
              )}
            </button>
          )}
        </div>
        {sublabel && <p className="text-xs text-muted-foreground/60 mt-0.5">{sublabel}</p>}
      </div>
    </div>
  );
}

export function SectionLabel({ children }: SectionLabelProps) {
  return (
    <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
      {children}
    </h3>
  );
}

export function StatBlock({ value, label, sublabel, color, pulse }: StatBlockProps) {
  const colorClasses = {
    emerald: 'text-emerald-500',
    red: 'text-red-500',
    amber: 'text-amber-500',
    blue: 'text-blue-500',
    purple: 'text-purple-500'
  };

  return (
    <div className="relative">
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          {pulse && (
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
          )}
          <span
            className={cn(
              'text-2xl font-bold tracking-tight capitalize',
              color && colorClasses[color]
            )}
          >
            {value}
          </span>
        </div>
        <p className="text-sm text-muted-foreground">{label}</p>
        {sublabel && <p className="text-xs text-muted-foreground/60">{sublabel}</p>}
      </div>
    </div>
  );
}

export function StatusIndicator({ status, size = 'md', showLabel = true }: StatusIndicatorProps) {
  const { indicatorDot, label, noStatusLabel, noStatusIndicatorDot, hasStatus } =
    useStatusIndicator({
      status,
      size,
      showLabel
    });

  if (!hasStatus) {
    return (
      <div className="flex items-center gap-2">
        {noStatusIndicatorDot}
        {noStatusLabel}
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      {indicatorDot}
      {label}
    </div>
  );
}
