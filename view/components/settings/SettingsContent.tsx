'use client';

import { GeneralSettingsContent } from './GeneralSettingsContent';
import { NotificationsSettingsContent } from './NotificationsSettingsContent';
import { TeamsSettingsContent } from './TeamsSettingsContent';
import { DomainsSettingsContent } from './DomainsSettingsContent';
import { FeatureFlagsSettingsContent } from './FeatureFlagsSettingsContent';
import { KeyboardShortcutsSettingsContent } from './KeyboardShortcutsSettingsContent';

interface SettingsContentProps {
  activeCategory: string;
}

export function SettingsContent({ activeCategory }: SettingsContentProps) {
  return (
    <div className="flex-1 overflow-y-auto p-6">
      {activeCategory === 'general' && <GeneralSettingsContent />}
      {activeCategory === 'notifications' && <NotificationsSettingsContent />}
      {activeCategory === 'teams' && <TeamsSettingsContent />}
      {activeCategory === 'domains' && <DomainsSettingsContent />}
      {activeCategory === 'feature-flags' && <FeatureFlagsSettingsContent />}
      {activeCategory === 'keyboard-shortcuts' && <KeyboardShortcutsSettingsContent />}
    </div>
  );
}
