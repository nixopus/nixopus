import React from 'react';
import { GithubRepository } from '@/redux/types/github';
import useGithubRepoPagination from '../hooks/applications/use_github_repo_pagination';
import PaginationWrapper from '@/components/ui/pagination';
import { SearchBar } from '@/components/ui/search-bar';
import { SortSelect } from '@/components/ui/sort-selector';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button } from '@/components/ui/button';
import { Settings } from 'lucide-react';
import useGithubConnectorSettings from '../hooks/applications/use-github-connector-settings';
import { useRouter } from 'next/navigation';
import { Skeleton } from '@/components/ui/skeleton';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { CardTitle } from '@/components/ui/card';
import { CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
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
    router.push('/self-host?github_setup=true');
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
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
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
        <div className="flex-1 min-w-[220px]">
          <SearchBar
            searchTerm={searchTerm}
            handleSearchChange={handleSearchChange}
            label={t('selfHost.repositories.search.placeholder')}
            isLoading={isSearching}
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
    <CardWrapper
      className="group relative w-full max-w-md cursor-pointer overflow-hidden transition-all duration-300 hover:bg-muted hover:shadow-lg"
      onClick={handleClick}
      header={
        <>
          <CardTitle className="flex items-center gap-2 text-lg font-bold">
            <Github className="text-primary" size={24} />
            {displayName}
            {url && (
              <a
                href={url}
                target="_blank"
                rel="noopener noreferrer"
                className="ml-auto text-muted-foreground transition-colors duration-200 hover:text-primary"
                title={t('selfHost.repositoryCard.viewOnGithub')}
                onClick={(e) => e.stopPropagation()}
              >
                <ExternalLink size={18} />
              </a>
            )}
          </CardTitle>
          {description && (
            <CardDescription className="line-clamp-2 text-sm text-muted-foreground">
              {description}
            </CardDescription>
          )}
        </>
      }
      headerClassName="pb-2"
      contentClassName="space-y-4"
    >
      <div className="flex flex-wrap items-center gap-2">
        <Badge variant={visibilityBadge.variant} className="text-xs font-medium">
          <visibilityBadge.icon size={12} className="mr-1" />
          {visibilityBadge.label}
        </Badge>
        {licenseBadge && (
          <Badge variant="outline" className="text-xs font-medium">
            {licenseBadge.label}
          </Badge>
        )}
      </div>
      <div className="flex items-center gap-4 text-sm text-muted-foreground">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <div key={stat.key} className="flex items-center gap-1">
              <Icon size={16} />
              <span>{stat.value}</span>
            </div>
          );
        })}
      </div>
      {displayTopics && (
        <div className="flex flex-wrap items-center gap-2">
          {displayTopics.visible.map((topic) => (
            <Badge key={topic} variant="secondary" className="text-xs font-medium">
              {topic}
            </Badge>
          ))}
          {displayTopics.remainingCount > 0 && (
            <Badge variant="secondary" className="text-xs font-medium">
              {t('selfHost.repositoryCard.topics.more').replace(
                '{count}',
                displayTopics.remainingCount.toString()
              )}
            </Badge>
          )}
        </div>
      )}
    </CardWrapper>
  );
};

export const GithubRepositoriesSkeletonLoader: React.FC = () => {
  const skeletonItems = Array.from({ length: 6 }, (_, i) => i);

  return (
    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-3">
      {skeletonItems.map((index) => (
        <CardWrapper
          key={index}
          className="group relative w-full max-w-md overflow-hidden transition-all duration-300 hover:bg-muted hover:shadow-lg"
          header={
            <>
              <CardTitle className="flex items-center gap-2 text-lg font-bold">
                <Skeleton className="h-6 w-6 rounded-full" />
                <Skeleton className="h-6 w-40" />
                <Skeleton className="ml-auto h-6 w-6 rounded-full" />
              </CardTitle>
              <CardDescription>
                <Skeleton className="mt-2 h-4 w-full" />
                <Skeleton className="mt-1 h-4 w-3/4" />
              </CardDescription>
            </>
          }
          headerClassName="pb-2"
          contentClassName="space-y-4"
        >
          <div className="flex flex-wrap items-center gap-2">
            <Skeleton className="h-5 w-16 rounded-full" />
            <Skeleton className="h-5 w-20 rounded-full" />
          </div>
          <div className="flex items-center gap-4 text-sm">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-4 w-12" />
            ))}
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} className="h-5 w-16 rounded-full" />
            ))}
          </div>
        </CardWrapper>
      ))}
    </div>
  );
};
