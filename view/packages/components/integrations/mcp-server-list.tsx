'use client';

import React, { useState } from 'react';
import { Button, Switch } from '@nixopus/ui';
import { Pencil, Trash2 } from 'lucide-react';
import type { MCPServer, MCPProvider } from '@/redux/types/mcp';
import { MCPProviderPicker } from './mcp-provider-picker';
import { MCPCredentialForm } from './mcp-credential-form';
import {
  useGetMCPCatalogQuery,
  useAddMCPServerMutation,
  useUpdateMCPServerMutation,
  useDeleteMCPServerMutation
} from '@/redux/services/settings/mcpApi';

interface Props {
  servers: MCPServer[];
  canConfigure: boolean;
  canDelete: boolean;
}

type Step = 'list' | 'pick-provider' | 'fill-credentials';

export function MCPServerList({ servers, canConfigure, canDelete }: Props) {
  const [step, setStep] = useState<Step>('list');
  const [selectedProvider, setSelectedProvider] = useState<MCPProvider | null>(null);
  const [editingServer, setEditingServer] = useState<MCPServer | null>(null);

  const { data: catalogResult } = useGetMCPCatalogQuery({});
  const catalog = catalogResult?.items ?? [];
  const [addServer, { isLoading: isAdding }] = useAddMCPServerMutation();
  const [updateServer, { isLoading: isUpdating }] = useUpdateMCPServerMutation();
  const [deleteServer] = useDeleteMCPServerMutation();

  const handleProviderSelect = (provider: MCPProvider) => {
    setSelectedProvider(provider);
    setStep('fill-credentials');
  };

  const handleSave = async (data: {
    name: string;
    credentials: Record<string, string>;
    custom_url?: string;
    enabled: boolean;
  }) => {
    if (editingServer) {
      await updateServer({ id: editingServer.id, ...data }).unwrap();
    } else if (selectedProvider) {
      await addServer({ provider_id: selectedProvider.id, ...data }).unwrap();
    }
    setStep('list');
    setSelectedProvider(null);
    setEditingServer(null);
  };

  const handleEdit = (server: MCPServer) => {
    const provider = catalog.find((p) => p.id === server.provider_id);
    if (provider) {
      setSelectedProvider(provider);
      setEditingServer(server);
      setStep('fill-credentials');
    }
  };

  const handleToggleEnabled = async (server: MCPServer) => {
    await updateServer({
      id: server.id,
      name: server.name,
      credentials: server.credentials,
      custom_url: server.custom_url,
      enabled: !server.enabled
    }).unwrap();
  };

  if (step === 'pick-provider') {
    return <MCPProviderPicker providers={catalog} onSelect={handleProviderSelect} />;
  }

  if (step === 'fill-credentials' && selectedProvider) {
    return (
      <MCPCredentialForm
        provider={selectedProvider}
        initialCredentials={editingServer?.credentials}
        initialCustomURL={editingServer?.custom_url}
        initialName={editingServer?.name}
        onSave={handleSave}
        onBack={() => {
          setStep(editingServer ? 'list' : 'pick-provider');
          setEditingServer(null);
        }}
        isSaving={isAdding || isUpdating}
      />
    );
  }

  return (
    <div className="flex flex-col gap-3">
      {servers.length === 0 && (
        <p className="text-sm text-muted-foreground text-center py-4">
          No MCP servers configured yet.
        </p>
      )}
      {servers.map((server) => (
        <div
          key={server.id}
          className="flex items-center justify-between p-3 rounded-lg border border-border"
        >
          <div className="flex flex-col gap-0.5">
            <span className="text-sm font-medium">{server.name}</span>
            <span className="text-xs text-muted-foreground capitalize">{server.provider_id}</span>
          </div>
          <div className="flex items-center gap-2">
            {canConfigure && (
              <Switch
                checked={server.enabled}
                onCheckedChange={() => handleToggleEnabled(server)}
              />
            )}
            {canConfigure && (
              <Button variant="ghost" size="icon" onClick={() => handleEdit(server)} title="Edit">
                <Pencil className="h-4 w-4" />
              </Button>
            )}
            {canDelete && (
              <Button
                variant="ghost"
                size="icon"
                className="text-destructive hover:text-destructive"
                onClick={() => deleteServer({ id: server.id })}
                title="Delete"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      ))}
      {canConfigure && (
        <Button variant="outline" className="mt-2" onClick={() => setStep('pick-provider')}>
          + Add Server
        </Button>
      )}
    </div>
  );
}
