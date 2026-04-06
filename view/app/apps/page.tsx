'use client';
import React, { useState, useMemo } from 'react';
import { formatDistanceToNow } from 'date-fns';
import GitHubAppSetup, { ManagedGitHubAppSetup } from '@/packages/components/github-connector';
import { ListRepositories } from '@/packages/components/github-repositories';
import AppItem from '../../packages/components/application';
import useGetDeployedApplications from '../../packages/hooks/applications/use_get_deployed_applications';
import {
  PaginationWrapper,
  SearchBar,
  Button,
  Badge,
  DataTable,
  TypographyH2,
  TypographyMuted,
  Skeleton,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@nixopus/ui';
import { SortSelect } from '@/components/ui/sort-selector';
import { Application } from '@/redux/types/applications';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';
import { Plus, MoveUpRight, MoreVertical } from 'lucide-react';
import { LabelFilter } from '@/components/ui/label-filter';
import { SSHBanner } from '@/packages/components/dashboard';
import { cn } from '@/lib/utils';
import { useRouter } from 'next/navigation';
import {
  DATA_TABLE_CLASS,
  LIST_GRID_CLASS,
  ViewToggle,
  RefreshButton,
  ListToolbar,
  CardSkeleton
} from '@/components/ui/list-page-chrome';

function AppsPageSkeleton() {
  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <Skeleton className="h-8 w-48" />
      <div className="flex items-center justify-between flex-wrap gap-3">
        <Skeleton className="h-10 w-64 rounded-md" />
        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-[68px] rounded-md" />
          <Skeleton className="h-8 w-8 rounded-md" />
          <Skeleton className="h-9 w-32 rounded-md" />
        </div>
      </div>
      <CardSkeleton />
    </PageLayout>
  );
}

const STATUS_DOT: Record<string, string> = {
  deployed: 'bg-foreground/60',
  running: 'bg-foreground/60',
  failed: 'bg-foreground/40',
  cancelled: 'bg-foreground/30',
  building: 'bg-foreground/50',
  deploying: 'bg-foreground/50',
  cloning: 'bg-foreground/50',
  started: 'bg-foreground/50',
  draft: 'bg-foreground/30',
  stopped: 'bg-foreground/20'
};

const STATUS_LABEL: Record<string, string> = {
  deployed: 'Live',
  running: 'Running',
  failed: 'Failed',
  cancelled: 'Cancelled',
  building: 'Building',
  deploying: 'Building',
  cloning: 'Building',
  started: 'Building',
  draft: 'Draft',
  stopped: 'Stopped'
};

const ENV_STYLES: Record<string, string> = {
  development: 'border-border text-muted-foreground bg-foreground/5',
  staging: 'border-border text-muted-foreground bg-foreground/5',
  production: 'border-border text-muted-foreground bg-foreground/5'
};

function resolveStatus(app: Application): string {
  const latestDeployment = app.deployments?.[0];
  const latestStatus = latestDeployment?.status?.status || app.status?.status;
  if (latestStatus === 'cancelled' || latestStatus === 'failed') {
    return (
      app.deployments?.find(
        (d) => d.status?.status === 'deployed' || d.status?.status === 'running'
      )?.status?.status ?? latestStatus
    );
  }
  return latestStatus || '';
}

function AppStatusCell({ app }: { app: Application }) {
  const status = resolveStatus(app);
  return (
    <span className="text-sm text-muted-foreground">{STATUS_LABEL[status] || 'Inactive'}</span>
  );
}

function AppDomainCell({ app }: { app: Application }) {
  const domain = app.domains?.[0]?.domain;
  if (!domain) return <span className="text-sm text-muted-foreground">-</span>;
  const count = app.domains?.length || 0;
  return (
    <div className="flex items-center gap-1.5 min-w-0">
      <span className="text-sm text-muted-foreground truncate">{domain}</span>
      {count > 1 && (
        <Badge variant="outline" className="rounded-sm text-[10px] px-1.5 shrink-0">
          +{count - 1}
        </Badge>
      )}
    </div>
  );
}

function AppActionsCell({ app }: { app: Application }) {
  const domains = app.domains;
  if (!domains || domains.length === 0) return null;

  return (
    <div className="flex items-center justify-end">
      {domains.length === 1 ? (
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0"
          onClick={(e) => {
            e.stopPropagation();
            const url = domains[0].domain.startsWith('http')
              ? domains[0].domain
              : `https://${domains[0].domain}`;
            window.open(url, '_blank', 'noopener,noreferrer');
          }}
        >
          <MoveUpRight className="h-4 w-4" />
        </Button>
      ) : (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" onClick={(e) => e.stopPropagation()}>
            {domains.map((d, i) => (
              <DropdownMenuItem
                key={i}
                onClick={() => {
                  const url = d.domain.startsWith('http') ? d.domain : `https://${d.domain}`;
                  window.open(url, '_blank', 'noopener,noreferrer');
                }}
              >
                {d.domain}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      )}
    </div>
  );
}

function useAppTableColumns() {
  const router = useRouter();

  return useMemo(
    () => [
      {
        key: 'name',
        title: 'Name',
        className: 'w-[30%]',
        sortable: true,
        render: (_: any, app: Application) => (
          <span className="text-sm font-medium text-foreground truncate">{app.name}</span>
        )
      },
      {
        key: 'status',
        title: 'Status',
        render: (_: any, app: Application) => <AppStatusCell app={app} />
      },
      {
        key: 'environment',
        title: 'Environment',
        sortable: true,
        render: (_: any, app: Application) => (
          <Badge
            variant="outline"
            className={cn(
              'text-xs font-medium capitalize rounded-full px-3 py-0.5 border',
              ENV_STYLES[app.environment] || 'border-border text-muted-foreground bg-foreground/5'
            )}
          >
            {app.environment}
          </Badge>
        )
      },
      {
        key: 'updated_at',
        title: 'Deployed',
        sortable: true,
        render: (_: any, app: Application) => (
          <span className="text-sm text-muted-foreground">
            {app.updated_at
              ? formatDistanceToNow(new Date(app.updated_at), { addSuffix: true })
              : '-'}
          </span>
        )
      },
      {
        key: 'domains',
        title: 'Domain',
        render: (_: any, app: Application) => <AppDomainCell app={app} />
      },
      {
        key: 'actions',
        title: '',
        align: 'right' as const,
        render: (_: any, app: Application) => <AppActionsCell app={app} />
      }
    ],
    [router]
  );
}

function AppsPage() {
  const { t } = useTranslation();
  const router = useRouter();
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid');
  const {
    connectors,
    GetGithubConnectors,
    isLoading,
    applications,
    GetApplications,
    isLoadingApplications,
    isFetchingApplications,
    searchTerm,
    handleSearchChange,
    onSortChange,
    sortOptions,
    sortConfig,
    handlePageChange,
    currentPage,
    totalPages,
    inGitHubFlow,
    showApplications,
    labelFilter,
    selfHosted
  } = useGetDeployedApplications();

  const columns = useAppTableColumns();

  if (selfHosted === null || isLoadingApplications || isLoading) {
    return <AppsPageSkeleton />;
  }

  const isManagedMode = !selfHosted;
  const isShowingGitHubSetup = isManagedMode
    ? inGitHubFlow || (!showApplications && !connectors?.length)
    : inGitHubFlow || (!showApplications && !inGitHubFlow && !connectors?.length);
  const isShowingRepositories =
    !showApplications && !inGitHubFlow && connectors?.length && connectors.length > 0;

  const renderContent = () => (
    <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
      {isManagedMode ? (
        <>
          {(inGitHubFlow || (!showApplications && !connectors?.length)) && (
            <ManagedGitHubAppSetup />
          )}
          {!inGitHubFlow && !showApplications && connectors?.length && connectors.length > 0 && (
            <ListRepositories />
          )}
        </>
      ) : (
        <>
          {inGitHubFlow && <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />}
          {!showApplications && !inGitHubFlow && (
            <>
              {!connectors?.length ? (
                <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />
              ) : (
                <ListRepositories />
              )}
            </>
          )}
        </>
      )}
    </AnyPermissionGuard>
  );

  return (
    <ResourceGuard
      resource="deploy"
      action="read"
      loadingFallback={<AppsPageSkeleton />}
      fallback={
        <div className="flex h-full items-center justify-center">
          <div className="text-center">
            <TypographyH2>{t('selfHost.page.accessDenied.title')}</TypographyH2>
            <TypographyMuted>{t('selfHost.page.accessDenied.description')}</TypographyMuted>
          </div>
        </div>
      }
    >
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        {!isShowingGitHubSetup && (
          <h1 className="text-2xl tracking-wide">
            {isShowingRepositories ? t('selfHost.repositories.title') : t('selfHost.page.title')}
          </h1>
        )}

        <SSHBanner />

        {renderContent()}

        {showApplications && !inGitHubFlow && (
          <>
            <ListToolbar
              left={
                <>
                  <SearchBar
                    searchTerm={searchTerm}
                    handleSearchChange={handleSearchChange}
                    label={t('selfHost.page.search.placeholder')}
                    className="w-full max-w-xs [&_input]:w-full!"
                  />
                  <SortSelect<Application>
                    options={sortOptions}
                    currentSort={{
                      value: sortConfig.key,
                      direction: sortConfig.direction,
                      label:
                        sortOptions.find(
                          (o) => o.value === sortConfig.key && o.direction === sortConfig.direction
                        )?.label || ''
                    }}
                    onSortChange={onSortChange}
                    placeholder="Sort by"
                    className="w-[160px]"
                  />
                </>
              }
            >
              <ViewToggle viewMode={viewMode} onViewChange={setViewMode} />
              <RefreshButton
                onClick={() => GetApplications()}
                isFetching={isFetchingApplications}
              />
              <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
                <Button onClick={() => router.push('/apps/create')} size="sm" className="gap-2">
                  <Plus className="h-4 w-4" />
                  {t('selfHost.page.createButton')}
                </Button>
              </AnyPermissionGuard>
            </ListToolbar>

            {labelFilter.availableLabels.length > 0 && (
              <LabelFilter
                availableLabels={labelFilter.availableLabels}
                selectedLabels={labelFilter.selectedLabels}
                onToggle={labelFilter.toggleLabel}
                onClear={labelFilter.clearFilters}
                className="pb-4"
              />
            )}

            {viewMode === 'grid' ? (
              <div className={LIST_GRID_CLASS}>
                {applications && applications.map((app: any) => <AppItem key={app.id} {...app} />)}
              </div>
            ) : (
              <div className="overflow-x-auto rounded-md border">
                <DataTable
                  data={applications || []}
                  columns={columns}
                  showBorder={false}
                  hoverable
                  onRowClick={(app) => router.push(`/apps/application/${app.id}`)}
                  tableClassName={DATA_TABLE_CLASS}
                  onSort={(field) =>
                    onSortChange({ value: field as keyof Application, label: '', direction: 'asc' })
                  }
                  sortConfig={{ field: String(sortConfig.key), order: sortConfig.direction }}
                />
              </div>
            )}

            {totalPages > 1 && (
              <div className="mt-6 flex justify-center">
                <PaginationWrapper
                  currentPage={currentPage}
                  totalPages={totalPages}
                  onPageChange={handlePageChange}
                />
              </div>
            )}
          </>
        )}
      </PageLayout>
    </ResourceGuard>
  );
}

export default AppsPage;
