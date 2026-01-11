'use client';

import { Trash2, Info, Terminal, Layers, ScrollText } from 'lucide-react';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { isNixopusContainer } from '@/lib/utils';
import { useContainerDetail } from '../../../packages/hooks/containers/use-container-detail';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import ContainerDetailsLoading from '../../../packages/components/container-skeletons';
import { ContainerDetailsHeader } from '../../../packages/components/container';
import {
  OverviewTab,
  LogsTab,
  Terminal as TerminalComponent,
  Images
} from '../../../packages/components/container-sections';
import { ResourceGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';

export default function ContainerDetailsPage() {
  const {
    container,
    isLoading,
    handleContainerAction,
    containerId,
    allLogs,
    t,
    handleLoadMoreLogs,
    handleRefreshLogs,
    handleDeleteConfirm,
    isDeleteDialogOpen,
    setIsDeleteDialogOpen
  } = useContainerDetail();

  if (isLoading || !container) {
    return <ContainerDetailsLoading />;
  }

  const isProtected = isNixopusContainer(container?.name);

  return (
    <ResourceGuard resource="container" action="read" loadingFallback={<ContainerDetailsLoading />}>
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        <ContainerDetailsHeader
          container={container}
          isLoading={isLoading}
          isProtected={isProtected}
          handleContainerAction={handleContainerAction}
          t={t}
        />

        {/* Tabs Section */}
        <Tabs defaultValue="overview" className="w-full">
          <TabsList className="w-full justify-start border-b rounded-none h-auto p-0 bg-transparent gap-2">
            <TabsTrigger
              value="overview"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2"
            >
              <Info className="mr-2 h-4 w-4" />
              {t('containers.overview')}
            </TabsTrigger>
            <TabsTrigger
              value="logs"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2"
            >
              <ScrollText className="mr-2 h-4 w-4" />
              {t('containers.logs.title')}
            </TabsTrigger>
            <TabsTrigger
              value="terminal"
              disabled={container.status !== 'running'}
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2"
            >
              <Terminal className="mr-2 h-4 w-4" />
              {t('terminal.title')}
            </TabsTrigger>
            <TabsTrigger
              value="images"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2"
            >
              <Layers className="mr-2 h-4 w-4" />
              {t('containers.images.title')}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="mt-6">
            <OverviewTab container={container} />
          </TabsContent>

          <TabsContent value="logs" className="mt-6">
            <LogsTab
              container={container}
              logs={allLogs}
              onLoadMore={handleLoadMoreLogs}
              onRefresh={handleRefreshLogs}
            />
          </TabsContent>

          <TabsContent value="terminal" className="mt-6">
            {container.status === 'running' ? (
              <div className="rounded-xl overflow-hidden">
                <TerminalComponent containerId={containerId} />
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                <Terminal className="h-12 w-12 mb-4 opacity-30" />
                <p className="font-medium">Container not running</p>
                <p className="text-sm text-muted-foreground/60 mt-1">
                  Start the container to access the terminal
                </p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="images" className="mt-6">
            <Images containerId={containerId} imagePrefix={(container.image || '') + '*'} />
          </TabsContent>
        </Tabs>

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
      </PageLayout>
    </ResourceGuard>
  );
}
