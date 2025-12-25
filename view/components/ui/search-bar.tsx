import React from 'react';
import { Search, Loader2 } from 'lucide-react';
import { Input } from '@/components/ui/input';

interface SearchBarProps {
  searchTerm: string;
  handleSearchChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  label?: string;
  isLoading?: boolean;
}

export const SearchBar: React.FC<SearchBarProps> = ({
  searchTerm,
  handleSearchChange,
  label = 'Search...',
  isLoading = false
}) => (
  <div className="relative w-full sm:w-64">
    <Search className="absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 transform text-muted-foreground" />
    <Input
      type="text"
      placeholder={label}
      value={searchTerm}
      onChange={handleSearchChange}
      className="w-full pl-8 sm:w-64"
    />
    {isLoading && (
      <Loader2 className="absolute right-2 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin text-muted-foreground" />
    )}
  </div>
);
