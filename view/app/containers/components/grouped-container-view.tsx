'use client';

import React, { useState } from 'react';
import { ChevronDown, ChevronRight, Box } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Container } from '@/redux/services/container/containerApi';
import { ContainerCard } from './card';
import ContainersTable from './table';

interface GroupedContainerViewProps {
  groups: Array<{ application_id: string; application_name: string; containers: Container[] }>;
  ungrouped?: Container[];
  viewMode: 'table' | 'card';
  onContainerClick: (container: Container) => void;
  onContainerAction: (id: string, action: 'start' | 'stop' | 'remove') => void;
  getGradientFromName: (name: string) => string;
  sortBy: 'name' | 'status';
  sortOrder: 'asc' | 'desc';
  onSort: (field: 'name' | 'status') => void;
}

export function GroupedContainerView({
  groups,
  ungrouped = [],
  viewMode,
  onContainerClick,
  onContainerAction,
  getGradientFromName,
  sortBy,
  sortOrder,
  onSort
}: GroupedContainerViewProps) {
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(
    new Set(groups.map((g) => g.application_id))
  );

  const toggleGroup = (applicationId: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(applicationId)) {
        next.delete(applicationId);
      } else {
        next.add(applicationId);
      }
      return next;
    });
  };

  if (groups.length === 0 && ungrouped.length === 0) {
    return null;
  }

  return (
    <div className="space-y-6">
      {groups.map((group) => {
        const isExpanded = expandedGroups.has(group.application_id);
        const runningCount = group.containers.filter((c) => c.status === 'running').length;
        const totalCount = group.containers.length;

        return (
          <div key={group.application_id} className="border rounded-lg overflow-hidden">
            <button
              onClick={() => toggleGroup(group.application_id)}
              className={cn(
                'w-full flex items-center justify-between p-4 text-left',
                'hover:bg-muted/50 transition-colors',
                'border-b border-border'
              )}
            >
              <div className="flex items-center gap-3 flex-1 min-w-0">
                {isExpanded ? (
                  <ChevronDown className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                ) : (
                  <ChevronRight className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                )}
                <Box className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                <div className="min-w-0 flex-1">
                  <h3 className="font-semibold truncate">{group.application_name}</h3>
                  <p className="text-sm text-muted-foreground">
                    {totalCount} container{totalCount !== 1 ? 's' : ''} • {runningCount} running
                  </p>
                </div>
              </div>
            </button>

            {isExpanded && (
              <div className="">
                {viewMode === 'card' ? (
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2 p-2">
                    {group.containers.map((container) => (
                      <ContainerCard
                        key={container.id}
                        container={container}
                        onClick={() => onContainerClick(container)}
                        getGradientFromName={getGradientFromName}
                        onAction={onContainerAction}
                      />
                    ))}
                  </div>
                ) : (
                  <ContainersTable
                    containersData={group.containers}
                    sortBy={sortBy}
                    sortOrder={sortOrder}
                    onSort={onSort}
                    onAction={onContainerAction}
                  />
                )}
              </div>
            )}
          </div>
        );
      })}

      {ungrouped.length > 0 && (
        <>
          {viewMode === 'card' ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
              {ungrouped.map((container) => (
                <ContainerCard
                  key={container.id}
                  container={container}
                  onClick={() => onContainerClick(container)}
                  getGradientFromName={getGradientFromName}
                  onAction={onContainerAction}
                />
              ))}
            </div>
          ) : (
            <ContainersTable
              containersData={ungrouped}
              sortBy={sortBy}
              sortOrder={sortOrder}
              onSort={onSort}
              onAction={onContainerAction}
              className="border rounded-lg"
            />
          )}
        </>
      )}
    </div>
  );
}
