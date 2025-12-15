'use client';

import {
  Play,
  StopCircle,
  Trash2,
  Info,
  Terminal,
  RotateCw,
  Layers,
  ScrollText
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { OverviewTab } from './components/OverviewTab';
import { LogsTab } from './components/LogsTab';
import { Terminal as TerminalComponent } from './components/Terminal';
import ContainerDetailsLoading from './components/ContainerDetailsLoading';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { Images } from './components/images';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { isNixopusContainer } from '@/lib/utils';
import PageLayout from '@/components/layout/page-layout';
import useContainerDetails from '../hooks/use-container-details';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

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

  const statusColor =
    container.status === 'running'
      ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20'
      : container.status === 'exited'
        ? 'bg-red-500/10 text-red-500 border-red-500/20'
        : 'bg-amber-500/10 text-amber-500 border-amber-500/20';

  return (
    <ResourceGuard resource="container" action="read" loadingFallback={<ContainerDetailsLoading />}>
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        {/* Header Section */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
          <div className="flex items-center gap-4">
            <div
              className={cn(
                'w-12 h-12 rounded-xl flex items-center justify-center',
                container.status === 'running'
                  ? 'bg-emerald-500/10'
                  : container.status === 'exited'
                    ? 'bg-red-500/10'
                    : 'bg-amber-500/10'
              )}
            >
              <div
                className={cn(
                  'w-3 h-3 rounded-full',
                  container.status === 'running'
                    ? 'bg-emerald-500 animate-pulse'
                    : container.status === 'exited'
                      ? 'bg-red-500'
                      : 'bg-amber-500'
                )}
              />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight">{container.name}</h1>
              <div className="flex items-center gap-2 mt-1">
                <code className="text-xs text-muted-foreground font-mono bg-muted px-2 py-0.5 rounded">
                  {container.id.slice(0, 12)}
                </code>
                <Badge variant="outline" className={cn('text-xs', statusColor)}>
                  {container.status}
                </Badge>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2">
            <ResourceGuard
              resource="container"
              action="update"
              loadingFallback={<Skeleton className="h-9 w-24" />}
            >
              {container.status !== 'running' ? (
                <Button
                  variant="default"
                  size="sm"
                  onClick={() => handleContainerAction('start')}
                  disabled={isLoading || isProtected}
                  className="bg-emerald-600 hover:bg-emerald-700"
                >
                  <Play className="mr-2 h-4 w-4" />
                  {t('containers.start')}
                </Button>
              ) : (
                <>
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
                </>
              )}
            </ResourceGuard>
            <ResourceGuard
              resource="container"
              action="delete"
              loadingFallback={<Skeleton className="h-9 w-20" />}
            >
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleContainerAction('remove')}
                disabled={isLoading || isProtected}
                className="text-red-500 hover:text-red-600 hover:bg-red-500/10 border-red-500/20"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                {t('containers.remove')}
              </Button>
            </ResourceGuard>
          </div>
        </div>

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
              {t('containers.logs')}
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
            <LogsTab container={container} logs={allLogs} onLoadMore={handleLoadMoreLogs} />
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
