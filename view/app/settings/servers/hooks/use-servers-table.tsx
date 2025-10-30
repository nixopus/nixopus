'use client';
import React, { useMemo, useState } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { formatDistanceToNow } from 'date-fns';
import { useDeleteServerMutation } from '@/redux/services/settings/serversApi';
import { Server, GetServersRequest } from '@/redux/types/server';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { TableColumn } from '@/components/ui/data-table';
import { Edit, MoreHorizontal, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { SortOption } from '@/components/ui/sort-selector';

export interface UseServersTableArgs {
  queryParams: GetServersRequest;
  onQueryChange: (params: Partial<GetServersRequest>) => void;
  onEditServer: (server: Server) => void;
}

export function useServersTable({ queryParams, onQueryChange, onEditServer }: UseServersTableArgs) {
  const { t } = useTranslation();
  const [deleteServer] = useDeleteServerMutation();
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<Server | null>(null);

  const sortOptions: SortOption<Server>[] = [
    { value: 'name', label: t('servers.table.sort.name_asc'), direction: 'asc' },
    { value: 'name', label: t('servers.table.sort.name_desc'), direction: 'desc' },
    { value: 'host', label: t('servers.table.sort.host_asc'), direction: 'asc' },
    { value: 'host', label: t('servers.table.sort.host_desc'), direction: 'desc' },
    { value: 'port', label: t('servers.table.sort.port_asc'), direction: 'asc' },
    { value: 'port', label: t('servers.table.sort.port_desc'), direction: 'desc' },
    { value: 'username', label: t('servers.table.sort.username_asc'), direction: 'asc' },
    { value: 'username', label: t('servers.table.sort.username_desc'), direction: 'desc' },
    { value: 'created_at', label: t('servers.table.sort.created_newest'), direction: 'desc' },
    { value: 'created_at', label: t('servers.table.sort.created_oldest'), direction: 'asc' }
  ];

  const headerSortConfig = useMemo(
    () => ({
      key: (queryParams.sort_by as keyof Server) ?? ('created_at' as keyof Server),
      direction: (queryParams.sort_order as 'asc' | 'desc') ?? 'desc'
    }),
    [queryParams.sort_by, queryParams.sort_order]
  );

  const handleDeleteClick = (server: Server) => {
    setServerToDelete(server);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!serverToDelete) return;
    try {
      setDeletingId(serverToDelete.id);
      await deleteServer({ id: serverToDelete.id }).unwrap();
      toast.success(t('servers.messages.deleteSuccess'));
      setDeleteDialogOpen(false);
      setServerToDelete(null);
    } catch (error) {
      toast.error(t('servers.messages.deleteError'));
    } finally {
      setDeletingId(null);
    }
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onQueryChange({ search: e.target.value, page: 1 });
  };

  const handleSortChange = (newSortOption: SortOption<Server>) => {
    onQueryChange({ sort_by: newSortOption.value as string, sort_order: newSortOption.direction, page: 1 });
  };

  const handlePageSizeChange = (newPageSize: string) => {
    onQueryChange({ page_size: Number(newPageSize), page: 1 });
  };

  const handlePageChange = (newPage: number) => {
    onQueryChange({ page: newPage });
  };

  const columns: TableColumn<Server>[] = useMemo(
    () => [
      {
        key: 'name',
        title: t('servers.table.headers.name'),
        render: (_, record) => (
          <div>
            <div className="font-medium">{record.name}</div>
            {record.description ? (
              <div className="text-sm text-muted-foreground truncate max-w-[200px]">{record.description}</div>
            ) : null}
          </div>
        )
      },
      {
        key: 'host',
        title: t('servers.table.headers.host'),
        render: (_, record) => <code className="text-sm bg-muted px-2 py-1 rounded">{record.host}</code>
      },
      {
        key: 'port',
        title: t('servers.table.headers.port'),
        render: (_, record) => <Badge variant="outline">{record.port}</Badge>
      },
      {
        key: 'username',
        title: t('servers.table.headers.username'),
        render: (_, record) => <code className="text-sm bg-muted px-2 py-1 rounded">{record.username}</code>
      },
      {
        key: 'created_at',
        title: t('servers.table.headers.created'),
        render: (_, record) => (
          <span className="text-sm text-muted-foreground">
            {formatDistanceToNow(new Date(record.created_at), { addSuffix: true })}
          </span>
        )
      },
      {
        key: 'actions',
        title: '',
        width: '70px',
        align: 'right',
        render: (_, record) => (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem className="cursor-pointer" onClick={() => onEditServer(record)}>
                <Edit className="mr-2 h-4 w-4" />
                {t('servers.actions.edit')}
              </DropdownMenuItem>
              <DropdownMenuItem
                className="cursor-pointer text-destructive"
                onClick={() => handleDeleteClick(record)}
                disabled={deletingId === record.id}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                {deletingId === record.id ? t('servers.actions.deleting') : t('servers.actions.delete')}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )
      }
    ],
    [t, deletingId, onEditServer]
  );

  return {
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
  };
}

export type UseServersTableReturn = ReturnType<typeof useServersTable>;

export default useServersTable;
