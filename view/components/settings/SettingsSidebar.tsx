'use client';

import { Settings } from 'lucide-react';
import { cn } from '@/lib/utils';
import { SettingsFooter } from './SettingsFooter';

interface Category {
  id: string;
  label: string;
  icon: typeof Settings;
  visible?: boolean;
}

interface SettingsSidebarProps {
  categories: Category[];
  activeCategory: string;
  onCategoryChange: (category: string) => void;
}

export function SettingsSidebar({
  categories,
  activeCategory,
  onCategoryChange
}: SettingsSidebarProps) {
  return (
    <div className="w-[240px] flex-shrink-0 bg-muted/50 border-r flex flex-col">
      <div className="flex-1 overflow-y-auto p-4 space-y-1">
        {categories
          .filter((cat) => cat.visible !== false)
          .map((cat) => {
            const Icon = cat.icon;
            return (
              <button
                key={cat.id}
                onClick={() => onCategoryChange(cat.id)}
                className={cn(
                  'w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors',
                  activeCategory === cat.id ? 'bg-muted font-medium' : 'hover:bg-muted/50'
                )}
              >
                <Icon className="h-4 w-4" />
                <span>{cat.label}</span>
              </button>
            );
          })}
      </div>
      <SettingsFooter />
    </div>
  );
}
