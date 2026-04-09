import React, { useState } from 'react';
import { GithubRepository } from '@/redux/types/github';
import useGithubRepoPagination from '../hooks/applications/use_github_repo_pagination';
import { PaginationWrapper, CardWrapper, CardTitle, DataTable } from '@nixopus/ui';
import { SearchBar } from '@nixopus/ui';
import { SortSelect } from '@/components/ui/sort-selector';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button } from '@nixopus/ui';
import { Settings, MoveUpRight, Loader2 } from 'lucide-react';
import useGithubConnectorSettings from '../hooks/applications/use-github-connector-settings';
import { useRouter } from 'next/navigation';
import { useRepositoryCard } from '@/packages/hooks/github/use_repository_card';
import { GitHubConnectorSettingsModal } from '@/packages/components/github-connector';
import { cn } from '@/lib/utils';
import {
  DATA_TABLE_CLASS,
  LIST_GRID_CLASS,
  ViewToggle,
  RefreshButton,
  ListToolbar,
  CardSkeleton,
  CARD_CLASS,
  CARD_HEADER_CLASS
} from '@/components/ui/list-page-chrome';

const REPO_TABLE_COLUMNS = [
  {
    key: 'name',
    title: 'Name',
    className: 'w-[30%]',
    sortable: true,
    render: (_: any, repo: GithubRepository) => (
      <span className="text-sm font-medium text-foreground truncate">{repo.name}</span>
    )
  },
  {
    key: 'description',
    title: 'Description',
    render: (_: any, repo: GithubRepository) => (
      <span className="text-sm text-muted-foreground truncate block max-w-[300px]">
        {repo.description || '-'}
      </span>
    )
  },
  {
    key: 'language',
    title: 'Language',
    render: (_: any, repo: GithubRepository) => (
      <span className="text-sm text-muted-foreground">{repo.language || '-'}</span>
    )
  },
  {
    key: 'default_branch',
    title: 'Branch',
    render: (_: any, repo: GithubRepository) => (
      <span className="text-sm text-muted-foreground">{repo.default_branch || '-'}</span>
    )
  },
  {
    key: 'actions',
    title: '',
    align: 'right' as const,
    render: (_: any, repo: GithubRepository) => (
      <div className="flex items-center justify-end">
        {repo.html_url && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={(e) => {
              e.stopPropagation();
              window.open(repo.html_url, '_blank', 'noopener,noreferrer');
            }}
          >
            <MoveUpRight className="h-4 w-4" />
          </Button>
        )}
      </div>
    )
  }
];

export function ListRepositories() {
  const { t } = useTranslation();
  const router = useRouter();
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid');
  const {
    isLoading,
    isSearching,
    isFetching,
    setSelectedRepository,
    searchTerm,
    handleSearchChange,
    onSortChange,
    sortOptions,
    sortConfig,
    handlePageChange,
    currentPage,
    totalPages,
    paginatedApplications,
    onSelectRepository,
    navigatingRepoId
  } = useGithubRepoPagination();
  const { isSettingsModalOpen, openSettingsModal, closeSettingsModal } =
    useGithubConnectorSettings();

  const handleAddNewConnector = () => {
    router.push('/apps?github_setup=true');
  };

  if (isLoading) {
    return <GithubRepositoriesSkeletonLoader />;
  }

  return (
    <div>
      <ListToolbar
        left={
          <>
            <SearchBar
              searchTerm={searchTerm}
              handleSearchChange={handleSearchChange}
              label={t('selfHost.repositories.search.placeholder')}
              isLoading={isSearching}
              className="w-full max-w-xs [&_input]:w-full!"
            />
            <SortSelect<GithubRepository>
              options={sortOptions}
              currentSort={{
                value: sortConfig.key,
                direction: sortConfig.direction,
                label:
                  sortOptions.find(
                    (option) =>
                      option.value === sortConfig.key && option.direction === sortConfig.direction
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
        <RefreshButton onClick={() => {}} isFetching={isFetching} />
        <Button
          variant="outline"
          size="sm"
          onClick={openSettingsModal}
          className="flex items-center gap-2"
        >
          <Settings size={16} />
          {t('selfHost.repositories.settings' as any)}
        </Button>
      </ListToolbar>

      <div className="mt-4">
        {isSearching ? (
          <GithubRepositoriesSkeletonLoader />
        ) : paginatedApplications?.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">
            {t('selfHost.repositories.noRepositories')}
          </div>
        ) : viewMode === 'grid' ? (
          <div className={LIST_GRID_CLASS}>
            {paginatedApplications?.map((repo: any) => (
              <GithubRepositoryCard
                key={repo.id}
                {...repo}
                setSelectedRepository={onSelectRepository}
                isNavigating={navigatingRepoId === repo.id.toString()}
              />
            ))}
          </div>
        ) : (
          <div className="overflow-x-auto rounded-md border">
            <DataTable
              data={paginatedApplications || []}
              columns={REPO_TABLE_COLUMNS}
              showBorder={false}
              hoverable
              onRowClick={(repo) => onSelectRepository(repo.id.toString())}
              tableClassName={DATA_TABLE_CLASS}
              onSort={(field) =>
                onSortChange({
                  value: field as keyof GithubRepository,
                  label: '',
                  direction: 'asc'
                })
              }
              sortConfig={{ field: String(sortConfig.key), order: sortConfig.direction }}
            />
          </div>
        )}
      </div>

      {totalPages > 1 && (
        <div className="mt-6 flex justify-center">
          <PaginationWrapper
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </div>
      )}

      <GitHubConnectorSettingsModal
        open={isSettingsModalOpen}
        onOpenChange={(open: boolean) => {
          if (!open) {
            closeSettingsModal();
          }
        }}
        onAddNew={handleAddNewConnector}
      />
    </div>
  );
}

export const GithubRepositoryCard = (
  props: GithubRepository & {
    setSelectedRepository: (repo: string) => void;
    isNavigating?: boolean;
  }
) => {
  const { t } = useTranslation();
  const { isNavigating } = props;
  const { displayName, url, description, handleClick } = useRepositoryCard(props);

  return (
    <CardWrapper
      className={cn(
        CARD_CLASS,
        isNavigating && 'ring-2 ring-primary/50 bg-muted/30 pointer-events-none opacity-80'
      )}
      onClick={handleClick}
      header={
        <div className="flex-1 min-w-0 w-full space-y-1.5">
          <div className="flex w-full min-w-0 items-start justify-between gap-2">
            <div className="min-w-0 flex-1 max-w-full flex items-center gap-1.5">
              {isNavigating && <Loader2 className="h-4 w-4 animate-spin text-primary shrink-0" />}
              <CardTitle
                className="text-base font-semibold wrap-break-word max-w-full"
                style={{ wordBreak: 'break-word', overflowWrap: 'break-word', maxWidth: '100%' }}
              >
                {displayName}
              </CardTitle>
            </div>
            <div className="flex items-center gap-1 ml-auto shrink-0">
              {url && !isNavigating && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 shrink-0"
                  onClick={(e) => {
                    e.stopPropagation();
                    window.open(url, '_blank', 'noopener,noreferrer');
                  }}
                  title={t('selfHost.repositoryCard.viewOnGithub')}
                >
                  <MoveUpRight className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>
          {description && (
            <span className="text-xs text-muted-foreground line-clamp-2">{description}</span>
          )}
        </div>
      }
      headerClassName={CARD_HEADER_CLASS}
      contentClassName="flex flex-col flex-1"
    >
      <div className="flex-1" />
      {props.language && (
        <div className="flex items-center gap-3 pt-1.5 border-t border-border/50">
          <span className="text-xs text-muted-foreground">{props.language}</span>
        </div>
      )}
    </CardWrapper>
  );
};

export { GithubRepositoryCard as GithubRepositories };

export const GithubRepositoriesSkeletonLoader: React.FC = () => <CardSkeleton />;
