'use client';

import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { SettingsSidebar } from './SettingsSidebar';
import { useSettingsModal } from '@/hooks/use-settings-modal';
import { useSettingsCategories } from '@/hooks/use-settings-categories';
import { SettingsContent } from './SettingsContent';

export function SettingsModal() {
  const { open, closeSettings, activeCategory, setActiveCategory } = useSettingsModal();
  const categories = useSettingsCategories();
  return (
    <Dialog open={open} onOpenChange={closeSettings}>
      <DialogContent className="!max-w-[1200px] w-[90vw] max-h-[90vh] h-[90vh] p-0 flex overflow-hidden">
        <DialogTitle className="sr-only">Settings</DialogTitle>
        <SettingsSidebar
          categories={categories}
          activeCategory={activeCategory}
          onCategoryChange={setActiveCategory}
        />
        <SettingsContent activeCategory={activeCategory} />
      </DialogContent>
    </Dialog>
  );
}
