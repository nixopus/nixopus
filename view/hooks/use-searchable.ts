import { useState, useMemo } from 'react';

type SortDirection = 'asc' | 'desc';

export interface SortConfig<T> {
  key: keyof T;
  direction: SortDirection;
}

export function useSearchable<T>(
  data: T[],
  searchKeys: (keyof T)[],
  initialSortConfig: SortConfig<T>
) {
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<SortConfig<T>>(initialSortConfig);

  const filteredAndSortedData = useMemo(() => {
    let result = data?.filter((item) =>
      searchKeys.some((key) => String(item[key]).toLowerCase().includes(searchTerm.toLowerCase()))
    );

    result.sort((a, b) => {
      if (a[sortConfig.key] < b[sortConfig.key]) {
        return sortConfig.direction === 'asc' ? -1 : 1;
      }
      if (a[sortConfig.key] > b[sortConfig.key]) {
        return sortConfig.direction === 'asc' ? 1 : -1;
      }
      return 0;
    });

    return result;
  }, [data, searchTerm, sortConfig, searchKeys]);

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
  };

  const handleSortChange = (key: keyof T) => {
    setSortConfig((prevConfig) => ({
      key,
      direction: prevConfig.key === key && prevConfig.direction === 'asc' ? 'desc' : 'asc'
    }));
  };

  return {
    filteredAndSortedData,
    searchTerm,
    handleSearchChange,
    handleSortChange,
    sortConfig
  };
}
