'use client';

import React, { useState } from 'react';
import { Box, RefreshCw, Loader2, LayoutGrid, List, ArrowLeft, Trash2 } from 'lucide-react';
import { Button, Skeleton } from '@nixopus/ui';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { cn, isNixopusContainer } from '@/lib/utils';
import { useApplicationContainers } from '@/packages/hooks/containers/use-application-containers';
import { useInlineContainerDetail } from '@/packages/hooks/containers/use-inline-container-detail';
import { useViewMode } from '@/packages/hooks/containers/use-view-mode';
import { Container } from '@/redux/services/container/containerApi';
import { ContainerCard, EmptyState, ContainerDetailsHeader } from '@/packages/components/container';
import ContainersTable from '@/packages/components/container';
import {
  OverviewTab,
  Terminal as TerminalComponent,
  Images
} from '@/packages/components/container-sections';
import ContainerDetailsLoading from '@/packages/components/container-skeletons';
import { ResourceGuard } from '@/packages/components/rbac';

interface ApplicationResourcesProps {
  applicationId: string;
}

function ApplicationResourcesSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <Skeleton className="h-9 w-24 rounded-md" />
        <div className="flex items-center gap-1 p-1 bg-muted/50 rounded-lg">
          <Skeleton className="h-8 w-8 rounded-md" />
          <Skeleton className="h-8 w-8 rounded-md" />
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="rounded-xl p-5 bg-muted/50 border border-transparent">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 space-y-2">
                <Skeleton className="h-5 w-32" />
                <Skeleton className="h-4 w-24 rounded-full" />
              </div>
              <Skeleton className="h-8 w-8 rounded-md" />
            </div>
            <div className="mt-4">
              <Skeleton className="h-4 w-40" />
            </div>
            <div className="mt-4 flex items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <Skeleton className="h-3.5 w-3.5 rounded-full" />
                <Skeleton className="h-5 w-16 rounded-full" />
              </div>
              <div className="flex items-center gap-2">
                <Skeleton className="h-3 w-3 rounded-full" />
                <Skeleton className="h-3 w-20" />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function InlineContainerDetail({
  containerId,
  onBack
}: {
  containerId: string;
  onBack: () => void;
}) {
  const {
    container,
    isLoading,
    isDeleteDialogOpen,
    setIsDeleteDialogOpen,
    handleContainerAction,
    handleDeleteConfirm,
    t
  } = useInlineContainerDetail(containerId, onBack);
  const [terminalOpen, setTerminalOpen] = useState(false);

  if (isLoading || !container) {
    return <ContainerDetailsLoading />;
  }

  const isProtected = isNixopusContainer(container?.name);

  return (
    <div className="space-y-6">
      <Button variant="ghost" size="sm" onClick={onBack} className="gap-2 -ml-2">
        <ArrowLeft className="h-4 w-4" />
        {t('containers.backToContainers') || 'Back to containers'}
      </Button>

      <ContainerDetailsHeader
        container={container}
        isLoading={isLoading}
        isProtected={isProtected}
        handleContainerAction={handleContainerAction}
        t={t}
        onExecute={() => setTerminalOpen(!terminalOpen)}
        terminalOpen={terminalOpen}
      />

      {terminalOpen ? (
        <TerminalComponent containerId={containerId} />
      ) : (
        <>
          <OverviewTab container={container} />
          <Images containerId={containerId} imagePrefix={(container.image || '') + '*'} />
        </>
      )}

      <ResourceGuard resource="container" action="delete" loadingFallback={null}>
        <DeleteDialog
          title={t('containers.deleteDialog.title')}
          description={t('containers.deleteDialog.description')}
          onConfirm={handleDeleteConfirm}
          open={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          variant="destructive"
          confirmText={t('containers.deleteDialog.confirm')}
          cancelText={t('containers.deleteDialog.cancel')}
          icon={Trash2}
        />
      </ResourceGuard>
    </div>
  );
}

export function ApplicationResources({ applicationId }: ApplicationResourcesProps) {
  const { viewMode, setViewMode } = useViewMode();
  const [selectedContainerId, setSelectedContainerId] = useState<string | null>(null);

  const {
    containers,
    isLoading,
    isFetching,
    handleRefresh,
    handleContainerAction,
    handleDeleteConfirm,
    isRefreshing,
    t,
    containerToDelete,
    setContainerToDelete,
    totalCount
  } = useApplicationContainers(applicationId);

  const loading = isRefreshing || isFetching;

  if (isLoading) {
    return <ApplicationResourcesSkeleton />;
  }

  if (selectedContainerId) {
    return (
      <InlineContainerDetail
        containerId={selectedContainerId}
        onBack={() => setSelectedContainerId(null)}
      />
    );
  }

  if (totalCount === 0) {
    return (
      <EmptyState
        icon={Box}
        message={t('containers.noContainersForApplication')}
        className="py-12"
      />
    );
  }

  const handleContainerClick = (container: Container) => {
    setSelectedContainerId(container.id);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <Button onClick={handleRefresh} variant="outline" size="sm" disabled={loading}>
            {loading ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="mr-2 h-4 w-4" />
            )}
            {t('containers.refresh')}
          </Button>
        </div>

        <div className="flex items-center gap-1 p-1 bg-muted/50 rounded-lg">
          <Button
            variant={viewMode === 'table' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('table')}
            className={cn('h-8 px-3', viewMode === 'table' && 'shadow-sm')}
          >
            <List className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'card' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('card')}
            className={cn('h-8 px-3', viewMode === 'card' && 'shadow-sm')}
          >
            <LayoutGrid className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {viewMode === 'card' ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {containers.map((container) => (
            <ContainerCard
              key={container.id}
              container={container}
              onClick={() => handleContainerClick(container)}
              onAction={handleContainerAction}
            />
          ))}
        </div>
      ) : (
        <ContainersTable
          containersData={containers}
          sortBy="name"
          sortOrder="asc"
          onSort={() => {}}
          onAction={handleContainerAction}
          onRowClick={handleContainerClick}
        />
      )}

      <DeleteDialog
        open={!!containerToDelete}
        onOpenChange={(open) => !open && setContainerToDelete(null)}
        onConfirm={handleDeleteConfirm}
        title={t('containers.delete_title')}
        description={t('containers.delete_description')}
      />
    </div>
  );
}
