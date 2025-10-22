'use client';

import {
  Play,
  StopCircle,
  Trash2,
  Info,
  Terminal,
  HardDrive,
  RotateCw,
  Layers
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { OverviewTab } from './components/OverviewTab';
import { LogsTab } from './components/LogsTab';
import { DetailsTab } from './components/DetailsTab';
import { Terminal as TerminalComponent } from './components/Terminal';
import ContainerDetailsLoading from './components/ContainerDetailsLoading';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { Images } from './components/images';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { isNixopusContainer } from '@/lib/utils';
import PageLayout from '@/components/layout/page-layout';
import useContainerDetails from '../hooks/use-container-details';

export default function ContainerDetailsPage() {
  const {
    container,
    isLoading,
    handleContainerAction,
    containerId,
    allLogs,
    t,
    handleLoadMoreLogs,
    handleDeleteConfirm,
    isDeleteDialogOpen,
    setIsDeleteDialogOpen
  } = useContainerDetails();

  if (isLoading || !container) {
    return <ContainerDetailsLoading />;
  }

  const isProtected = isNixopusContainer(container?.name);

  return (
    <ResourceGuard resource="container" action="read" loadingFallback={<ContainerDetailsLoading />}>
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        <div className="flex items-center justify-between mb-6 pb-4 border-b">
          <div>
            <h1 className="text-2xl font-bold">{container.name}</h1>
            <p className="text-muted-foreground">{container.id.slice(0, 12)}...</p>
          </div>
          <div className="flex items-center gap-2">
            <ResourceGuard
              resource="container"
              action="update"
              loadingFallback={<Skeleton className="h-8 w-16" />}
            >
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleContainerAction('start')}
                disabled={isLoading || isProtected}
              >
                <Play className="mr-2 h-4 w-4" />
                {t('containers.start')}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleContainerAction('stop')}
                disabled={isLoading || isProtected}
              >
                <StopCircle className="mr-2 h-4 w-4" />
                {t('containers.stop')}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleContainerAction('restart')}
                disabled={isLoading || isProtected}
              >
                <RotateCw className="mr-2 h-4 w-4" />
                {t('containers.restart')}
              </Button>
            </ResourceGuard>
            <ResourceGuard
              resource="container"
              action="delete"
              loadingFallback={<Skeleton className="h-8 w-20" />}
            >
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleContainerAction('remove')}
                disabled={isLoading || isProtected}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                {t('containers.remove')}
              </Button>
            </ResourceGuard>
          </div>
        </div>

        <div className="space-y-4">
          <Tabs defaultValue="overview" className="w-full">
            <TabsList className="grid w-full grid-cols-5">
              <TabsTrigger value="overview">
                <Info className="mr-2 h-4 w-4" />
                {t('containers.overview')}
              </TabsTrigger>
              <TabsTrigger value="images">
                <Layers className="mr-2 h-4 w-4" />
                {t('containers.images.title')}
              </TabsTrigger>
              <TabsTrigger value="terminal" disabled={container.status !== 'running'}>
                <Terminal className="mr-2 h-4 w-4" />
                {t('terminal.title')}
              </TabsTrigger>
              <TabsTrigger value="logs">
                <Terminal className="mr-2 h-4 w-4" />
                {t('containers.logs')}
              </TabsTrigger>
              <TabsTrigger value="details">
                <HardDrive className="mr-2 h-4 w-4" />
                {t('containers.details')}
              </TabsTrigger>
            </TabsList>
            <TabsContent value="overview" className="mt-4">
              <OverviewTab container={container} />
            </TabsContent>
            <TabsContent value="logs" className="mt-4">
              <LogsTab container={container} logs={allLogs} onLoadMore={handleLoadMoreLogs} />
            </TabsContent>
            <TabsContent value="details" className="mt-4">
              <DetailsTab container={container} />
            </TabsContent>
            <TabsContent value="terminal" className="mt-4">
              {container.status === 'running' ? (
                <TerminalComponent containerId={containerId} />
              ) : (
                <div className="flex items-center justify-center h-48 text-muted-foreground">
                  Start the container to use the terminal
                </div>
              )}
            </TabsContent>
            <TabsContent value="images" className="mt-4">
              {container.image ? (
                <Images containerId={containerId} imagePrefix={container.image + '*'} />
              ) : (
                <div className="flex items-center justify-center h-full">
                  <p>{t('containers.images.none')}</p>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </div>
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
