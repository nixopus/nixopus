import { useState, useMemo, useCallback } from 'react';

const DEFAULT_PAGE_SIZE = 10;

export function usePagination<T>(items: T[], pageSize = DEFAULT_PAGE_SIZE) {
  const [currentPage, setCurrentPage] = useState(1);

  const totalPages = Math.ceil(items.length / pageSize);
  const safeCurrentPage = Math.min(currentPage, Math.max(totalPages, 1));

  const paginatedItems = useMemo(() => {
    const start = (safeCurrentPage - 1) * pageSize;
    return items.slice(start, start + pageSize);
  }, [items, safeCurrentPage, pageSize]);

  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
  }, []);

  if (currentPage !== safeCurrentPage) {
    setCurrentPage(safeCurrentPage);
  }

  return {
    paginatedItems,
    currentPage: safeCurrentPage,
    totalPages,
    handlePageChange,
  };
}
