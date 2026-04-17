'use client';

import * as React from 'react';
import { useMemo, type ReactNode } from 'react';
import PageLayout from '@/packages/layouts/page-layout';
import {
  DataTable,
  Button,
  Badge,
  DropdownMenuItem,
  DialogWrapper,
  Input,
  Alert,
  AlertDescription,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  SearchBar,
} from '@nixopus/ui';
import {
  Plus,
  Trash2,
  RotateCw,
  ExternalLink,
  Globe,
  Copy,
  CheckCircle,
  AlertTriangle,
} from 'lucide-react';
import { format } from 'date-fns';
import type { CustomDomain, DNSInstruction } from '@/redux/types/custom-domains';
import {
  useDomainsView,
  useAddDomainDialog,
  useRemoveDomainDialog,
  useDnsSetupWizard,
  useCopyField,
  statusConfig,
  providerNames,
} from '@/packages/hooks/domains/use-domains-view';
import {
  DATA_TABLE_CLASS,
  RowActionsDropdown,
  TablePanel,
  TableSkeleton,
  EmptyState,
  ListToolbar,
  RefreshButton,
} from '@/components/ui/list-page-chrome';

const DOMAINS_TABLE_SKELETON = {
  headerWidths: ['w-28', 'w-14', 'w-16', 'w-24'],
  rows: 3,
  rowConfig: {
    widths: [
      'h-3.5 w-36 flex-1 max-w-[240px]',
      'h-5 w-16 rounded-full',
      'h-5 w-20 rounded-full',
      'h-3.5 w-32',
      'h-7 w-7 ml-auto rounded',
    ],
  },
};

const DomainStatusBadge = React.memo(function DomainStatusBadge({
  status,
}: {
  status: CustomDomain['status'];
}) {
  const config = statusConfig[status];
  return (
    <Badge variant={config.variant} className="rounded-sm opacity-80">
      {config.label}
    </Badge>
  );
});

const CopyField = React.memo(function CopyField({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  const { copied, handleCopy } = useCopyField(value);

  return (
    <div className="space-y-1">
      <span className="text-xs font-medium text-muted-foreground">{label}</span>
      <div className="flex items-center gap-2">
        <code className="flex-1 rounded bg-muted px-3 py-2 text-sm font-mono break-all">
          {value}
        </code>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0 shrink-0" onClick={handleCopy}>
          {copied ? (
            <CheckCircle className="h-4 w-4 text-green-500" />
          ) : (
            <Copy className="h-4 w-4" />
          )}
        </Button>
      </div>
    </div>
  );
});

const AddDomainDialog = React.memo(function AddDomainDialog({
  open,
  onOpenChange,
  onAdd,
  isAdding,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdd: (name: string) => void;
  isAdding: boolean;
}) {
  const { domainName, validationError, handleAdd, handleInputChange, handleKeyDown, reset } =
    useAddDomainDialog(onAdd);

  const handleOpenChange = React.useCallback(
    (value: boolean) => {
      if (!value) reset();
      onOpenChange(value);
    },
    [onOpenChange, reset],
  );

  const actions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => handleOpenChange(false), variant: 'outline' as const },
      {
        label: isAdding ? 'Adding...' : 'Add Domain',
        onClick: handleAdd,
        variant: 'default' as const,
        disabled: isAdding || !domainName.trim(),
        loading: isAdding,
      },
    ],
    [handleOpenChange, handleAdd, isAdding, domainName],
  );

  return (
    <DialogWrapper
      open={open}
      onOpenChange={handleOpenChange}
      title="Add Domain"
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[450px]"
      headerClassName="text-center"
    >
      <div className="space-y-2">
        <Input
          id="domain-name"
          placeholder="app.example.com"
          value={domainName}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          autoFocus
        />
        {validationError && <p className="text-sm text-destructive">{validationError}</p>}
      </div>
    </DialogWrapper>
  );
});

const RemoveDomainDialog = React.memo(function RemoveDomainDialog({
  open,
  onOpenChange,
  domainName,
  onRemove,
  isRemoving,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  domainName: string;
  onRemove: () => void;
  isRemoving: boolean;
}) {
  const { handleRemove } = useRemoveDomainDialog(onRemove);

  const actions = useMemo(
    () => [
      { label: 'Cancel', onClick: () => onOpenChange(false), variant: 'outline' as const },
      {
        label: isRemoving ? 'Removing...' : 'Remove',
        onClick: handleRemove,
        variant: 'destructive' as const,
        disabled: isRemoving,
        loading: isRemoving,
      },
    ],
    [onOpenChange, handleRemove, isRemoving],
  );

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title="Remove Domain"
      description={`Are you sure you want to remove "${domainName}"?`}
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[450px]"
    >
      <Alert variant="destructive">
        <AlertTriangle className="h-4 w-4" />
        <AlertDescription>
          This action cannot be undone. Any services using this domain will become unreachable.
        </AlertDescription>
      </Alert>
    </DialogWrapper>
  );
});

const DNSSetupWizard = React.memo(function DNSSetupWizard({
  open,
  instructions,
  dnsProvider,
  domainId,
  onVerify,
  onClose,
  isVerifying,
}: {
  open: boolean;
  instructions: DNSInstruction[];
  dnsProvider: string;
  domainId: string;
  onVerify: (id: string) => void;
  onClose: () => void;
  isVerifying: boolean;
}) {
  const { handleVerify, handleOpenChange } = useDnsSetupWizard(onVerify, domainId, onClose);
  const friendlyProvider = providerNames[dnsProvider.toLowerCase()] || dnsProvider;

  const actions = useMemo(
    () => [
      { label: 'Close', onClick: onClose, variant: 'outline' as const },
      {
        label: isVerifying ? 'Verifying...' : 'Verify DNS',
        onClick: handleVerify,
        variant: 'default' as const,
        disabled: isVerifying,
        loading: isVerifying,
      },
    ],
    [onClose, handleVerify, isVerifying],
  );

  return (
    <DialogWrapper
      open={open}
      onOpenChange={handleOpenChange}
      title="DNS Setup"
      description="Add the following DNS records to your domain provider"
      actions={actions}
      size="lg"
      contentClassName="sm:max-w-[600px]"
    >
      <div className="space-y-4">
        {dnsProvider && dnsProvider !== 'other' && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Detected provider:</span>
            <Badge variant="secondary">{friendlyProvider}</Badge>
          </div>
        )}

        {instructions.map((instruction, index) => (
          <Card key={index}>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-sm">
                <Badge variant="outline">{instruction.record_type}</Badge>
                <span className="text-muted-foreground font-normal">{instruction.description}</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <CopyField label="Name" value={instruction.name} />
              <CopyField label="Value" value={instruction.value} />
            </CardContent>
          </Card>
        ))}

        <p className="text-xs text-muted-foreground text-center">
          DNS changes can take up to 48 hours to propagate
        </p>
      </div>
    </DialogWrapper>
  );
});

function CustomDomainsList() {
  const {
    filteredDomains,
    isLoading,
    error,
    isFetching,
    isAdding,
    isVerifying,
    isRemoving,
    dnsInstructions,
    dnsProvider,
    showDnsWizard,
    currentDomainId,
    showAddDialog,
    setShowAddDialog,
    removeTarget,
    setRemoveTarget,
    searchTerm,
    setSearchTerm,
    sortConfig,
    handleSort,
    handleVerifyDomain,
    onAdd,
    onConfirmRemove,
    onCloseRemoveDialog,
    onCloseDnsWizard,
    refetch,
  } = useDomainsView();

  const columns = useMemo(
    () => [
      {
        key: 'name',
        title: 'Domain',
        className: 'w-[40%]',
        sortable: true,
        render: (_: unknown, domain: CustomDomain) => (
          <div className="flex flex-col gap-0.5 py-0.5 min-w-0">
            <span className="text-sm font-medium text-foreground truncate">{domain.name}</span>
            {domain.target_subdomain && (
              <span className="text-[11px] text-muted-foreground/70 font-mono leading-none truncate">
                {domain.target_subdomain}
              </span>
            )}
          </div>
        ),
      },
      {
        key: 'type',
        title: 'Type',
        sortable: true,
        render: (_: unknown, domain: CustomDomain) => (
          <Badge variant="outline" className="rounded-sm opacity-80">
            {domain.type === 'custom' ? 'Custom' : 'System'}
          </Badge>
        ),
      },
      {
        key: 'status',
        title: 'Status',
        sortable: true,
        render: (_: unknown, domain: CustomDomain) => (
          <DomainStatusBadge status={domain.status} />
        ),
      },
      {
        key: 'created',
        title: 'Created',
        sortable: true,
        render: (_: unknown, domain: CustomDomain) => (
          <span className="text-sm text-muted-foreground">
            {format(new Date(domain.created_at), 'd MMM yy HH:mm:ss')}
          </span>
        ),
      },
      {
        key: 'actions',
        title: '',
        render: (_: unknown, domain: CustomDomain) => {
          if (domain.type === 'system') return null;

          let visitItem: ReactNode = null;
          if (domain.status === 'active') {
            visitItem = (
              <DropdownMenuItem asChild>
                <a href={`https://${domain.name}`} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-4 w-4 mr-2" />
                  Visit
                </a>
              </DropdownMenuItem>
            );
          }

          let verifyItem: ReactNode = null;
          if (domain.status === 'pending_dns') {
            verifyItem = (
              <DropdownMenuItem onClick={() => handleVerifyDomain(domain.id)}>
                <RotateCw className="h-4 w-4 mr-2" />
                Verify DNS
              </DropdownMenuItem>
            );
          }

          return (
            <div className="flex items-center justify-end">
              <RowActionsDropdown ariaLabel="Domain actions">
                {visitItem}
                {verifyItem}
                <DropdownMenuItem
                  onClick={() => setRemoveTarget(domain)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Remove
                </DropdownMenuItem>
              </RowActionsDropdown>
            </div>
          );
        },
      },
    ],
    [handleVerifyDomain, setRemoveTarget],
  );

  let body: ReactNode;
  if (isLoading) {
    body = <TableSkeleton {...DOMAINS_TABLE_SKELETON} />;
  } else if (error) {
    body = (
      <EmptyState
        icon={Globe}
        message="Failed to load domains"
        description="Something went wrong. Please try again."
        onRetry={refetch}
        bordered={false}
      />
    );
  } else if (filteredDomains.length === 0) {
    body = (
      <EmptyState
        icon={Globe}
        message="No domains configured"
        description="Add a custom domain to use your own URL for your services."
        bordered={false}
      />
    );
  } else {
    body = (
      <DataTable
        data={filteredDomains}
        columns={columns}
        showBorder={false}
        hoverable
        onSort={handleSort}
        sortConfig={sortConfig}
        tableClassName={DATA_TABLE_CLASS}
      />
    );
  }

  return (
    <>
      <h1 className="text-2xl tracking-wide pb-4 sm:pb-6">Domains</h1>

      <div className="pb-4 pt-6 min-w-0">
        <ListToolbar
          left={
            <SearchBar
              label="Search domains"
              searchTerm={searchTerm}
              handleSearchChange={(e) => setSearchTerm(e.target.value)}
              className="w-52 sm:w-64 [&_input]:w-full!"
            />
          }
        >
          <RefreshButton onClick={refetch} isFetching={isFetching} ariaLabel="Refresh domains" />
          <Button
            onClick={() => setShowAddDialog(true)}
            size="sm"
            className="gap-2"
            disabled={isAdding}
          >
            <Plus className="h-4 w-4" />
            Add Domain
          </Button>
        </ListToolbar>
      </div>

      <TablePanel>{body}</TablePanel>

      {showAddDialog && (
        <AddDomainDialog
          open={showAddDialog}
          onOpenChange={setShowAddDialog}
          onAdd={onAdd}
          isAdding={isAdding}
        />
      )}

      {!!removeTarget && (
        <RemoveDomainDialog
          open={!!removeTarget}
          onOpenChange={onCloseRemoveDialog}
          domainName={removeTarget.name}
          onRemove={onConfirmRemove}
          isRemoving={isRemoving}
        />
      )}

      {showDnsWizard && (
        <DNSSetupWizard
          open={showDnsWizard}
          instructions={dnsInstructions}
          dnsProvider={dnsProvider}
          domainId={currentDomainId}
          onVerify={handleVerifyDomain}
          onClose={onCloseDnsWizard}
          isVerifying={isVerifying}
        />
      )}
    </>
  );
}

export const DomainsSettings = React.memo(function DomainsSettings() {
  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <CustomDomainsList />
    </PageLayout>
  );
});
