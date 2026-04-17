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
  Alert,
  AlertDescription,
  SearchBar,
} from '@nixopus/ui';
import { Pencil, Trash2, Fingerprint, Plus, Loader2 } from 'lucide-react';
import PageLayout from '@/packages/layouts/page-layout';
import { usePasskeys, type Passkey } from '@/packages/hooks/security/use-passkeys';
import { usePasskeyList } from '@/packages/hooks/security/use-security';
import {
  DATA_TABLE_CLASS,
  RowActionsDropdown,
  TablePanel,
  TableSkeleton,
  EmptyState,
  ListToolbar,
  RefreshButton,
} from '@/components/ui/list-page-chrome';
import { format } from 'date-fns';

const PASSKEY_TABLE_SKELETON = {
  headerWidths: ['w-20', 'w-14', 'w-20', 'w-16'],
  rows: 2,
  rowConfig: {
    widths: [
      'h-4 w-4 rounded shrink-0',
      'h-3.5 w-36',
      'h-5 w-20 rounded-sm',
      'h-5 w-16 rounded-sm',
      'h-3.5 w-24',
      'h-7 w-7 ml-auto rounded',
    ],
  },
};

interface PasskeyListProps {
  passkeys: Passkey[];
  addDialogOpen: boolean;
  setAddDialogOpen: (open: boolean) => void;
  addPasskey: (name?: string) => Promise<boolean>;
  deletePasskey: (id: string) => Promise<void>;
  updatePasskey: (id: string, name: string) => Promise<boolean>;
  isAdding: boolean;
  isDeleting: boolean;
  isUpdating: boolean;
}

function PasskeyList({
  passkeys,
  addDialogOpen,
  setAddDialogOpen,
  addPasskey,
  deletePasskey,
  updatePasskey,
  isAdding,
  isDeleting,
  isUpdating,
}: PasskeyListProps) {
  const {
    renameTarget,
    setRenameTarget,
    deleteTarget,
    setDeleteTarget,
    newName,
    setNewName,
    passkeyName,
    setPasskeyName,
    addActions,
    renameActions,
    deleteActions,
  } = usePasskeyList(setAddDialogOpen, {
    addPasskey,
    deletePasskey,
    updatePasskey,
    isAdding,
    isDeleting,
    isUpdating,
  });

  const columns = useMemo(
    () => [
      {
        key: 'name',
        title: 'Name',
        render: (_: unknown, pk: Passkey) => (
          <div className="flex items-center gap-3 min-w-0">
            <Fingerprint className="h-4 w-4 text-muted-foreground shrink-0" />
            <span className="text-sm font-medium truncate">{pk.name || 'Unnamed passkey'}</span>
          </div>
        ),
      },
      {
        key: 'type',
        title: 'Type',
        render: (_: unknown, pk: Passkey) => (
          <Badge variant="secondary" className="rounded-sm opacity-80">
            {pk.deviceType === 'singleDevice' ? 'This device' : 'Multi-device'}
          </Badge>
        ),
      },
      {
        key: 'backedUp',
        title: 'Backed Up',
        render: (_: unknown, pk: Passkey) => (
          <Badge variant={pk.backedUp ? 'default' : 'outline'} className="rounded-sm opacity-80">
            {pk.backedUp ? 'Yes' : 'No'}
          </Badge>
        ),
      },
      {
        key: 'created',
        title: 'Created',
        render: (_: unknown, pk: Passkey) => (
          <span className="text-sm text-muted-foreground">
            {pk.createdAt ? format(new Date(pk.createdAt), 'MMM d, yyyy') : '-'}
          </span>
        ),
      },
      {
        key: 'actions',
        title: '',
        render: (_: unknown, pk: Passkey) => (
          <div className="flex items-center justify-end">
            <RowActionsDropdown ariaLabel="Passkey actions">
              <DropdownMenuItem
                onClick={() => {
                  setRenameTarget(pk);
                  setNewName(pk.name || '');
                }}
              >
                <Pencil className="h-4 w-4 mr-2" />
                Rename
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => setDeleteTarget(pk)}
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
    [setRenameTarget, setDeleteTarget, setNewName],
  );

  return (
    <>
      <TablePanel>
        {passkeys.length === 0 ? (
          <EmptyState
            icon={Fingerprint}
            message="No passkeys registered"
            description="Add a passkey to sign in securely without a password."
            bordered={false}
          />
        ) : (
          <DataTable
            data={passkeys}
            columns={columns}
            showBorder={false}
            hoverable
            tableClassName={DATA_TABLE_CLASS}
          />
        )}
      </TablePanel>

      {addDialogOpen && (
        <DialogWrapper
          open={addDialogOpen}
          onOpenChange={setAddDialogOpen}
          title="Add Passkey"
          actions={addActions}
          size="sm"
          contentClassName="sm:max-w-[450px]"
        >
          <div className="grid gap-3 py-2">
            <Label htmlFor="passkey-name">Name (optional)</Label>
            <Input
              id="passkey-name"
              placeholder="e.g. MacBook Touch ID"
              value={passkeyName}
              onChange={(e) => setPasskeyName(e.target.value)}
            />
          </div>
        </DialogWrapper>
      )}

      {renameTarget && (
        <DialogWrapper
          open={!!renameTarget}
          onOpenChange={(open) => !open && setRenameTarget(null)}
          title="Rename Passkey"
          actions={renameActions}
          size="sm"
          contentClassName="sm:max-w-[450px]"
        >
          <div className="grid gap-3 py-2">
            <Label htmlFor="rename-passkey">Name</Label>
            <Input
              id="rename-passkey"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              autoFocus
            />
          </div>
        </DialogWrapper>
      )}

      {deleteTarget && (
        <DialogWrapper
          open={!!deleteTarget}
          onOpenChange={(open) => !open && setDeleteTarget(null)}
          title="Delete Passkey"
          description={`Are you sure you want to delete "${deleteTarget.name || 'this passkey'}"?`}
          actions={deleteActions}
          size="sm"
          contentClassName="sm:max-w-[450px]"
        >
          <Alert variant="destructive">
            <AlertDescription>
              This action cannot be undone. You will no longer be able to verify your identity with
              this passkey.
            </AlertDescription>
          </Alert>
        </DialogWrapper>
      )}
    </>
  );
}

export const SecuritySettings = React.memo(function SecuritySettings() {
  const {
    passkeys,
    isLoading,
    isFetching,
    error,
    isAdding,
    isDeleting,
    isUpdating,
    addPasskey,
    deletePasskey,
    updatePasskey,
    refetch,
    searchTerm,
    setSearchTerm,
    addDialogOpen,
    setAddDialogOpen,
  } = usePasskeys();

  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <h1 className="text-2xl tracking-wide pb-4 sm:pb-6">Security</h1>

      <div className="pb-4 pt-6 min-w-0">
        <ListToolbar
          left={
            <SearchBar
              label="Search passkeys..."
              searchTerm={searchTerm}
              handleSearchChange={(e) => setSearchTerm(e.target.value)}
              className="w-full max-w-xs [&_input]:w-full!"
            />
          }
        >
          <RefreshButton onClick={refetch} isFetching={isFetching} ariaLabel="Refresh passkeys" />
          <Button
            size="sm"
            onClick={() => setAddDialogOpen(true)}
            className="gap-2"
            disabled={isAdding}
          >
            {isAdding ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Plus className="h-4 w-4" />
            )}
            {isAdding ? 'Registering...' : 'Add Passkey'}
          </Button>
        </ListToolbar>
      </div>

      {isLoading ? (
        <TableSkeleton {...PASSKEY_TABLE_SKELETON} />
      ) : error ? (
        <EmptyState
          icon={Fingerprint}
          message="Failed to load passkeys"
          description="Something went wrong. Please try again."
          onRetry={refetch}
        />
      ) : (
        <PasskeyList
          passkeys={passkeys}
          addDialogOpen={addDialogOpen}
          setAddDialogOpen={setAddDialogOpen}
          addPasskey={addPasskey}
          deletePasskey={deletePasskey}
          updatePasskey={updatePasskey}
          isAdding={isAdding}
          isDeleting={isDeleting}
          isUpdating={isUpdating}
        />
      )}
    </PageLayout>
  );
});
