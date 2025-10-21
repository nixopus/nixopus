'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { Card, CardDescription, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { ExternalLink, Check, GitFork, Trash2 } from 'lucide-react';

import { Extension } from '@/redux/types/extension';
import ExtensionForkDialog from './extension-fork-dialog';
import { useDeleteExtensionMutation } from '@/redux/services/extensions/extensionsApi';
import { toast } from 'sonner';

const MAX_DESCRIPTION_CHARS = 90;

interface ExtensionCardProps {
  extension: Extension;
  onInstall?: (extension: Extension) => void;
  onViewDetails?: (extension: Extension) => void;
}

export function ExtensionCard({ extension, onInstall, onViewDetails }: ExtensionCardProps) {
  const { t } = useTranslation();
  const [forkOpen, setForkOpen] = React.useState(false);
  const [confirmOpen, setConfirmOpen] = React.useState(false);
  const [deleteExtension] = useDeleteExtensionMutation();
  const [expanded, setExpanded] = React.useState(false);

  const onDelete = async () => {
    try {
      await deleteExtension({ id: extension.id }).unwrap();
      toast.success(t('extensions.deleteSuccess') || 'Removed');
    } catch (e) {
      toast.error(t('extensions.deleteFailed') || 'Remove failed');
    }
  };

  return (
    <Card className="group h-full transition-all duration-200 hover:shadow-lg bg-card border-border p-6">
      <div className="space-y-4">
        <div className="flex items-start gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-primary-foreground flex-shrink-0">
            <div className="text-muted-foreground text-lg font-bold">{extension.icon}</div>
          </div>
          <div className="flex-1 min-w-0">
            <CardTitle className="text-lg font-bold text-card-foreground mb-1">
              {extension.name}
            </CardTitle>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">
                {t('extensions.madeBy')} {extension.author}
              </span>
              {extension.is_verified && (
                <div className="flex h-4 w-4 items-center justify-center rounded-full bg-primary">
                  <Check className="h-2.5 w-2.5 text-primary-foreground" />
                </div>
              )}
            </div>
          </div>
          <div className="ml-auto -mt-2 flex items-center gap-1">
            {!extension.parent_extension_id && (
              <button
                aria-label={t('extensions.fork') || 'Fork'}
                className="p-2 rounded-md hover:bg-accent text-muted-foreground"
                onClick={() => setForkOpen(true)}
              >
                <GitFork className="h-4 w-4" />
              </button>
            )}
            {extension.parent_extension_id && (
              <button
                aria-label={t('extensions.remove') || 'Remove'}
                className="p-2 rounded-md hover:bg-destructive/10 text-destructive"
                onClick={() => setConfirmOpen(true)}
              >
                <Trash2 className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>

        <CardDescription className="text-sm leading-relaxed text-muted-foreground">
          {expanded || extension.description.length <= MAX_DESCRIPTION_CHARS
            ? extension.description
            : `${extension.description.slice(0, MAX_DESCRIPTION_CHARS)}…`}
          {extension.description.length > MAX_DESCRIPTION_CHARS && (
            <button
              className="ml-2 text-primary hover:underline text-sm"
              onClick={() => setExpanded((v) => !v)}
            >
              {expanded ? t('common.readLess') || 'Read less' : t('common.readMore') || 'Read more'}
            </button>
          )}
        </CardDescription>
        <div className="flex gap-2 pt-6 justify-start">
          <Button className="font-medium min-w-[100px]" onClick={() => onInstall?.(extension)}>
            {extension.extension_type === 'install' ? t('extensions.install') : t('extensions.run')}
          </Button>
          <Button
            variant="ghost"
            onClick={() => onViewDetails?.(extension)}
            className="border-border hover:bg-accent text-card-foreground min-w-[100px] whitespace-nowrap"
          >
            {t('extensions.viewDetails')}
            <ExternalLink className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </div>
      <ExtensionForkDialog open={forkOpen} onOpenChange={setForkOpen} extension={extension} />
      <DeleteDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title={t('extensions.confirmDeleteTitle') || 'Remove fork?'}
        description={
          t('extensions.confirmDeleteMessage') ||
          'This will remove your forked extension. This action cannot be undone.'
        }
        confirmText={t('common.delete') || 'Delete'}
        cancelText={t('common.cancel') || 'Cancel'}
        variant="destructive"
        onConfirm={async () => {
          await onDelete();
        }}
      />
    </Card>
  );
}
