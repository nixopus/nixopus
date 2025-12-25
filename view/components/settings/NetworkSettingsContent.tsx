'use client';

import { useAdvancedSettings } from '@/app/settings/hooks/use-advanced-settings';
import { useTranslation } from '@/hooks/use-translation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
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

function NumberInput({
  id,
  value,
  onChange,
  min,
  max,
  step = 1,
  suffix
}: {
  id: string;
  value: number;
  onChange: (value: number) => void;
  min: number;
  max: number;
  step?: number;
  suffix?: string;
}) {
  return (
    <div className="flex items-center gap-2">
      <Input
        id={id}
        type="number"
        value={value}
        onChange={(e) => {
          const val = parseInt(e.target.value, 10);
          if (!isNaN(val) && val >= min && val <= max) {
            onChange(val);
          }
        }}
        min={min}
        max={max}
        step={step}
        className="w-24 text-right"
      />
      {suffix && <span className="text-xs text-muted-foreground">{suffix}</span>}
    </div>
  );
}

export function NetworkSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  if (isLoading) {
    return <NetworkSettingsSkeleton />;
  }

  const hasChanges =
    settings.websocketReconnectAttempts !== DEFAULT_SETTINGS.websocketReconnectAttempts ||
    settings.websocketReconnectInterval !== DEFAULT_SETTINGS.websocketReconnectInterval ||
    settings.apiRetryAttempts !== DEFAULT_SETTINGS.apiRetryAttempts ||
    settings.disableApiCache !== DEFAULT_SETTINGS.disableApiCache;

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.network.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.network.description')}
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
              htmlFor="websocketReconnectAttempts"
              title={t('settings.network.reconnectAttempts.title')}
              description={t('settings.network.reconnectAttempts.description')}
            />
            <NumberInput
              id="websocketReconnectAttempts"
              value={settings.websocketReconnectAttempts}
              onChange={(val) => updateSetting('websocketReconnectAttempts', val)}
              min={1}
              max={20}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="websocketReconnectInterval"
              title={t('settings.network.reconnectInterval.title')}
              description={t('settings.network.reconnectInterval.description')}
            />
            <NumberInput
              id="websocketReconnectInterval"
              value={settings.websocketReconnectInterval}
              onChange={(val) => updateSetting('websocketReconnectInterval', val)}
              min={1000}
              max={30000}
              step={500}
              suffix="ms"
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="apiRetryAttempts"
              title={t('settings.network.apiRetryAttempts.title')}
              description={t('settings.network.apiRetryAttempts.description')}
            />
            <NumberInput
              id="apiRetryAttempts"
              value={settings.apiRetryAttempts}
              onChange={(val) => updateSetting('apiRetryAttempts', val)}
              min={0}
              max={5}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="disableApiCache"
              title={t('settings.network.disableApiCache.title')}
              description={t('settings.network.disableApiCache.description')}
            />
            <Switch
              id="disableApiCache"
              checked={settings.disableApiCache}
              onCheckedChange={(checked) => updateSetting('disableApiCache', checked)}
            />
          </SettingRow>
        </div>
      </div>
    </div>
  );
}

function NetworkSettingsSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-4 w-72" />
      <Skeleton className="h-48 w-full" />
    </div>
  );
}
