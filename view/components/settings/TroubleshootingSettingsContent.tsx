'use client';

import { useAdvancedSettings } from '@/app/settings/hooks/use-advanced-settings';
import { useTranslation } from '@/hooks/use-translation';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { TypographyMuted } from '@/components/ui/typography';
import { RotateCcw } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

function SettingRow({ children, className }: { children: React.ReactNode; className?: string }) {
  return <div className={cn('flex items-center justify-between py-3', className)}>{children}</div>;
}

function SettingLabel({
  htmlFor,
  title,
  description
}: {
  htmlFor: string;
  title: string;
  description: string;
}) {
  return (
    <div className="space-y-0.5">
      <Label htmlFor={htmlFor} className="text-sm font-medium">
        {title}
      </Label>
      <TypographyMuted className="text-xs">{description}</TypographyMuted>
    </div>
  );
}

export function TroubleshootingSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  if (isLoading) {
    return <TroubleshootingSettingsSkeleton />;
  }

  const hasChanges =
    settings.debugMode !== DEFAULT_SETTINGS.debugMode ||
    settings.showApiErrorDetails !== DEFAULT_SETTINGS.showApiErrorDetails;

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.troubleshooting.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.troubleshooting.description')}
            </TypographyMuted>
          </div>
          {hasChanges && (
            <Button variant="outline" size="sm" onClick={resetToDefaults} className="gap-2">
              <RotateCcw className="h-4 w-4" />
              {t('settings.advanced.resetDefaults')}
            </Button>
          )}
        </div>

        <div className="space-y-1">
          <SettingRow>
            <SettingLabel
              htmlFor="debugMode"
              title={t('settings.troubleshooting.debugMode.title')}
              description={t('settings.troubleshooting.debugMode.description')}
            />
            <Switch
              id="debugMode"
              checked={settings.debugMode}
              onCheckedChange={(val) => updateSetting('debugMode', val)}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="showApiErrorDetails"
              title={t('settings.troubleshooting.showApiErrorDetails.title')}
              description={t('settings.troubleshooting.showApiErrorDetails.description')}
            />
            <Switch
              id="showApiErrorDetails"
              checked={settings.showApiErrorDetails}
              onCheckedChange={(val) => updateSetting('showApiErrorDetails', val)}
            />
          </SettingRow>
        </div>
      </div>
    </div>
  );
}

function TroubleshootingSettingsSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-4 w-72" />
      <Skeleton className="h-48 w-full" />
    </div>
  );
}
