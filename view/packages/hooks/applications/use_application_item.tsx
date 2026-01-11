import React, { useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { formatDistanceToNow } from 'date-fns';
import { Application } from '@/redux/types/applications';
import { GitBranch, Package } from 'lucide-react';

interface StatusConfig {
  bg: string;
  dot: string;
  pulse: boolean;
  label: string;
}

const getStatusConfig = (statusValue?: string): StatusConfig => {
  switch (statusValue) {
    case 'deployed':
      return { bg: 'bg-emerald-500/10', dot: 'bg-emerald-500', pulse: true, label: 'Live' };
    case 'running':
      return { bg: 'bg-emerald-500/10', dot: 'bg-emerald-500', pulse: true, label: 'Running' };
    case 'failed':
      return { bg: 'bg-red-500/10', dot: 'bg-red-500', pulse: false, label: 'Failed' };
    case 'building':
    case 'deploying':
    case 'cloning':
    case 'started':
      return { bg: 'bg-amber-500/10', dot: 'bg-amber-500', pulse: true, label: 'Building' };
    case 'draft':
      return { bg: 'bg-blue-500/10', dot: 'bg-blue-500', pulse: false, label: 'Draft' };
    case 'stopped':
      return { bg: 'bg-zinc-500/10', dot: 'bg-zinc-500', pulse: false, label: 'Stopped' };
    default:
      return { bg: 'bg-zinc-500/10', dot: 'bg-zinc-500', pulse: false, label: 'Inactive' };
  }
};

const getEnvironmentStyles = (environment: string): string => {
  switch (environment) {
    case 'development':
      return 'border-blue-500/30 text-blue-500 bg-blue-500/10';
    case 'staging':
      return 'border-amber-500/30 text-amber-500 bg-amber-500/10';
    case 'production':
      return 'border-emerald-500/30 text-emerald-500 bg-emerald-500/10';
    default:
      return 'border-zinc-500/30 text-zinc-500 bg-zinc-500/10';
  }
};

const getStatusTextColor = (status?: string): string => {
  if (status === 'deployed' || status === 'running') return 'text-emerald-500';
  if (status === 'failed') return 'text-red-500';
  if (status === 'draft') return 'text-blue-500';
  if (
    status === 'building' ||
    status === 'deploying' ||
    status === 'cloning' ||
    status === 'started'
  )
    return 'text-amber-500';
  if (status === 'stopped') return 'text-zinc-500';
  return 'text-muted-foreground';
};

export function useApplicationItem(application: Application) {
  const router = useRouter();
  const {
    name,
    domain,
    environment,
    updated_at,
    build_pack,
    branch,
    id,
    status,
    deployments,
    labels
  } = application;

  const latestDeployment = deployments?.[0];
  const currentStatus = latestDeployment?.status?.status || status?.status;

  const statusConfig = useMemo(() => getStatusConfig(currentStatus), [currentStatus]);

  const formattedBuildPack = useMemo(
    () =>
      build_pack
        .replace(/([A-Z])/g, ' $1')
        .trim()
        .toLowerCase(),
    [build_pack]
  );

  const environmentStyles = useMemo(() => getEnvironmentStyles(environment), [environment]);

  const statusTextColor = useMemo(() => getStatusTextColor(currentStatus), [currentStatus]);

  const timeAgo = useMemo(
    () => (updated_at ? formatDistanceToNow(new Date(updated_at), { addSuffix: true }) : ''),
    [updated_at]
  );

  const metadataItems = useMemo(
    () =>
      [
        branch && {
          icon: GitBranch,
          label: branch,
          key: 'branch'
        },
        {
          icon: Package,
          label: formattedBuildPack,
          key: 'buildPack'
        }
      ].filter(Boolean) as Array<{
        icon: React.ComponentType<{ className?: string }>;
        label: string;
        key: string;
      }>,
    [branch, formattedBuildPack]
  );

  const displayLabels = useMemo(() => {
    if (!labels || labels.length === 0) return null;
    const visibleLabels = labels.slice(0, 2);
    const remainingCount = labels.length - 2;
    return {
      visible: visibleLabels,
      remainingCount: remainingCount > 0 ? remainingCount : 0
    };
  }, [labels]);

  const handleClick = () => {
    router.push(`/self-host/application/${id}`);
  };

  return {
    name,
    domain,
    currentStatus,
    statusConfig,
    environmentStyles,
    statusTextColor,
    timeAgo,
    metadataItems,
    displayLabels,
    handleClick
  };
}
