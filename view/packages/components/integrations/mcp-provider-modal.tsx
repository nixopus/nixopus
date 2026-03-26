'use client';

import React from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, Button } from '@nixopus/ui';
import { MCPCredentialForm } from './mcp-credential-form';
import {
  useAddMCPServerMutation,
  useUpdateMCPServerMutation,
  useDeleteMCPServerMutation
} from '@/redux/services/settings/mcpApi';
import type { MCPProvider, MCPServer } from '@/redux/types/mcp';

interface MCPProviderModalProps {
  provider: MCPProvider;
  server: MCPServer | null;
  onClose: () => void;
  canDelete: boolean;
}

export function MCPProviderModal({ provider, server, onClose, canDelete }: MCPProviderModalProps) {
  const [addServer, { isLoading: isAdding }] = useAddMCPServerMutation();
  const [updateServer, { isLoading: isUpdating }] = useUpdateMCPServerMutation();
  const [deleteServer, { isLoading: isDeleting }] = useDeleteMCPServerMutation();

  const handleSave = async (data: {
    name: string;
    credentials: Record<string, string>;
    custom_url?: string;
    enabled: boolean;
  }) => {
    if (server) {
      await updateServer({ id: server.id, ...data }).unwrap();
    } else {
      await addServer({ provider_id: provider.id, ...data }).unwrap();
    }
    onClose();
  };

  const handleDelete = async () => {
    if (!server) return;
    await deleteServer({ id: server.id }).unwrap();
    onClose();
  };

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{provider.name}</DialogTitle>
        </DialogHeader>
        <MCPCredentialForm
          provider={provider}
          initialCredentials={server?.credentials}
          initialCustomURL={server?.custom_url}
          initialName={server?.name}
          onSave={handleSave}
          onBack={onClose}
          isSaving={isAdding || isUpdating}
        />
        {server && canDelete && (
          <div className="pt-2 border-t border-border">
            <Button
              variant="ghost"
              size="sm"
              className="text-destructive hover:text-destructive w-full"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? 'Removing…' : 'Remove Server'}
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
