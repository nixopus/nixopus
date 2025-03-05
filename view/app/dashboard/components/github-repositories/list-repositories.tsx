import React from 'react';
import { GithubRepository } from '@/redux/types/github';
import useGithubRepoPagination from '../../hooks/use_github_repo_pagination';
import { DahboardUtilityHeader } from '@/components/dashboard-page-header';
import GithubRepositories, { GithubRepositoriesSkeletonLoader } from './repository-card';
import PaginationWrapper from '@/components/pagination';

function ListRepositories() {
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
    paginatedApplications
  } = useGithubRepoPagination();

  const renderGithubRepositories = () => {
    if (isLoading) {
      return <GithubRepositoriesSkeletonLoader />;
    }

    if (paginatedApplications?.length === 0 && !isLoading) {
      return <div className="text-center">No repositories found</div>;
    }
    return (
      <>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {paginatedApplications &&
            paginatedApplications?.map((repo: any) => (
              <GithubRepositories
                key={repo.id}
                {...repo}
                setSelectedRepository={setSelectedRepository}
              />
            ))}
        </div>
        <div className="mt-8 flex justify-center">
          <PaginationWrapper
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </div>
      </>
    );
  };

  return (
    <div>
      <DahboardUtilityHeader<GithubRepository>
        searchTerm={searchTerm}
        handleSearchChange={handleSearchChange}
        sortConfig={sortConfig}
        onSortChange={onSortChange}
        sortOptions={sortOptions}
        label="Github Repositories"
        className="mt-5 mb-5"
      />
      {renderGithubRepositories()}
    </div>
  );
}

export default ListRepositories;
