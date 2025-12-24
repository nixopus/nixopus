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
          const val = parseFloat(e.target.value);
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

export function TerminalSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  if (isLoading) {
    return <TerminalSettingsSkeleton />;
  }

  const hasChanges =
    settings.terminalScrollback !== DEFAULT_SETTINGS.terminalScrollback ||
    settings.terminalFontSize !== DEFAULT_SETTINGS.terminalFontSize ||
    settings.terminalCursorStyle !== DEFAULT_SETTINGS.terminalCursorStyle ||
    settings.terminalCursorBlink !== DEFAULT_SETTINGS.terminalCursorBlink ||
    settings.terminalLineHeight !== DEFAULT_SETTINGS.terminalLineHeight ||
    settings.terminalCursorWidth !== DEFAULT_SETTINGS.terminalCursorWidth ||
    settings.terminalTabStopWidth !== DEFAULT_SETTINGS.terminalTabStopWidth ||
    settings.terminalFontFamily !== DEFAULT_SETTINGS.terminalFontFamily ||
    settings.terminalFontWeight !== DEFAULT_SETTINGS.terminalFontWeight ||
    settings.terminalLetterSpacing !== DEFAULT_SETTINGS.terminalLetterSpacing;

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.terminal.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.terminal.description')}
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
              htmlFor="terminalScrollback"
              title={t('settings.terminal.scrollback.title')}
              description={t('settings.terminal.scrollback.description')}
            />
            <NumberInput
              id="terminalScrollback"
              value={settings.terminalScrollback}
              onChange={(val) => updateSetting('terminalScrollback', val)}
              min={1000}
              max={50000}
              step={1000}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalFontSize"
              title={t('settings.terminal.fontSize.title')}
              description={t('settings.terminal.fontSize.description')}
            />
            <NumberInput
              id="terminalFontSize"
              value={settings.terminalFontSize}
              onChange={(val) => updateSetting('terminalFontSize', val)}
              min={8}
              max={24}
              suffix="px"
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalCursorStyle"
              title={t('settings.terminal.cursorStyle.title')}
              description={t('settings.terminal.cursorStyle.description')}
            />
            <Select
              value={settings.terminalCursorStyle}
              onValueChange={(value) =>
                updateSetting('terminalCursorStyle', value as 'bar' | 'block' | 'underline')
              }
            >
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="bar">
                  {t('settings.terminal.cursorStyle.options.bar')}
                </SelectItem>
                <SelectItem value="block">
                  {t('settings.terminal.cursorStyle.options.block')}
                </SelectItem>
                <SelectItem value="underline">
                  {t('settings.terminal.cursorStyle.options.underline')}
                </SelectItem>
              </SelectContent>
            </Select>
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalCursorBlink"
              title={t('settings.terminal.cursorBlink.title')}
              description={t('settings.terminal.cursorBlink.description')}
            />
            <Switch
              id="terminalCursorBlink"
              checked={settings.terminalCursorBlink}
              onCheckedChange={(checked) => updateSetting('terminalCursorBlink', checked)}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalLineHeight"
              title={t('settings.terminal.lineHeight.title')}
              description={t('settings.terminal.lineHeight.description')}
            />
            <NumberInput
              id="terminalLineHeight"
              value={settings.terminalLineHeight}
              onChange={(val) => updateSetting('terminalLineHeight', val)}
              min={1.0}
              max={2.5}
              step={0.1}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalCursorWidth"
              title={t('settings.terminal.cursorWidth.title')}
              description={t('settings.terminal.cursorWidth.description')}
            />
            <NumberInput
              id="terminalCursorWidth"
              value={settings.terminalCursorWidth}
              onChange={(val) => updateSetting('terminalCursorWidth', val)}
              min={1}
              max={5}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalTabStopWidth"
              title={t('settings.terminal.tabStopWidth.title')}
              description={t('settings.terminal.tabStopWidth.description')}
            />
            <NumberInput
              id="terminalTabStopWidth"
              value={settings.terminalTabStopWidth}
              onChange={(val) => updateSetting('terminalTabStopWidth', val)}
              min={2}
              max={8}
            />
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalFontFamily"
              title={t('settings.terminal.fontFamily.title')}
              description={t('settings.terminal.fontFamily.description')}
            />
            <Select
              value={settings.terminalFontFamily}
              onValueChange={(value) => updateSetting('terminalFontFamily', value)}
            >
              <SelectTrigger className="w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="JetBrains Mono">
                  {t('settings.terminal.fontFamily.options.jetbrainsMono')}
                </SelectItem>
                <SelectItem value="Fira Code">
                  {t('settings.terminal.fontFamily.options.firaCode')}
                </SelectItem>
                <SelectItem value="Cascadia Code">
                  {t('settings.terminal.fontFamily.options.cascadiaCode')}
                </SelectItem>
                <SelectItem value="SF Mono">
                  {t('settings.terminal.fontFamily.options.sfMono')}
                </SelectItem>
                <SelectItem value="Menlo">
                  {t('settings.terminal.fontFamily.options.menlo')}
                </SelectItem>
                <SelectItem value="Monaco">
                  {t('settings.terminal.fontFamily.options.monaco')}
                </SelectItem>
                <SelectItem value="Courier New">
                  {t('settings.terminal.fontFamily.options.courierNew')}
                </SelectItem>
              </SelectContent>
            </Select>
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalFontWeight"
              title={t('settings.terminal.fontWeight.title')}
              description={t('settings.terminal.fontWeight.description')}
            />
            <Select
              value={settings.terminalFontWeight}
              onValueChange={(value) =>
                updateSetting('terminalFontWeight', value as 'normal' | 'bold')
              }
            >
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="normal">
                  {t('settings.terminal.fontWeight.options.normal')}
                </SelectItem>
                <SelectItem value="bold">
                  {t('settings.terminal.fontWeight.options.bold')}
                </SelectItem>
              </SelectContent>
            </Select>
          </SettingRow>

          <SettingRow>
            <SettingLabel
              htmlFor="terminalLetterSpacing"
              title={t('settings.terminal.letterSpacing.title')}
              description={t('settings.terminal.letterSpacing.description')}
            />
            <NumberInput
              id="terminalLetterSpacing"
              value={settings.terminalLetterSpacing}
              onChange={(val) => updateSetting('terminalLetterSpacing', val)}
              min={0}
              max={2}
              step={0.1}
              suffix="px"
            />
          </SettingRow>
        </div>
      </div>
    </div>
  );
}

function TerminalSettingsSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-4 w-72" />
      <Skeleton className="h-48 w-full" />
    </div>
  );
}
