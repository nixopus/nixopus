import React from 'react';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { ExternalLink, GitBranch, Package } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Application } from '@/redux/types/applications';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';

function AppItem({
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
}: Application) {
  const router = useRouter();

  const latestDeployment = deployments?.[0];
  const currentStatus = latestDeployment?.status?.status || status?.status;

  const getStatusConfig = (statusValue?: string) => {
    switch (statusValue) {
      case 'deployed':
        return { bg: 'bg-emerald-500/10', dot: 'bg-emerald-500', pulse: true, label: 'Live' };
      case 'failed':
        return { bg: 'bg-red-500/10', dot: 'bg-red-500', pulse: false, label: 'Failed' };
      case 'building':
      case 'deploying':
      case 'cloning':
        return { bg: 'bg-amber-500/10', dot: 'bg-amber-500', pulse: true, label: 'Building' };
      case 'draft':
        return { bg: 'bg-blue-500/10', dot: 'bg-blue-500', pulse: false, label: 'Draft' };
      default:
        return { bg: 'bg-zinc-500/10', dot: 'bg-zinc-500', pulse: false, label: 'Inactive' };
    }
  };

  const statusConfig = getStatusConfig(currentStatus);

  const formattedBuildPack = build_pack
    .replace(/([A-Z])/g, ' $1')
    .trim()
    .toLowerCase();

  const getEnvironmentStyles = () => {
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

  const timeAgo = updated_at ? formatDistanceToNow(new Date(updated_at), { addSuffix: true }) : '';

  return (
    <Card
      className="relative w-full cursor-pointer overflow-hidden transition-all duration-200 hover:shadow-lg hover:border-primary/30 group"
      onClick={() => router.push(`/self-host/application/${id}`)}
    >
      <CardContent className="p-5">
        <div className="flex items-start gap-4">
          <div
            className={cn(
              'w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors',
              statusConfig.bg
            )}
          >
            <div
              className={cn(
                'w-2.5 h-2.5 rounded-full',
                statusConfig.dot,
                statusConfig.pulse && 'animate-pulse'
              )}
            />
          </div>

          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between gap-2">
              <h3 className="font-semibold text-base tracking-tight truncate group-hover:text-primary transition-colors">
                {name}
              </h3>
              {domain && (
                <a
                  href={`https://${domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-shrink-0 p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                  title={`Open ${domain}`}
                  onClick={(e) => e.stopPropagation()}
                >
                  <ExternalLink className="h-4 w-4" />
                </a>
              )}
            </div>

            <div className="flex flex-wrap items-center gap-2 mt-2">
              {domain && (
                <span className="text-xs text-muted-foreground font-mono bg-muted px-2 py-0.5 rounded truncate max-w-[180px]">
                  {domain}
                </span>
              )}
              <Badge variant="outline" className={cn('text-xs capitalize', getEnvironmentStyles())}>
                {environment}
              </Badge>
              {labels && labels.length > 0 && (
                <>
                  {labels.slice(0, 2).map((label, index) => (
                    <Badge
                      key={index}
                      variant="outline"
                      className="text-xs border-violet-500/30 text-violet-500 bg-violet-500/10"
                    >
                      {label}
                    </Badge>
                  ))}
                  {labels.length > 2 && (
                    <Badge
                      variant="outline"
                      className="text-xs border-muted-foreground/30 text-muted-foreground bg-muted"
                    >
                      +{labels.length - 2}
                    </Badge>
                  )}
                </>
              )}
            </div>

            <div className="flex items-center gap-4 mt-4 text-xs text-muted-foreground">
              {branch && (
                <span className="flex items-center gap-1">
                  <GitBranch className="h-3 w-3" />
                  {branch}
                </span>
              )}
              <span className="flex items-center gap-1">
                <Package className="h-3 w-3" />
                <span className="capitalize">{formattedBuildPack}</span>
              </span>
            </div>

            <div className="flex items-center justify-between mt-4 pt-3 border-t border-border/50">
              <span className="text-xs text-muted-foreground">{timeAgo}</span>
              <span
                className={cn(
                  'text-xs font-medium',
                  currentStatus === 'deployed'
                    ? 'text-emerald-500'
                    : currentStatus === 'failed'
                      ? 'text-red-500'
                      : currentStatus === 'draft'
                        ? 'text-blue-500'
                        : 'text-muted-foreground'
                )}
              >
                {statusConfig.label}
              </span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default AppItem;

export function AppItemSkeleton() {
  return (
    <Card className="relative w-full">
      <CardContent className="p-5">
        <div className="flex items-start gap-4">
          <Skeleton className="w-10 h-10 rounded-lg flex-shrink-0" />
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between gap-2">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-7 w-7 rounded-md" />
            </div>
            <Skeleton className="h-3 w-48 mt-2" />
            <div className="flex items-center gap-3 mt-3">
              <Skeleton className="h-5 w-20 rounded-full" />
              <Skeleton className="h-4 w-16" />
            </div>
            <div className="flex items-center justify-between mt-3 pt-3 border-t border-border/50">
              <Skeleton className="h-3 w-24" />
              <Skeleton className="h-3 w-12" />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
