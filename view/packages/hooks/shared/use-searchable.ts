import { useState, useMemo } from 'react';

type SortDirection = 'asc' | 'desc';

export interface SortConfig<T> {
  key: keyof T;
  direction: SortDirection;
}

/**
 * Converts a value to a string for search/sort operations.
 * Handles arrays by extracting string values (e.g., ApplicationDomain[] -> domain strings).
 */
function valueToString(value: any): string {
  if (value === null || value === undefined) {
    return '';
  }

  // Handle arrays
  if (Array.isArray(value)) {
    if (value.length === 0) {
      return '';
    }

    // Extract string values from array elements
    const stringValues = value
      .map((item) => {
        // If item is an object with a 'domain' property (e.g., ApplicationDomain)
        if (item && typeof item === 'object' && 'domain' in item) {
          return String(item.domain || '');
        }
        // If item is already a string or can be converted to string
        return String(item || '');
      })
      .filter((str) => str.trim() !== ''); // Filter out empty strings

    return stringValues.join(' '); // Join with space for search/sort
  }

  // Handle non-array values
  return String(value);
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
      searchKeys.some((key) =>
        valueToString(item[key]).toLowerCase().includes(searchTerm.toLowerCase())
      )
    );

    result.sort((a, b) => {
      const aValue = valueToString(a[sortConfig.key]);
      const bValue = valueToString(b[sortConfig.key]);
      if (aValue < bValue) {
        return sortConfig.direction === 'asc' ? -1 : 1;
      }
      if (aValue > bValue) {
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
