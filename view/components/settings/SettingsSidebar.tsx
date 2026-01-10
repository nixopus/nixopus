'use client';

import { Settings } from 'lucide-react';
import { cn } from '@/lib/utils';
import { SettingsFooter } from './SettingsFooter';
import { SettingsCategory } from '@/hooks/use-settings-categories';
import { useTranslation } from '@/hooks/use-translation';

interface SettingsSidebarProps {
  categories: SettingsCategory[];
  activeCategory: string;
  onCategoryChange: (category: string) => void;
}

export function SettingsSidebar({
  categories,
  activeCategory,
  onCategoryChange
}: SettingsSidebarProps) {
  const { t } = useTranslation();
  const visibleCategories = categories.filter((cat) => cat.visible !== false);
  const accountCategories = visibleCategories.filter((cat) => cat.scope === 'account');
  const orgCategories = visibleCategories.filter((cat) => cat.scope === 'organization');

  const renderCategoryButton = (cat: SettingsCategory) => {
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
  };

  return (
    <div className="w-[240px] flex-shrink-0 bg-muted/50 border-r flex flex-col">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {accountCategories.length > 0 && (
          <div className="space-y-1">
            <div className="px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t('settings.sidebar.account')}
            </div>
            {accountCategories.map(renderCategoryButton)}
          </div>
        )}
        {orgCategories.length > 0 && (
          <div className="space-y-1">
            <div className="px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t('settings.sidebar.organization')}
            </div>
            {orgCategories.map(renderCategoryButton)}
          </div>
        )}
      </div>
      <SettingsFooter />
    </div>
  );
}
