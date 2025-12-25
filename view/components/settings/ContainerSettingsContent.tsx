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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';

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

export function ContainerSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  if (isLoading) {
    return <ContainerSettingsSkeleton />;
  }

  const hasChanges =
    settings.containerLogTailLines !== DEFAULT_SETTINGS.containerLogTailLines ||
    settings.containerDefaultRestartPolicy !== DEFAULT_SETTINGS.containerDefaultRestartPolicy ||
    settings.containerStopTimeout !== DEFAULT_SETTINGS.containerStopTimeout ||
    settings.containerAutoPruneDanglingImages !==
      DEFAULT_SETTINGS.containerAutoPruneDanglingImages ||
    settings.containerAutoPruneBuildCache !== DEFAULT_SETTINGS.containerAutoPruneBuildCache;

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.container.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.container.description')}
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
              htmlFor="containerLogTailLines"
              title={t('settings.container.logTailLines.title')}
              description={t('settings.container.logTailLines.description')}
            />
            <NumberInput
              id="containerLogTailLines"
              value={settings.containerLogTailLines}
              onChange={(val) => updateSetting('containerLogTailLines', val)}
              min={50}
              max={10000}
              step={50}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="containerDefaultRestartPolicy"
              title={t('settings.container.defaultRestartPolicy.title')}
              description={t('settings.container.defaultRestartPolicy.description')}
            />
            <Select
              value={settings.containerDefaultRestartPolicy}
              onValueChange={(value) =>
                updateSetting(
                  'containerDefaultRestartPolicy',
                  value as 'no' | 'always' | 'on-failure' | 'unless-stopped'
                )
              }
            >
              <SelectTrigger className="w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="no">
                  {t('settings.container.defaultRestartPolicy.options.no')}
                </SelectItem>
                <SelectItem value="always">
                  {t('settings.container.defaultRestartPolicy.options.always')}
                </SelectItem>
                <SelectItem value="on-failure">
                  {t('settings.container.defaultRestartPolicy.options.onFailure')}
                </SelectItem>
                <SelectItem value="unless-stopped">
                  {t('settings.container.defaultRestartPolicy.options.unlessStopped')}
                </SelectItem>
              </SelectContent>
            </Select>
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="containerStopTimeout"
              title={t('settings.container.stopTimeout.title')}
              description={t('settings.container.stopTimeout.description')}
            />
            <NumberInput
              id="containerStopTimeout"
              value={settings.containerStopTimeout}
              onChange={(val) => updateSetting('containerStopTimeout', val)}
              min={1}
              max={300}
              suffix="s"
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="containerAutoPruneDanglingImages"
              title={t('settings.container.autoPruneDanglingImages.title')}
              description={t('settings.container.autoPruneDanglingImages.description')}
            />
            <Switch
              id="containerAutoPruneDanglingImages"
              checked={settings.containerAutoPruneDanglingImages}
              onCheckedChange={(checked) =>
                updateSetting('containerAutoPruneDanglingImages', checked)
              }
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="containerAutoPruneBuildCache"
              title={t('settings.container.autoPruneBuildCache.title')}
              description={t('settings.container.autoPruneBuildCache.description')}
            />
            <Switch
              id="containerAutoPruneBuildCache"
              checked={settings.containerAutoPruneBuildCache}
              onCheckedChange={(checked) => updateSetting('containerAutoPruneBuildCache', checked)}
            />
          </SettingRow>
        </div>
      </div>
    </div>
  );
}

function ContainerSettingsSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-4 w-72" />
      <Skeleton className="h-48 w-full" />
    </div>
  );
}
