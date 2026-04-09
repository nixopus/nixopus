import React from 'react';
import { Badge, CardWrapper, CardTitle } from '@nixopus/ui';
import { Application } from '@/redux/types/applications';
import { cn } from '@/lib/utils';
import { useApplicationItem } from '@/packages/hooks/applications/use_application_item';
import { MoreVertical, MoveUpRight } from 'lucide-react';
import { CARD_CLASS, CARD_HEADER_CLASS, CardSkeleton } from '@/components/ui/list-page-chrome';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@nixopus/ui';
import { Button } from '@nixopus/ui';

const STATUS_BADGE: Record<string, string> = {
  deployed: 'bg-foreground/5 text-foreground/70 border-foreground/10',
  running: 'bg-foreground/5 text-foreground/70 border-foreground/10',
  failed: 'bg-foreground/5 text-red-400 border-foreground/10',
  cancelled: 'bg-foreground/5 text-foreground/50 border-foreground/10',
  building: 'bg-foreground/5 text-orange-400 border-foreground/10',
  deploying: 'bg-foreground/5 text-orange-400 border-foreground/10',
  cloning: 'bg-foreground/5 text-orange-400 border-foreground/10',
  started: 'bg-foreground/5 text-orange-400 border-foreground/10',
  draft: 'bg-foreground/5 text-foreground/50 border-foreground/10',
  stopped: 'bg-foreground/5 text-foreground/50 border-foreground/10'
};

const STATUS_LABEL: Record<string, string> = {
  deployed: 'Live',
  running: 'Running',
  failed: 'Failed',
  cancelled: 'Cancelled',
  building: 'Building',
  deploying: 'Building',
  cloning: 'Building',
  started: 'Building',
  draft: 'Draft',
  stopped: 'Stopped'
};

function AppItem(application: Application) {
  const { name, domains, currentStatus, environmentStyles, timeAgo, handleClick } =
    useApplicationItem(application);

  const domainActions = (
    <>
      {domains && domains.length === 1 && (
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7 shrink-0"
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
              className="h-7 w-7 shrink-0"
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
    </>
  );

  return (
    <CardWrapper
      className={CARD_CLASS}
      onClick={handleClick}
      header={
        <div className="flex-1 min-w-0 w-full space-y-1.5">
          <div className="flex w-full min-w-0 items-start justify-between gap-2">
            <div className="min-w-0 flex-1 max-w-full">
              <CardTitle
                className="text-base font-semibold wrap-break-word max-w-full"
                style={{ wordBreak: 'break-word', overflowWrap: 'break-word', maxWidth: '100%' }}
              >
                {name}
              </CardTitle>
            </div>
            <div className="flex items-center gap-1 ml-auto shrink-0">{domainActions}</div>
          </div>
          <div className="flex items-center gap-2">
            <Badge
              variant="secondary"
              className={cn(
                'uppercase text-[10px] rounded-sm border px-1.5 py-0',
                STATUS_BADGE[currentStatus] || 'bg-foreground/5 text-muted-foreground border-border'
              )}
            >
              {STATUS_LABEL[currentStatus] || 'Inactive'}
            </Badge>
            <Badge
              variant="outline"
              className={cn(
                'text-xs font-medium capitalize rounded-sm px-2 py-0.5 border',
                environmentStyles
              )}
            >
              {application.environment}
            </Badge>
          </div>
        </div>
      }
      headerClassName={CARD_HEADER_CLASS}
      contentClassName="space-y-1.5"
    >
      {timeAgo && (
        <div className="flex items-center gap-3 pt-1.5 border-t border-border/50">
          <span className="text-xs text-muted-foreground">Deployed {timeAgo}</span>
        </div>
      )}
    </CardWrapper>
  );
}

export default AppItem;

export { CardSkeleton as AppItemSkeleton };
