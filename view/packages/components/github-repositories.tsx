import React from 'react';
import { GithubRepository } from '@/redux/types/github';
import useGithubRepoPagination from '../hooks/applications/use_github_repo_pagination';
import { PaginationWrapper } from '@nixopus/ui';
import { SearchBar } from '@nixopus/ui';
import { SortSelect } from '@/components/ui/sort-selector';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button } from '@nixopus/ui';
import { Settings } from 'lucide-react';
import useGithubConnectorSettings from '../hooks/applications/use-github-connector-settings';
import { useRouter } from 'next/navigation';
import { Skeleton } from '@nixopus/ui';
import { Card, CardContent } from '@nixopus/ui';
import { Badge } from '@nixopus/ui';
import { Github } from 'lucide-react';
import { ExternalLink } from 'lucide-react';
import { useRepositoryCard } from '@/packages/hooks/github/use_repository_card';
import { GitHubConnectorSettingsModal } from '@/packages/components/github-connector';

export function ListRepositories() {
  const { t } = useTranslation();
  const router = useRouter();
  const {
    isLoading,
    isSearching,
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
    onSelectRepository
  } = useGithubRepoPagination();
  const { isSettingsModalOpen, openSettingsModal, closeSettingsModal } =
    useGithubConnectorSettings();

  const handleAddNewConnector = () => {
    router.push('/apps?github_setup=true');
  };

  const renderGithubRepositories = () => {
    if (isSearching || isLoading) {
      return <GithubRepositoriesSkeletonLoader />;
    }

    if (paginatedApplications?.length === 0 && !isLoading) {
      return <div className="text-center">{t('selfHost.repositories.noRepositories')}</div>;
    }
    return (
      <>
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-4">
          {paginatedApplications &&
            paginatedApplications?.map((repo: any) => (
              <GithubRepositories
                key={repo.id}
                {...repo}
                setSelectedRepository={onSelectRepository}
              />
            ))}
        </div>
        {totalPages > 1 && (
          <div className="mt-8 flex justify-center">
            <PaginationWrapper
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={handlePageChange}
            />
          </div>
        )}
      </>
    );
  };

  return (
    <div>
      <div className="flex items-center justify-between gap-4 flex-wrap mb-4">
        <div className="flex-1 max-w-fit ">
          <SearchBar
            searchTerm={searchTerm}
            handleSearchChange={handleSearchChange}
            label={t('selfHost.repositories.search.placeholder')}
            isLoading={isSearching}
            className="min-w-[280px] [&_input]:pl-12 [&_svg]:ml-3 [&_svg]:left-3"
          />
        </div>
        <div className="flex items-center gap-2">
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
          />
          <Button
            variant="outline"
            size="sm"
            onClick={openSettingsModal}
            className="flex items-center gap-2"
          >
            <Settings size={16} />
            {t('selfHost.repositories.settings' as any)}
          </Button>
        </div>
      </div>
      {renderGithubRepositories()}
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

export const GithubRepositories = (
  props: GithubRepository & { setSelectedRepository: (repo: string) => void }
) => {
  const { t } = useTranslation();
  const {
    displayName,
    url,
    description,
    stats,
    visibilityBadge,
    licenseBadge,
    displayTopics,
    handleClick
  } = useRepositoryCard(props);

  return (
    <Card
      className="relative w-full cursor-pointer overflow-hidden border border-white/[0.06] transition-colors duration-200 hover:bg-muted/50"
      onClick={handleClick}
    >
      <CardContent className="p-3">
        <div className="flex flex-col gap-1">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-1.5 flex-1 min-w-0">
              <h3 className="font-semibold text-base tracking-tight truncate">{displayName}</h3>
            </div>
            {url && (
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 shrink-0 text-muted-foreground hover:text-foreground"
                onClick={(e) => {
                  e.stopPropagation();
                  window.open(url, '_blank', 'noopener,noreferrer');
                }}
                title={t('selfHost.repositoryCard.viewOnGithub')}
              >
                <ExternalLink className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export const GithubRepositoriesSkeletonLoader: React.FC = () => {
  const skeletonItems = Array.from({ length: 6 }, (_, i) => i);

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
      {skeletonItems.map((index) => (
        <Card key={index} className="relative w-full">
          <CardContent className="p-3">
            <div className="flex flex-col gap-1">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-1.5 flex-1">
                  <Skeleton className="h-5 w-5 rounded-full" />
                  <Skeleton className="h-6 w-36" />
                </div>
                <Skeleton className="h-7 w-7 rounded-md" />
              </div>
              <Skeleton className="h-4 w-full mt-0.5" />
              <Skeleton className="h-4 w-3/4 mt-0.5" />
              <div className="flex items-center gap-1.5 mt-1">
                <Skeleton className="h-6 w-18 rounded-full" />
                <Skeleton className="h-6 w-14 rounded-full" />
              </div>
              <div className="flex items-center gap-2 mt-0.5">
                {[1, 2, 3].map((i) => (
                  <Skeleton key={i} className="h-4 w-12" />
                ))}
              </div>
              <div className="flex flex-wrap items-center gap-1.5 mt-0.5">
                {[1, 2, 3].map((i) => (
                  <Skeleton key={i} className="h-6 w-16 rounded-full" />
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
};
