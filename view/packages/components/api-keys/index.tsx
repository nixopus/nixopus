'use client';

import * as React from 'react';
import { useMemo } from 'react';
import {
  DataTable,
  Badge,
  Button,
  DropdownMenuItem,
  DialogWrapper,
  Input,
  Label,
  SelectWrapper,
  Alert,
  AlertDescription,
  PaginationWrapper,
  SearchBar,
  type SelectOption,
} from '@nixopus/ui';
import { usePagination } from '@/packages/hooks/shared/use-pagination';
import { Pencil, Trash2, Key, Copy, Check, AlertTriangle, Plus } from 'lucide-react';
import PageLayout from '@/packages/layouts/page-layout';
import { useApiKeys } from '@/packages/hooks/api-keys/use-api-keys';
import {
  useApiKeyCreatedDialog,
  useCreateApiKeyDialog,
  useRenameApiKeyDialog,
  useDeleteApiKeyDialog,
  useApiKeysListDialogs,
  getExpiryStatus,
  type ApiKey,
} from '@/packages/hooks/api-keys/use-api-keys-view';
import {
  DATA_TABLE_CLASS,
  RowActionsDropdown,
  TablePanel,
  TableSkeleton,
  EmptyState,
  ListToolbar,
  RefreshButton,
} from '@/components/ui/list-page-chrome';
import { format, formatDistanceToNow } from 'date-fns';

interface ApiKeyCreatedDialogProps {
  apiKey: string | null;
  onClose: () => void;
}

export const ApiKeyCreatedDialog = React.memo(function ApiKeyCreatedDialog({
  apiKey,
  onClose,
}: ApiKeyCreatedDialogProps) {
  const { copied, handleCopy, actions } = useApiKeyCreatedDialog(apiKey, onClose);

  return (
    <DialogWrapper
      open={!!apiKey}
      onOpenChange={onClose}
      title="API Key Created"
      description="Your new API key has been generated."
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[550px]"
    >
      <div className="space-y-4">
        <Alert variant="default">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            Copy this key now. You won&apos;t be able to see it again.
          </AlertDescription>
        </Alert>
        <div className="flex gap-2">
          <Input readOnly value={apiKey || ''} className="font-mono text-sm" />
          <Button variant="outline" size="icon" onClick={handleCopy} className="shrink-0">
            {copied ? (
              <Check className="h-4 w-4 text-green-500" />
            ) : (
              <Copy className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>
    </DialogWrapper>
  );
});

interface CreateApiKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreate: (data: { name: string; expiresIn?: number }) => Promise<unknown>;
  isCreating: boolean;
  expiryOptions: readonly { label: string; value: number }[];
}

export const CreateApiKeyDialog = React.memo(function CreateApiKeyDialog({
  open,
  onOpenChange,
  onCreate,
  isCreating,
  expiryOptions,
}: CreateApiKeyDialogProps) {
  const { name, expiresIn, setExpiresIn, handleSubmit, actions, handleNameChange } =
    useCreateApiKeyDialog(open, onOpenChange, onCreate, isCreating);

  const expirySelectOptions: SelectOption[] = useMemo(
    () => expiryOptions.map((opt) => ({ value: String(opt.value), label: opt.label })),
    [expiryOptions],
  );

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title="Create API Key"
      description="Generate a new API key for programmatic access to your resources."
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[500px]"
    >
      <form id="create-api-key-form" onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="key-name">Name</Label>
          <Input
            id="key-name"
            placeholder="e.g. CI/CD Pipeline"
            value={name}
            onChange={handleNameChange}
            disabled={isCreating}
            autoFocus
          />
        </div>
        <div className="space-y-2">
          <Label>Expiration</Label>
          <SelectWrapper
            value={expiresIn}
            onValueChange={setExpiresIn}
            options={expirySelectOptions}
            placeholder="Expiration"
            disabled={isCreating}
            className="w-full"
          />
        </div>
      </form>
    </DialogWrapper>
  );
});

interface RenameApiKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentName: string;
  onRename: (name: string) => Promise<boolean>;
  isUpdating: boolean;
}

const RenameApiKeyDialog = React.memo(function RenameApiKeyDialog({
  open,
  onOpenChange,
  currentName,
  onRename,
  isUpdating,
}: RenameApiKeyDialogProps) {
  const { name, handleSubmit, actions, handleNameChange } = useRenameApiKeyDialog(
    open,
    onOpenChange,
    currentName,
    onRename,
    isUpdating,
  );

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title="Rename API Key"
      description="Give this API key a new name to help you identify it."
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[450px]"
    >
      <form id="rename-api-key-form" onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="new-name">Name</Label>
          <Input
            id="new-name"
            value={name}
            onChange={handleNameChange}
            disabled={isUpdating}
            autoFocus
          />
        </div>
      </form>
    </DialogWrapper>
  );
});

interface DeleteApiKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  keyName: string;
  onDelete: () => void;
  isDeleting: boolean;
}

const DeleteApiKeyDialog = React.memo(function DeleteApiKeyDialog({
  open,
  onOpenChange,
  keyName,
  onDelete,
  isDeleting,
}: DeleteApiKeyDialogProps) {
  const { actions } = useDeleteApiKeyDialog(onOpenChange, onDelete, isDeleting);

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title="Delete API Key"
      description={`Are you sure you want to delete "${keyName}"?`}
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[450px]"
    >
      <Alert variant="destructive">
        <AlertTriangle className="h-4 w-4" />
        <AlertDescription>
          This action cannot be undone. Any integrations using this key will stop working
          immediately.
        </AlertDescription>
      </Alert>
    </DialogWrapper>
  );
});

interface ApiKeysListProps {
  apiKeys: ApiKey[];
  onRename: (keyId: string, name: string) => Promise<boolean>;
  onDelete: (keyId: string) => Promise<void>;
  isUpdating: boolean;
  isDeleting: boolean;
}

export const ApiKeysList = React.memo(function ApiKeysList({
  apiKeys,
  onRename,
  onDelete,
  isUpdating,
  isDeleting,
}: ApiKeysListProps) {
  const {
    renameTarget,
    setRenameTarget,
    deleteTarget,
    setDeleteTarget,
    handleCloseRename,
    handleCloseDelete,
    handleRenameKey,
    handleDeleteKey,
  } = useApiKeysListDialogs(onRename, onDelete);

  const columns = useMemo(
    () => [
      {
        key: 'name',
        title: 'Name',
        render: (_: unknown, key: ApiKey) => (
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-medium truncate">{key.name || 'Untitled'}</span>
            <span className="text-xs text-muted-foreground font-mono truncate">
              {key.prefix ? `${key.prefix}` : ''}
              {key.start || '••••'}...
            </span>
          </div>
        ),
      },
      {
        key: 'status',
        title: 'Status',
        render: (_: unknown, key: ApiKey) => {
          const expiry = getExpiryStatus(key.expiresAt);
          return (
            <div className="flex items-center gap-2">
              <Badge
                variant={key.enabled ? 'default' : 'destructive'}
                className="rounded-sm opacity-80"
              >
                {key.enabled ? 'Active' : 'Disabled'}
              </Badge>
              <Badge variant={expiry.variant} className="rounded-sm opacity-80">
                {expiry.label}
              </Badge>
            </div>
          );
        },
      },
      {
        key: 'lastUsed',
        title: 'Last Used',
        render: (_: unknown, key: ApiKey) => (
          <span className="text-sm text-muted-foreground">
            {key.lastRequest
              ? formatDistanceToNow(new Date(key.lastRequest), { addSuffix: true })
              : 'Never'}
          </span>
        ),
      },
      {
        key: 'created',
        title: 'Created',
        render: (_: unknown, key: ApiKey) => (
          <span className="text-sm text-muted-foreground">
            {format(new Date(key.createdAt), 'MMM d, yyyy')}
          </span>
        ),
      },
      {
        key: 'actions',
        title: '',
        render: (_: unknown, key: ApiKey) => (
          <div className="flex items-center justify-end">
            <RowActionsDropdown ariaLabel="Key actions">
              <DropdownMenuItem onClick={() => setRenameTarget(key)}>
                <Pencil className="h-4 w-4 mr-2" />
                Rename
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => setDeleteTarget(key)}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </RowActionsDropdown>
          </div>
        ),
      },
    ],
    [setRenameTarget, setDeleteTarget],
  );

  const { paginatedItems, currentPage, totalPages, handlePageChange } = usePagination(apiKeys);

  return (
    <>
      <TablePanel>
        {apiKeys.length === 0 ? (
          <EmptyState
            icon={Key}
            message="No API keys yet"
            description="Create a key to get started with programmatic access."
            bordered={false}
          />
        ) : (
          <>
            <DataTable
              data={paginatedItems}
              columns={columns}
              showBorder={false}
              hoverable
              tableClassName={DATA_TABLE_CLASS}
            />
            {totalPages > 1 && (
              <div className="py-4 flex justify-center border-t border-border">
                <PaginationWrapper
                  currentPage={currentPage}
                  totalPages={totalPages}
                  onPageChange={handlePageChange}
                />
              </div>
            )}
          </>
        )}
      </TablePanel>

      {renameTarget && (
        <RenameApiKeyDialog
          open={!!renameTarget}
          onOpenChange={handleCloseRename}
          currentName={renameTarget.name || ''}
          onRename={handleRenameKey}
          isUpdating={isUpdating}
        />
      )}

      {deleteTarget && (
        <DeleteApiKeyDialog
          open={!!deleteTarget}
          onOpenChange={handleCloseDelete}
          keyName={deleteTarget.name || 'this key'}
          onDelete={handleDeleteKey}
          isDeleting={isDeleting}
        />
      )}
    </>
  );
});

const API_KEYS_TABLE_SKELETON = {
  headerWidths: ['w-20', 'w-16', 'w-20', 'w-16'],
  rows: 3,
  rowConfig: {
    widths: [
      'h-3.5 w-32 flex-1 max-w-[200px]',
      'h-5 w-16 rounded-sm',
      'h-5 w-28 rounded-sm',
      'h-3.5 w-20',
      'h-3.5 w-24',
      'h-7 w-7 ml-auto rounded',
    ],
  },
};

export const ApiKeysSettings = React.memo(function ApiKeysSettings() {
  const {
    apiKeys,
    isLoading,
    isFetching,
    error,
    isCreating,
    isUpdating,
    isDeleting,
    newlyCreatedKey,
    clearNewKey,
    handleCreate,
    handleRename,
    handleDelete,
    expiryOptions,
    refetch,
    searchTerm,
    setSearchTerm,
    createDialogOpen,
    setCreateDialogOpen,
  } = useApiKeys();

  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <h1 className="text-2xl tracking-wide pb-4 sm:pb-6">API Keys</h1>

      <div className="pb-4 pt-6 min-w-0">
        <ListToolbar
          left={
            <SearchBar
              label="Search keys..."
              searchTerm={searchTerm}
              handleSearchChange={(e) => setSearchTerm(e.target.value)}
              className="w-full max-w-xs [&_input]:w-full!"
            />
          }
        >
          <RefreshButton onClick={refetch} isFetching={isFetching} ariaLabel="Refresh API keys" />
          <Button
            onClick={() => setCreateDialogOpen(true)}
            size="sm"
            className="gap-2"
            disabled={isCreating}
          >
            <Plus className="h-4 w-4" />
            Create API Key
          </Button>
        </ListToolbar>
      </div>

      {isLoading ? (
        <TableSkeleton {...API_KEYS_TABLE_SKELETON} />
      ) : error ? (
        <EmptyState
          icon={Key}
          message="Failed to load API keys"
          description="Something went wrong. Please try again."
          onRetry={refetch}
        />
      ) : (
        <ApiKeysList
          apiKeys={apiKeys}
          onRename={handleRename}
          onDelete={handleDelete}
          isUpdating={isUpdating}
          isDeleting={isDeleting}
        />
      )}

      {createDialogOpen && (
        <CreateApiKeyDialog
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          onCreate={handleCreate}
          isCreating={isCreating}
          expiryOptions={expiryOptions}
        />
      )}

      {newlyCreatedKey && (
        <ApiKeyCreatedDialog apiKey={newlyCreatedKey} onClose={clearNewKey} />
      )}
    </PageLayout>
  );
});
