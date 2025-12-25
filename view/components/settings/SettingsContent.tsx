'use client';

import { GeneralSettingsContent } from './GeneralSettingsContent';
import { NotificationsSettingsContent } from './NotificationsSettingsContent';
import { TeamsSettingsContent } from './TeamsSettingsContent';
import { DomainsSettingsContent } from './DomainsSettingsContent';
import { FeatureFlagsSettingsContent } from './FeatureFlagsSettingsContent';
import { KeyboardShortcutsSettingsContent } from './KeyboardShortcutsSettingsContent';
import { NetworkSettingsContent } from './NetworkSettingsContent';
import { TerminalSettingsContent } from './TerminalSettingsContent';
import { ContainerSettingsContent } from './ContainerSettingsContent';
import { TroubleshootingSettingsContent } from './TroubleshootingSettingsContent';

interface SettingsContentProps {
  activeCategory: string;
}

export function SettingsContent({ activeCategory }: SettingsContentProps) {
  return (
    <div className="flex-1 flex flex-col overflow-hidden p-6">
      {activeCategory === 'general' && <GeneralSettingsContent />}
      {activeCategory === 'notifications' && <NotificationsSettingsContent />}
      {activeCategory === 'teams' && <TeamsSettingsContent />}
      {activeCategory === 'domains' && <DomainsSettingsContent />}
      {activeCategory === 'feature-flags' && <FeatureFlagsSettingsContent />}
      {activeCategory === 'keyboard-shortcuts' && <KeyboardShortcutsSettingsContent />}
      {activeCategory === 'network' && <NetworkSettingsContent />}
      {activeCategory === 'terminal' && <TerminalSettingsContent />}
      {activeCategory === 'container' && <ContainerSettingsContent />}
      {activeCategory === 'troubleshooting' && <TroubleshootingSettingsContent />}
    </div>
  );
}
