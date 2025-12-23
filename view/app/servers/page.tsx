'use client';

import React from 'react';
import {
  Plus,
  Server,
  LayoutGrid,
  Network,
  Search,
  RefreshCw,
  Loader2,
  Trash2,
  Zap
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import PageLayout from '@/components/layout/page-layout';
import { useServers } from './hooks/use-servers';
import { ServerTopology } from './components/server-topology';
import { ServerCard } from './components/server-card';
import { ServerDetailsPanel } from './components/server-details-panel';
import { AddServerDialog } from './components/add-server-dialog';
import { ServersPageSkeleton } from './components/servers-skeleton';
import { ServerRole } from '@/redux/types/servers';

export default function ServersPage() {
  const {
    servers,
    cluster,
    isLoading,
    selectedServer,
    setSelectedServer,
    viewMode,
    setViewMode,
    searchQuery,
    setSearchQuery,
    filteredServers,
    stats,
    handleAddServer,
    handleDeleteServer,
    handleTestConnection,
    handleInitCluster,
    handleJoinCluster,
    handleLeaveCluster,
    isAddDialogOpen,
    setIsAddDialogOpen,
    isDeleteDialogOpen,
    setIsDeleteDialogOpen,
    serverToDelete,
    setServerToDelete,
    isTestingConnection,
    t
  } = useServers();

  const [isRefreshing, setIsRefreshing] = React.useState(false);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await new Promise((resolve) => setTimeout(resolve, 1000));
    setIsRefreshing(false);
  };

  if (isLoading) {
    return <ServersPageSkeleton />;
  }

  return (
    <>
      <PageLayout
        maxWidth="full"
        padding="md"
        spacing="lg"
        className="relative z-10 h-full overflow-hidden flex flex-col"
      >
        <div className="flex-shrink-0 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
              <Server className="h-6 w-6" />
              {t('servers.title')}
            </h1>
            <p className="text-muted-foreground mt-1">{t('servers.description')}</p>
          </div>
          <div className="flex items-center gap-2">
            <Button onClick={handleRefresh} variant="outline" size="sm" disabled={isRefreshing}>
              {isRefreshing ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="mr-2 h-4 w-4" />
              )}
              Refresh
            </Button>
            {!cluster?.is_initialized && servers.length > 0 && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const firstOnlineServer = servers.find((s) => s.status === 'online');
                  if (firstOnlineServer) {
                    handleInitCluster(firstOnlineServer.id);
                  }
                }}
              >
                <Zap className="mr-2 h-4 w-4" />
                {t('servers.cluster.initSwarm')}
              </Button>
            )}
            <Button onClick={() => setIsAddDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              {t('servers.addServer')}
            </Button>
          </div>
        </div>

        {stats.total > 0 && (
          <div className="flex-shrink-0 flex items-center gap-6">
            <StatPill value={stats.total} label={t('servers.stats.total')} />
            <StatPill value={stats.online} label={t('servers.stats.online')} color="emerald" />
            <StatPill value={stats.offline} label={t('servers.stats.offline')} color="zinc" />
            <StatPill value={stats.managers} label={t('servers.stats.managers')} color="amber" />
            <StatPill value={stats.workers} label={t('servers.stats.workers')} color="blue" />
          </div>
        )}

        <div className="flex-shrink-0 flex items-center gap-3">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search servers..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex items-center border rounded-lg p-0.5 ml-auto">
            <button
              onClick={() => setViewMode('topology')}
              className={cn(
                'p-2 rounded-md transition-colors flex items-center gap-1.5',
                viewMode === 'topology' ? 'bg-muted' : 'hover:bg-muted/50'
              )}
            >
              <Network className="h-4 w-4" />
              <span className="text-sm hidden sm:inline">{t('servers.views.topology')}</span>
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={cn(
                'p-2 rounded-md transition-colors flex items-center gap-1.5',
                viewMode === 'list' ? 'bg-muted' : 'hover:bg-muted/50'
              )}
            >
              <LayoutGrid className="h-4 w-4" />
              <span className="text-sm hidden sm:inline">{t('servers.views.list')}</span>
            </button>
          </div>
        </div>

        <div className="flex gap-4 flex-1 min-h-0 overflow-hidden relative">
          <div className="flex-1 min-w-0 overflow-y-auto overflow-x-hidden">
            {filteredServers.length === 0 ? (
              <EmptyState
                hasServers={servers.length > 0}
                onAddServer={() => setIsAddDialogOpen(true)}
                t={t}
              />
            ) : viewMode === 'topology' ? (
              <div className="h-full min-h-[500px]">
                <ServerTopology
                  servers={filteredServers}
                  selectedServer={selectedServer}
                  onServerSelect={setSelectedServer}
                />
              </div>
            ) : (
              <div
                className="grid gap-4"
                style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))' }}
              >
                {filteredServers.map((server) => (
                  <ServerCard
                    key={server.id}
                    server={server}
                    isSelected={selectedServer?.id === server.id}
                    onSelect={setSelectedServer}
                    onEdit={() => {}}
                    onDelete={(server) => {
                      setServerToDelete(server);
                      setIsDeleteDialogOpen(true);
                    }}
                    onTestConnection={handleTestConnection}
                    isTestingConnection={isTestingConnection}
                  />
                ))}
              </div>
            )}
          </div>

          {selectedServer && (
            <>
              <div
                className="fixed inset-0 bg-black/50 z-40 lg:hidden animate-in fade-in-0 duration-200"
                onClick={() => setSelectedServer(null)}
                aria-hidden="true"
              />
              <div className="fixed right-0 top-0 h-screen w-full max-w-sm z-50 lg:absolute lg:h-full lg:top-0 lg:right-0 lg:w-96 lg:max-w-none lg:shadow-xl animate-in slide-in-from-right duration-300">
                <ServerDetailsPanel
                  server={selectedServer}
                  onClose={() => setSelectedServer(null)}
                  onEdit={() => {}}
                  onDelete={(server) => {
                    setServerToDelete(server);
                    setIsDeleteDialogOpen(true);
                  }}
                  onTestConnection={handleTestConnection}
                  onJoinCluster={handleJoinCluster}
                  onLeaveCluster={handleLeaveCluster}
                  onPromote={(id) => handleJoinCluster(id, 'manager' as ServerRole)}
                  onDemote={(id) => handleJoinCluster(id, 'worker' as ServerRole)}
                  isTestingConnection={isTestingConnection}
                  hasCluster={!!cluster?.is_initialized}
                />
              </div>
            </>
          )}
        </div>
      </PageLayout>

      <AddServerDialog
        open={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        onSubmit={handleAddServer}
      />

      <DeleteDialog
        title={t('servers.dialog.delete.title')}
        description={t('servers.dialog.delete.description')}
        onConfirm={() => serverToDelete && handleDeleteServer(serverToDelete.id)}
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        variant="destructive"
        confirmText={t('servers.dialog.delete.confirm')}
        cancelText={t('servers.dialog.delete.cancel')}
        icon={Trash2}
      />
    </>
  );
}

function StatPill({
  value,
  label,
  color
}: {
  value: number;
  label: string;
  color?: 'emerald' | 'zinc' | 'amber' | 'blue';
}) {
  const colorClasses = {
    emerald: 'bg-emerald-500',
    zinc: 'bg-zinc-500',
    amber: 'bg-amber-500',
    blue: 'bg-blue-500'
  };

  return (
    <div className="flex items-center gap-2">
      {color && <span className={cn('w-2 h-2 rounded-full', colorClasses[color])} />}
      <span className="text-xl font-bold">{value}</span>
      <span className="text-sm text-muted-foreground">{label}</span>
    </div>
  );
}

function EmptyState({
  hasServers,
  onAddServer,
  t
}: {
  hasServers: boolean;
  onAddServer: () => void;
  t: ReturnType<typeof useServers>['t'];
}) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
      <div className="relative mb-6">
        <div className="absolute inset-0 bg-primary/10 blur-3xl rounded-full" />
        <Server className="relative h-20 w-20 opacity-30" />
      </div>
      <h3 className="text-xl font-semibold text-foreground mb-2">
        {hasServers ? 'No servers match your search' : t('servers.empty.title')}
      </h3>
      <p className="text-center max-w-md mb-6">
        {hasServers ? 'Try adjusting your search criteria' : t('servers.empty.description')}
      </p>
      {!hasServers && (
        <Button onClick={onAddServer}>
          <Plus className="mr-2 h-4 w-4" />
          {t('servers.addServer')}
        </Button>
      )}
    </div>
  );
}
