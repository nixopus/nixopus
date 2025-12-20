import React from 'react';
import { GithubRepository } from '@/redux/types/github';
import useGithubRepoPagination from '../../hooks/use_github_repo_pagination';
import GithubRepositories, { GithubRepositoriesSkeletonLoader } from './repository-card';
import PaginationWrapper from '@/components/ui/pagination';
import { SearchBar } from '@/components/ui/search-bar';
import { SortSelect } from '@/components/ui/sort-selector';
import { useTranslation } from '@/hooks/use-translation';
import { Button } from '@/components/ui/button';
import { Settings } from 'lucide-react';
import GitHubConnectorSettingsModal from '../github-connector/github-connector-settings-modal';
import useGithubConnectorSettings from '../../hooks/use-github-connector-settings';
import { useRouter } from 'next/navigation';

function ListRepositories() {
  const { t } = useTranslation();
  const router = useRouter();
  const {
    isLoading,
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
    if (isLoading) {
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
        onOpenChange={closeSettingsModal}
        onAddNew={handleAddNewConnector}
      />
    </div>
  );
}

export default ListRepositories;
