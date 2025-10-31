import { useState } from 'react';
import { useGetAllServersQuery } from '@/redux/services/settings/serversApi';
import { GetServersRequest, Server } from '@/redux/types/server';

export function useServers() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<Server | null>(null);
  const [queryParams, setQueryParams] = useState<GetServersRequest>({
    page: 1,
    page_size: 10,
    search: '',
    sort_by: 'created_at',
    sort_order: 'desc'
  });

  const { data: serverResponse, isLoading, error } = useGetAllServersQuery(queryParams);

  const handleQueryChange = (newParams: Partial<GetServersRequest>) => {
    setQueryParams(prev => ({ ...prev, ...newParams }));
  };

  const handleEditServer = (server: Server) => {
    setEditingServer(server);
    setEditDialogOpen(true);
  };

  const handleEditDialogClose = () => {
    setEditDialogOpen(false);
    setEditingServer(null);
  };

  return {
    createDialogOpen,
    setCreateDialogOpen,
    editDialogOpen,
    setEditDialogOpen,
    editingServer,
    setEditingServer,
    queryParams,
    setQueryParams,
    serverResponse,
    isLoading,
    error,
    handleQueryChange,
    handleEditServer,
    handleEditDialogClose
  };
}

export default useServers;
