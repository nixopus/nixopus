'use client';
import React from 'react';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import CreateServerDialog from '@/app/settings/servers/components/create-server';
import ServersTable from '@/app/settings/servers/components/servers-table';
import useServers from '@/app/settings/servers/hooks/use-servers';

function Page() {
  const { t } = useTranslation();
  const {
    createDialogOpen,
    setCreateDialogOpen,
    editDialogOpen,
    setEditDialogOpen,
    editingServer,
    setEditingServer,
    queryParams,
    serverResponse,
    isLoading,
    error,
    handleQueryChange,
    handleEditServer,
    handleEditDialogClose
  } = useServers();

  return (
    <>
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader
        label="Server Settings"
        description="Manage your servers and their configurations"
      />
    </div>
    <ResourceGuard
      resource="server"
      action="read"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <div className="container mx-auto py-6 space-y-8 max-w-6xl">
        <div className="flex justify-between items-center">
          <DashboardPageHeader
            label={t('servers.page.title')}
            description={t('servers.page.description')}
          />
          <ResourceGuard resource="server" action="create">
            <CreateServerDialog
              open={createDialogOpen}
              setOpen={setCreateDialogOpen}
            />
          </ResourceGuard>
        </div>

        <div className="space-y-6">
          {error ? (
            <div className="text-center py-12">
              <p className="text-destructive">{t('servers.page.error')}</p>
            </div>
          ) : (
            <ServersTable
              servers={serverResponse?.servers || []}
              pagination={serverResponse?.pagination}
              isLoading={isLoading}
              queryParams={queryParams}
              onQueryChange={handleQueryChange}
              onEditServer={handleEditServer}
            />
          )}
        </div>

        <ResourceGuard resource="server" action="update">
          <CreateServerDialog
            open={editDialogOpen}
            setOpen={handleEditDialogClose}
            mode="edit"
            serverId={editingServer?.id}
            serverData={editingServer || undefined}
          />
          <CreateServerDialog
            open={editDialogOpen}
            setOpen={handleEditDialogClose}
            mode="edit"
            serverId={editingServer?.id}
            serverData={editingServer || undefined}
          />
          <CreateServerDialog
            open={editDialogOpen}
            setOpen={(open: boolean) => {
              setEditDialogOpen(open);
              if (!open) setEditingServer(null);
            }}
            mode="edit"
            serverId={editingServer?.id}
            serverData={editingServer || undefined}
          />
        </ResourceGuard>
      </div>
    </ResourceGuard>
    </>
  );
}

export default Page;