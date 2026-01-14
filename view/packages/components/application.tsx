import React from 'react';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Application } from '@/redux/types/applications';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import { useApplicationItem } from '@/packages/hooks/applications/use_application_item';
import { DomainDropdown } from '@/packages/components/multi-domains';

function AppItem(application: Application) {
  const {
    name,
    domain,
    domains,
    currentStatus,
    statusConfig,
    environmentStyles,
    statusTextColor,
    timeAgo,
    metadataItems,
    displayLabels,
    handleClick
  } = useApplicationItem(application);

  return (
    <Card
      className="relative w-full cursor-pointer overflow-hidden transition-all duration-200 hover:shadow-lg hover:border-primary/30 group"
      onClick={handleClick}
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
              <DomainDropdown domains={domains} variant="icon" />
            </div>

            <div className="flex flex-wrap items-center gap-2 mt-2">
              {domains && domains.length > 0 && (
                <span className="text-xs text-muted-foreground font-mono bg-muted px-2 py-0.5 rounded truncate max-w-[180px]">
                  {domains[0].domain}
                </span>
              )}
              <Badge variant="outline" className={cn('text-xs capitalize', environmentStyles)}>
                {application.environment}
              </Badge>
              {displayLabels && (
                <>
                  {displayLabels.visible.map((label, index) => (
                    <Badge
                      key={index}
                      variant="outline"
                      className="text-xs border-violet-500/30 text-violet-500 bg-violet-500/10"
                    >
                      {label}
                    </Badge>
                  ))}
                  {displayLabels.remainingCount > 0 && (
                    <Badge
                      variant="outline"
                      className="text-xs border-muted-foreground/30 text-muted-foreground bg-muted"
                    >
                      +{displayLabels.remainingCount}
                    </Badge>
                  )}
                </>
              )}
            </div>

            <div className="flex items-center gap-4 mt-4 text-xs text-muted-foreground">
              {metadataItems.map((item) => {
                const Icon = item.icon;
                return (
                  <span key={item.key} className="flex items-center gap-1">
                    <Icon className="h-3 w-3" />
                    <span className={item.key === 'buildPack' ? 'capitalize' : ''}>
                      {item.label}
                    </span>
                  </span>
                );
              })}
            </div>

            <div className="flex items-center justify-between mt-4 pt-3 border-t border-border/50">
              <span className="text-xs text-muted-foreground">{timeAgo}</span>
              <span className={cn('text-xs font-medium', statusTextColor)}>
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
