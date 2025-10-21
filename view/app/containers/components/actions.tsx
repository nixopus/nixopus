'use client';

import { Play, Pause, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { isNixopusContainer } from '@/lib/utils';
import { Skeleton } from '@/components/ui/skeleton';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Action } from './card';

const resourceName = 'container';
const ghostVariant = 'ghost';
const iconSize = 'icon';
const iconStyle = 'h-4 w-4';
const containerStatusRunning = 'running';

enum ResourceActions {
  UPDATE = 'update',
  DELETE = 'delete'
}

interface ContainerActionsProps {
  container: any;
  onAction: (id: string, action: Action) => void;
}

export const ContainerActions = ({ container, onAction }: ContainerActionsProps) => {
  const containerName: string = typeof container?.name === 'string' ? container.name : '';
  const isProtected = isNixopusContainer(containerName);
  const isContainerRunning =
    (container.state || '').toLowerCase() === containerStatusRunning ||
    (container.status || '').toLowerCase() === containerStatusRunning;
  const containerId = container.id;

  function onClickHandler(e: any, action: Action) {
    e.stopPropagation();
    onAction(containerId, action);
  }

  return (
    <div className="flex gap-2">
      <ResourceGuard
        resource={resourceName}
        action={ResourceActions.UPDATE}
        loadingFallback={<LoadingFallback />}
      >
        <ActionIconsRenderer
          isContainerRunning={isContainerRunning}
          isProtected={isProtected}
          onClickHandler={onClickHandler}
        />
      </ResourceGuard>
      <ResourceGuard
        resource={resourceName}
        action={ResourceActions.DELETE}
        loadingFallback={<LoadingFallback />}
      >
        <Button
          variant={ghostVariant}
          size={iconSize}
          disabled={isProtected}
          onClick={(e) => onClickHandler(e, Action.REMOVE)}
        >
          <Trash2 className={iconStyle} />
        </Button>
      </ResourceGuard>
    </div>
  );
};

function ActionIconsRenderer({
  onClickHandler,
  isContainerRunning,
  isProtected
}: {
  onClickHandler(e: any, action: Action): void;
  isContainerRunning: boolean;
  isProtected: boolean;
}) {
  // toggle: show Play when not running, Pause when running
  if (isContainerRunning) {
    return (
      <Button
        variant={ghostVariant}
        size={iconSize}
        disabled={isProtected}
        onClick={(e) => onClickHandler(e, Action.STOP)}
      >
        <Pause className={iconStyle} />
      </Button>
    );
  }
  return (
    <Button
      variant={ghostVariant}
      size={iconSize}
      disabled={isProtected}
      onClick={(e) => onClickHandler(e, Action.START)}
    >
      <Play className={iconStyle} />
    </Button>
  );
}

function LoadingFallback() {
  return <Skeleton className="h-8 w-8" />;
}
