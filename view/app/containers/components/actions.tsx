'use client';

import { Play, Square, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { isNixopusContainer } from '@/lib/utils';
import { Skeleton } from '@/components/ui/skeleton';
import { ResourceGuard } from '@/packages/components/rbac';
import { Action } from './card';
import { cn } from '@/lib/utils';

interface ContainerActionsProps {
  container: any;
  onAction: (id: string, action: Action) => void;
}

export const ContainerActions = ({ container, onAction }: ContainerActionsProps) => {
  const containerName: string = typeof container?.name === 'string' ? container.name : '';
  const isProtected = isNixopusContainer(containerName);
  const isRunning =
    (container.state || '').toLowerCase() === 'running' ||
    (container.status || '').toLowerCase() === 'running';
  const containerId = container.id;

  function handleClick(e: React.MouseEvent, action: Action) {
    e.stopPropagation();
    onAction(containerId, action);
  }

  return (
    <div className="flex items-center gap-1">
      <ResourceGuard
        resource="container"
        action="update"
        loadingFallback={<Skeleton className="h-8 w-8 rounded-lg" />}
      >
        {isRunning ? (
          <ActionButton
            icon={Square}
            onClick={(e) => handleClick(e, Action.STOP)}
            disabled={isProtected}
            tooltip="Stop container"
            variant="warning"
          />
        ) : (
          <ActionButton
            icon={Play}
            onClick={(e) => handleClick(e, Action.START)}
            disabled={isProtected}
            tooltip="Start container"
            variant="success"
          />
        )}
      </ResourceGuard>
      <ResourceGuard
        resource="container"
        action="delete"
        loadingFallback={<Skeleton className="h-8 w-8 rounded-lg" />}
      >
        <ActionButton
          icon={Trash2}
          onClick={(e) => handleClick(e, Action.REMOVE)}
          disabled={isProtected}
          tooltip="Remove container"
          variant="danger"
        />
      </ResourceGuard>
    </div>
  );
};

function ActionButton({
  icon: Icon,
  onClick,
  disabled,
  tooltip,
  variant
}: {
  icon: React.ElementType;
  onClick: (e: React.MouseEvent) => void;
  disabled?: boolean;
  tooltip?: string;
  variant?: 'success' | 'warning' | 'danger';
}) {
  const variantStyles = {
    success: 'hover:bg-emerald-500/10 hover:text-emerald-500',
    warning: 'hover:bg-amber-500/10 hover:text-amber-500',
    danger: 'hover:bg-red-500/10 hover:text-red-500'
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'h-8 w-8 text-muted-foreground transition-colors',
        variant && variantStyles[variant],
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      title={tooltip}
    >
      <Icon className="h-4 w-4" />
    </Button>
  );
}
