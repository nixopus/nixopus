'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import PageLayout from '@/components/layout/page-layout';
import ExtensionsHeader from '@/app/extensions/components/extensions-header';
import ExtensionsGrid from '@/app/extensions/components/extensions-grid';
import ExtensionsHero from '@/app/extensions/components/extensions-hero';
import CategoryBadges from '@/app/extensions/components/category-badges';
import { useExtensions } from './hooks/use-extensions';
import PaginationWrapper from '@/components/ui/pagination';
import ExtensionInput from '@/app/extensions/components/extension-input';

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
    handleRun,
    runModalOpen,
    setRunModalOpen,
    selectedExtension
  } = useExtensions();

  return (
    <>
      <PageLayout maxWidth="7xl" padding="md" spacing="lg">
        <ExtensionsHero isLoading={isLoading} />
        <ExtensionsHeader
          searchTerm={searchTerm}
          onSearchChange={handleSearchChange}
          sortConfig={sortConfig}
          onSortChange={handleSortChange}
          isLoading={isLoading}
        />

        <div className="space-y-6">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <CategoryBadges
                categories={categories}
                selected={selectedCategory}
                onChange={handleCategoryChange}
              />
            </div>
            {totalExtensions > 0 && (
              <div className="text-sm text-muted-foreground whitespace-nowrap">
                Showing {extensions.length} of {totalExtensions} extensions
              </div>
            )}
          </div>

          <ExtensionsGrid
            extensions={extensions}
            isLoading={isLoading}
            error={error || undefined}
            onInstall={handleInstall}
            onViewDetails={handleViewDetails}
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
      </PageLayout>
      <ExtensionInput
        open={runModalOpen}
        onOpenChange={setRunModalOpen}
        extension={selectedExtension}
        onSubmit={handleRun}
      />
    </>
  );
}
