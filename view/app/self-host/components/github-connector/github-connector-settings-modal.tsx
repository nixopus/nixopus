'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import DialogWrapper, { DialogAction } from '@/components/ui/dialog-wrapper';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Github,
  Settings,
  Trash2,
  RefreshCw,
  Plus,
  Check,
  ExternalLink,
  AlertCircle
} from 'lucide-react';
import { GithubConnector } from '@/redux/types/github';
import useGithubConnectorSettings from '../../hooks/use-github-connector-settings';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { DeleteDialog } from '@/components/ui/delete-dialog';

interface ConnectorItemProps {
  connector: GithubConnector;
  isActive: boolean;
  onSetActive: (id: string) => void;
  onDelete: (id: string) => void;
  onReset: (id: string) => void;
  isDeleting: boolean;
  isUpdating: boolean;
}

const ConnectorItem: React.FC<ConnectorItemProps> = ({
  connector,
  isActive,
  onSetActive,
  onDelete,
  onReset,
  isDeleting,
  isUpdating
}) => {
  const { t } = useTranslation();
  const [showDeleteDialog, setShowDeleteDialog] = React.useState(false);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowDeleteDialog(true);
  };

  const confirmDelete = () => {
    onDelete(connector.id);
    setShowDeleteDialog(false);
  };

  const handleCardClick = () => {
    if (!isActive && !isDeleting && !isUpdating) {
      onSetActive(connector.id);
    }
  };

  return (
    <>
      <div
        onClick={handleCardClick}
        className={`flex items-center justify-between p-4 rounded-lg border transition-all ${
          isActive
            ? 'border-primary bg-primary/5 cursor-default'
            : 'border-border bg-card cursor-pointer hover:border-primary/50 hover:bg-muted/50'
        } ${isDeleting || isUpdating ? 'opacity-50 cursor-not-allowed' : ''}`}
      >
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <div className="rounded-full bg-primary/10 p-2 shrink-0">
            <Github size={20} className="text-primary" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <p className="font-medium text-sm truncate">{connector.name || connector.slug}</p>
              {isActive && (
                <Badge variant="default" className="text-xs">
                  <Check size={12} className="mr-1" />
                  {t('selfHost.connectorSettings.active' as any)}
                </Badge>
              )}
            </div>
            <p className="text-xs text-muted-foreground truncate">
              {t('selfHost.connectorSettings.connector.slug' as any)}: {connector.slug || 'N/A'}
            </p>
            {connector.installation_id && (
              <p className="text-xs text-muted-foreground truncate">
                {t('selfHost.connectorSettings.connector.installationId' as any)}:{' '}
                {connector.installation_id.substring(0, 8)}...
              </p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0" onClick={(e) => e.stopPropagation()}>
          <Button
            variant="destructive"
            size="sm"
            onClick={handleDelete}
            disabled={isDeleting || isUpdating}
            title={t('selfHost.connectorSettings.actions.delete.label' as any)}
          >
            <Trash2 size={16} />
          </Button>
        </div>
      </div>

      <DeleteDialog
        open={showDeleteDialog}
        onOpenChange={setShowDeleteDialog}
        title={t('selfHost.connectorSettings.actions.delete.dialog.title' as any)}
        description={t(
          'selfHost.connectorSettings.actions.delete.dialog.description' as any
        ).replace('{name}', connector.name || connector.slug)}
        onConfirm={confirmDelete}
        confirmText={
          isDeleting
            ? t('selfHost.connectorSettings.actions.delete.dialog.deleting' as any)
            : t('selfHost.connectorSettings.actions.delete.dialog.confirm' as any)
        }
        cancelText={t('selfHost.connectorSettings.actions.delete.dialog.cancel' as any)}
        isDeleting={isDeleting}
        variant="destructive"
        icon={Trash2}
      />
    </>
  );
};

interface GitHubConnectorSettingsModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAddNew?: () => void;
}

const GitHubConnectorSettingsModal: React.FC<GitHubConnectorSettingsModalProps> = ({
  open,
  onOpenChange,
  onAddNew
}) => {
  const { t } = useTranslation();
  const {
    connectors,
    activeConnector,
    activeConnectorId,
    isLoadingConnectors,
    isDeleting,
    isUpdating,
    handleSetActiveConnector,
    handleDeleteConnector,
    handleResetConnector,
    setIsSettingsModalOpen
  } = useGithubConnectorSettings();

  const handleClose = () => {
    setIsSettingsModalOpen(false);
    onOpenChange(false);
  };

  const handleAddNew = () => {
    handleClose();
    onAddNew?.();
  };

  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: handleClose,
      variant: 'outline'
    },
    ...(onAddNew
      ? [
          {
            label: t('selfHost.connectorSettings.actions.addNew' as any),
            onClick: handleAddNew,
            variant: 'default' as const,
            icon: Plus
          }
        ]
      : [])
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={handleClose}
      title={t('selfHost.connectorSettings.title' as any)}
      description={t('selfHost.connectorSettings.description' as any)}
      actions={actions}
      size="lg"
    >
      <div className="space-y-4">
        {isLoadingConnectors ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-20 w-full" />
            ))}
          </div>
        ) : connectors.length === 0 ? (
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {t('selfHost.connectorSettings.noConnectors' as any)}
            </AlertDescription>
          </Alert>
        ) : (
          <>
            <div className="space-y-3">
              {connectors.map((connector) => (
                <ConnectorItem
                  key={connector.id}
                  connector={connector}
                  isActive={connector.id === activeConnectorId}
                  onSetActive={handleSetActiveConnector}
                  onDelete={handleDeleteConnector}
                  onReset={handleResetConnector}
                  isDeleting={isDeleting === connector.id}
                  isUpdating={isUpdating === connector.id}
                />
              ))}
            </div>
            {(() => {
              const currentActiveConnector = connectors.find((c) => c.id === activeConnectorId);

              if (!currentActiveConnector) {
                return null;
              }

              const installationUrl = `https://github.com/settings/installations/${currentActiveConnector.installation_id}`;

              return (
                <div className="pt-4 border-t">
                  <div className="flex items-start gap-3">
                    <Settings size={18} className="text-muted-foreground mt-0.5 shrink-0" />
                    <div className="flex-1 space-y-2">
                      <div>
                        <p className="text-sm font-medium mb-1">
                          {t('selfHost.connectorSettings.currentConnector.title' as any)}
                        </p>
                        <div className="flex items-center gap-2 mb-2">
                          <Badge variant="secondary" className="text-xs">
                            {currentActiveConnector.name || currentActiveConnector.slug}
                          </Badge>
                        </div>
                      </div>
                      <p className="text-xs text-muted-foreground leading-relaxed">
                        {t('selfHost.connectorSettings.currentConnector.description' as any)}
                      </p>
                      {currentActiveConnector.installation_id && (
                        <a
                          href={installationUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          onClick={(e) => {
                            console.log('[GitHubConnectorSettingsModal] Link clicked:', {
                              href: installationUrl,
                              connectorId: currentActiveConnector.id,
                              installationId: currentActiveConnector.installation_id,
                              activeConnectorId,
                              timestamp: new Date().toISOString()
                            });
                          }}
                          className="inline-flex items-center gap-1.5 text-xs text-primary hover:underline font-medium"
                        >
                          {t('selfHost.connectorSettings.currentConnector.viewOnGithub' as any)}
                          <ExternalLink size={12} />
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              );
            })()}
          </>
        )}
      </div>
    </DialogWrapper>
  );
};

export default GitHubConnectorSettingsModal;
