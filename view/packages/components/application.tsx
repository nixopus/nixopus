import React from 'react';
import { Badge } from '@nixopus/ui';
import { Card, CardContent } from '@nixopus/ui';
import { Application } from '@/redux/types/applications';
import { Skeleton } from '@nixopus/ui';
import { cn } from '@/lib/utils';
import { useApplicationItem } from '@/packages/hooks/applications/use_application_item';
import { MoreVertical, MoveUpRight } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@nixopus/ui';
import { Button } from '@nixopus/ui';

function AppItem(application: Application) {
  const { name, domains, currentStatus, statusConfig, environmentStyles, timeAgo, handleClick } =
    useApplicationItem(application);

  return (
    <Card
      className="relative w-full max-w-xs cursor-pointer overflow-hidden border border-white/[0.06] transition-colors duration-200 hover:bg-muted/50"
      onClick={handleClick}
    >
      <CardContent className="px-6 pb-6">
        <div className="flex flex-col gap-2">
          <div className="flex items-start justify-between">
            <h3 className="font-semibold text-base tracking-tight truncate">{name}</h3>
            {domains && domains.length === 1 && (
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 shrink-0 text-muted-foreground hover:text-foreground"
                onClick={(e) => {
                  e.stopPropagation();
                  const domain = domains[0].domain;
                  const url = domain.startsWith('http') ? domain : `https://${domain}`;
                  window.open(url, '_blank', 'noopener,noreferrer');
                }}
              >
                <MoveUpRight className="h-4 w-4" />
              </Button>
            )}
            {domains && domains.length > 1 && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 shrink-0 text-muted-foreground hover:text-foreground"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" onClick={(e) => e.stopPropagation()}>
                  {domains.map((domainItem, index) => (
                    <DropdownMenuItem
                      key={index}
                      onClick={(e) => {
                        e.stopPropagation();
                        const url = domainItem.domain.startsWith('http')
                          ? domainItem.domain
                          : `https://${domainItem.domain}`;
                        window.open(url, '_blank', 'noopener,noreferrer');
                      }}
                    >
                      {domainItem.domain}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>

          {timeAgo && <span className="text-xs text-muted-foreground">Deployed {timeAgo}</span>}

          <div className="flex items-center gap-2 mt-1">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <span
                className={cn(
                  'inline-block h-2 w-2 rounded-full',
                  currentStatus === 'deployed' || currentStatus === 'running'
                    ? 'bg-emerald-500'
                    : currentStatus === 'failed'
                      ? 'bg-red-500'
                      : currentStatus === 'building' ||
                          currentStatus === 'deploying' ||
                          currentStatus === 'cloning' ||
                          currentStatus === 'started'
                        ? 'bg-amber-500'
                        : 'bg-zinc-500'
                )}
              />
              {statusConfig.label}
            </div>
            <Badge
              variant="outline"
              className={cn(
                'text-xs font-medium capitalize rounded-full px-3 py-0.5 border',
                environmentStyles
              )}
            >
              {application.environment}
            </Badge>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default AppItem;

export function AppItemSkeleton() {
  return (
    <Card className="relative w-full max-w-sm">
      <CardContent className="px-6 pb-6">
        <div className="flex flex-col gap-2">
          <div className="flex items-start justify-between">
            <Skeleton className="h-6 w-36" />
            <Skeleton className="h-7 w-7 rounded-md" />
          </div>
          <Skeleton className="h-4 w-40 mt-0.5" />
          <div className="flex items-center gap-2 mt-2">
            <Skeleton className="h-6 w-18 rounded-full" />
            <Skeleton className="h-6 w-14 rounded-full" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
