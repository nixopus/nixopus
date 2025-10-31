'use client';
import React from 'react';
import { Server as ServerIcon, Trash2 } from 'lucide-react';
import { Server, Pagination, GetServersRequest } from '@/redux/types/server';
import { useTranslation } from '@/hooks/use-translation';
import ServersTableSkeleton from '@/app/settings/servers/components/servers-table-skeleton';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { DahboardUtilityHeader } from '@/components/layout/dashboard-page-header';
import PaginationWrapper from '@/components/ui/pagination';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { DataTable } from '@/components/ui/data-table';
import useServersTable from '@/app/settings/servers/hooks/use-servers-table';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import CreateServerDialog from '@/app/settings/servers/components/create-server';

interface ServersTableProps {
  servers: Server[];
  pagination?: Pagination;
  isLoading: boolean;
  queryParams: GetServersRequest;
  onQueryChange: (params: Partial<GetServersRequest>) => void;
  onEditServer: (server: Server) => void;
  createDialogOpen: boolean;
  setCreateDialogOpen: (open: boolean) => void;
}

function ServersTable({ servers, pagination, isLoading, queryParams, onQueryChange, onEditServer, createDialogOpen, setCreateDialogOpen }: ServersTableProps) {
  const { t } = useTranslation();
  const {
    columns,
    headerSortConfig,
    sortOptions,
    deleteDialogOpen,
    setDeleteDialogOpen,
    serverToDelete,
    deletingId,
    handleDeleteClick,
    handleDeleteConfirm,
    handleSearchChange,
    handleSortChange,
    handlePageSizeChange,
    handlePageChange
  } = useServersTable({ queryParams, onQueryChange, onEditServer });

  if (isLoading) {
    return <ServersTableSkeleton />;
  }

  return (
    <div className="space-y-4">
      <DahboardUtilityHeader<Server>
        searchTerm={queryParams.search || ''}
        handleSearchChange={handleSearchChange}
        sortConfig={headerSortConfig as any}
        onSortChange={handleSortChange}
        sortOptions={sortOptions}
        label={t('servers.page.title')}
        searchPlaceHolder={t('servers.table.search.placeholder')}
      >
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground whitespace-nowrap">{t('servers.table.pagination.itemsPerPage')}</span>
          <SelectWrapper
            value={queryParams.page_size?.toString() || '10'}
            onValueChange={handlePageSizeChange}
            options={[
              { value: '5', label: '5' },
              { value: '10', label: '10' },
              { value: '20', label: '20' },
              { value: '50', label: '50' }
            ]}
            className="w-[90px]"
          />
                <ResourceGuard resource="server" action="create">
            <CreateServerDialog
              open={createDialogOpen}
              setOpen={setCreateDialogOpen}
            />
          </ResourceGuard>
        </div>
      </DahboardUtilityHeader>

      <DataTable<Server>
        data={servers || []}
        columns={columns}
        loading={false}
        emptyStateComponent={
          <div className="text-center py-12">
            <ServerIcon className="mx-auto h-12 w-12 text-muted-foreground" />
            <h3 className="mt-2 text-sm font-semibold text-foreground">{t('servers.table.empty.noServers.title')}</h3>
            <p className="mt-1 text-sm text-muted-foreground">{t('servers.table.empty.noServers.description')}</p>
          </div>
        }
        showBorder
      />

      {pagination && pagination.total_pages > 1 && (
        <div className="flex justify-center">
          <PaginationWrapper
            currentPage={pagination.current_page}
            totalPages={pagination.total_pages}
            onPageChange={handlePageChange}
          />
        </div>
      )}

      <DeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title={t('servers.delete.dialog.title')}
        description={t('servers.delete.dialog.description', { name: serverToDelete?.name || '' })}
        onConfirm={handleDeleteConfirm}
        confirmText={deletingId ? t('servers.delete.dialog.buttons.deleting') : t('servers.delete.dialog.buttons.delete')}
        isDeleting={!!deletingId}
        variant="destructive"
        icon={Trash2}
      />
    </div>
  );
}

export default ServersTable;
