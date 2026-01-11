'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import PageLayout from '@/packages/layouts/page-layout';
import { useExtensions } from '../../packages/hooks/extensions/use-extensions';
import PaginationWrapper from '@/components/ui/pagination';
import MainPageHeader from '@/components/ui/main-page-header';
import { SearchBar } from '@/components/ui/search-bar';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { Skeleton } from '@/components/ui/skeleton';
import { ExtensionSortField, SortDirection } from '@/redux/types/extension';
import { ExtensionGrid, ExtensionInput } from '@/packages/components/extension';
import CategoryBadges from '@/packages/components/extension';

export default function ExtensionsPage() {
  const { t } = useTranslation();
  const {
    extensions,
    isLoading,
    error,
    categories,
    searchTerm,
    sortConfig,
    currentPage,
    totalPages,
    totalExtensions,
    handleSearchChange,
    handleSortChange,
    selectedCategory,
    handleCategoryChange,
    handlePageChange,
    handleInstall,
    handleViewDetails,
    handleForkClick,
    handleRun,
    runModalOpen,
    setRunModalOpen,
    selectedExtension,
    sortOptions,
    forkOpen,
    setForkOpen,
    confirmOpen,
    setConfirmOpen,
    expanded,
    setExpanded,
    forkYaml,
    setForkYaml,
    preview,
    variableColumns,
    doFork,
    actions,
    isOnlyProxyDomain,
    noFieldsToShow,
    values,
    errors,
    handleChange,
    handleSubmit,
    requiredFields
  } = useExtensions();

  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <MainPageHeader
        label={t('extensions.title')}
        actions={
          <div className="flex items-center gap-12">
            <SearchBar
              searchTerm={searchTerm}
              handleSearchChange={(e) => handleSearchChange(e.target.value)}
              label={t('extensions.searchPlaceholder')}
              className="w-full sm:w-[300px]"
            />
            <SelectWrapper
              value={sortConfig ? `${sortConfig.key}_${sortConfig.direction}` : 'name_asc'}
              onValueChange={(value) => {
                const [key, direction] = value.split('_') as [ExtensionSortField, SortDirection];
                handleSortChange(key, direction);
              }}
              options={sortOptions}
              placeholder={t('extensions.sortBy')}
              className="w-full sm:w-[180px]"
            />
          </div>
        }
      />
      <div className="space-y-6">
        <div className="flex items-start justify-between gap-4">
          <CategoryBadges
            categories={categories}
            selected={selectedCategory}
            onChange={handleCategoryChange}
          />
          {totalExtensions > 0 && (
            <div className="text-sm text-muted-foreground whitespace-nowrap">
              Showing {extensions.length} of {totalExtensions} extensions
            </div>
          )}
        </div>

        <ExtensionGrid
          extensions={extensions}
          isLoading={isLoading}
          error={error || undefined}
          onInstall={handleInstall}
          onViewDetails={handleViewDetails}
          onForkClick={handleForkClick}
          setConfirmOpen={setConfirmOpen}
          expanded={expanded}
          setExpanded={setExpanded}
          forkOpen={forkOpen}
          setForkOpen={setForkOpen}
          confirmOpen={confirmOpen}
          forkYaml={forkYaml}
          setForkYaml={setForkYaml}
          preview={preview}
          variableColumns={variableColumns}
          doFork={doFork}
          selectedExtension={selectedExtension}
        />

        {totalPages > 1 && (
          <div className="flex justify-center pt-6">
            <PaginationWrapper
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={handlePageChange}
            />
          </div>
        )}
      </div>
      <ExtensionInput
        open={runModalOpen}
        onOpenChange={setRunModalOpen}
        extension={selectedExtension}
        onSubmit={handleRun}
        t={t}
        isOnlyProxyDomain={isOnlyProxyDomain}
        noFieldsToShow={noFieldsToShow}
        values={values}
        errors={errors}
        handleChange={handleChange}
        handleSubmit={handleSubmit}
        requiredFields={requiredFields}
        actions={actions}
      />
    </PageLayout>
  );
}
