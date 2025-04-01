import React from 'react';
import { Search } from 'lucide-react';
import { Input } from '@/components/ui/input';

interface SearchBarProps {
  searchTerm: string;
  handleSearchChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  label?: string;
}

export const SearchBar: React.FC<SearchBarProps> = ({
  searchTerm,
  handleSearchChange,
  label = 'Search...'
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
  </div>
);
