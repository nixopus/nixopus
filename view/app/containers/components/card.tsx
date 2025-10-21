'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { ContainerActions } from './actions';
import { ContainerInfo } from './info';

interface ContainerCardProps {
  container: any;
  onClick: () => void;
  getGradientFromName: (name: string) => string;
  onAction: (id: string, action: Action) => void;
}

export enum Action {
  START = 'start',
  STOP = 'stop',
  REMOVE = 'remove'
}

export const ContainerCard = ({
  container,
  onClick,
  getGradientFromName,
  onAction
}: ContainerCardProps) => {
  return (
    <Card
      className={cn(
        'group relative overflow-hidden transition-all duration-300 hover:shadow-lg cursor-pointer h-full flex flex-col',
        getGradientFromName(container.name)
      )}
      onClick={onClick}
    >
      <CardContent className="relative p-6 z-10 flex-1 flex flex-col">
        <div className="flex items-start justify-between mb-4">
          <div className="space-y-2 flex-1 min-w-0">
            <h3 className="text-xl font-semibold truncate">{container.name}</h3>
            <p className="text-sm text-muted-foreground truncate" title={container.image}>
              {container.image}
            </p>
            <Badge variant={container.status === 'running' ? 'default' : 'secondary'}>
              {container.status}
            </Badge>
          </div>
          <div className="flex-shrink-0 ml-4">
            <ContainerActions container={container} onAction={onAction} />
          </div>
        </div>
        <div className="mt-auto">
          <ContainerInfo container={container} />
        </div>
      </CardContent>
    </Card>
  );
};
