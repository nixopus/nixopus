'use client';

import React from 'react';
import { BadgeGroup, BadgeGroupItem } from '@nixopus/ui';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Skeleton } from '@nixopus/ui';
import { Alert, AlertDescription } from '@nixopus/ui';
import { AlertCircle } from 'lucide-react';
import { Button } from '@nixopus/ui';
import { CardWrapper } from '@nixopus/ui';
import { ExternalLink, ChevronDown } from 'lucide-react';
import { CardDescription } from '@nixopus/ui';
import {
  CategoryBadgesProps,
  ExtensionsGridProps,
  ExtensionCardProps,
  ExtensionInputProps
} from '@/packages/types/extension';
import { Sparkles } from 'lucide-react';
import { Input } from '@nixopus/ui';
import { Textarea } from '@nixopus/ui';
import { Checkbox } from '@nixopus/ui';
import { Label } from '@nixopus/ui';
import { ExtensionVariable } from '@/redux/types/extension';
import { DialogWrapper } from '@nixopus/ui';
import { ServerSelector } from './server-selector';

export default function CategoryBadges({
  categories,
  selected = null,
  onChange,
  className,
  showAll = true
}: CategoryBadgesProps) {
  const handleSelect = (value: string | null) => {
    onChange?.(value);
  };

  return (
    <BadgeGroup className={cn('w-full', className)} selectable>
      {showAll && (
        <BadgeGroupItem
          selected={selected === null}
          clickable
          onClick={() => handleSelect(null)}
          className="px-3 py-1"
        >
          All
        </BadgeGroupItem>
      )}
      {categories.map((cat) => (
        <BadgeGroupItem
          key={cat}
          selected={selected === cat}
          clickable
          onClick={() => handleSelect(cat === selected ? null : cat)}
          className="px-3 py-1"
        >
          {cat}
        </BadgeGroupItem>
      ))}
    </BadgeGroup>
  );
}

export function ExtensionGrid({
  extensions = [],
  isLoading = false,
  error,
  onInstall,
  onViewDetails,
  expanded,
  setExpanded
}: ExtensionsGridProps) {
  const { t } = useTranslation();

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (isLoading) {
    return <ExtensionGridSkeleton />;
  }

  if (extensions.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="mx-auto max-w-md">
          <div className="text-6xl mb-4">🔍</div>
          <h3 className="text-lg font-semibold mb-2">{t('extensions.noExtensions')}</h3>
          <p className="text-muted-foreground">{t('extensions.noExtensionsHint')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {extensions.map((extension) => (
        <ExtensionCard
          key={extension.id}
          extension={extension}
          onInstall={onInstall}
          onViewDetails={onViewDetails}
          expanded={expanded}
          setExpanded={setExpanded}
          t={t}
        />
      ))}
    </div>
  );
}

export function ExtensionGridSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <div
          key={i}
          className="group h-full transition-all duration-200 bg-card border-border p-6 rounded-xl"
        >
          <div className="space-y-4">
            <div className="flex items-start gap-4">
              <Skeleton className="h-12 w-12 rounded-full flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <Skeleton className="h-6 w-48 mb-2" />
                <Skeleton className="h-4 w-32" />
              </div>
            </div>
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
            <div className="flex gap-2 pt-6">
              <Skeleton className="h-10 flex-1" />
              <Skeleton className="h-10 w-10" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

const MAX_DESCRIPTION_CHARS = 90;

export function ExtensionCard({
  extension,
  onInstall,
  onViewDetails,
  expanded,
  setExpanded,
  t
}: ExtensionCardProps) {
  const customHeader = (
    <div className="flex items-center gap-4 w-full">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted shrink-0 overflow-hidden">
        {extension.icon?.startsWith('http') ? (
          <img src={extension.icon} alt={extension.name} className="size-8 object-contain" />
        ) : (
          <span className="text-lg font-bold text-muted-foreground">{extension.icon}</span>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <h3 className="text-lg font-bold mb-1">{extension.name}</h3>
      </div>
    </div>
  );

  return (
    <CardWrapper
      className="h-full transition-shadow hover:shadow-lg"
      header={customHeader}
      contentClassName="space-y-4"
    >
      <CardDescription>
        {expanded || extension.description.length <= MAX_DESCRIPTION_CHARS
          ? extension.description
          : `${extension.description.slice(0, MAX_DESCRIPTION_CHARS)}…`}
        {extension.description.length > MAX_DESCRIPTION_CHARS && (
          <Button
            variant="link"
            className="ml-2 text-primary hover:underline"
            size="sm"
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? t('common.readLess') || 'Read less' : t('common.readMore') || 'Read more'}{' '}
            <ChevronDown className="size-4" />
          </Button>
        )}
      </CardDescription>
      <div className="flex gap-2">
        <Button onClick={() => onInstall?.(extension)} className="min-w-[100px]">
          {t('extensions.install')}
        </Button>
        <Button
          variant="ghost"
          onClick={() => onViewDetails?.(extension)}
          className="min-w-[100px]"
        >
          {t('extensions.viewDetails')}
          <ExternalLink className="ml-2 h-4 w-4" />
        </Button>
      </div>
    </CardWrapper>
  );
}

export function ExtensionInput({
  open,
  onOpenChange,
  extension,
  onSubmit,
  t,
  actions,
  isOnlyProxyDomain,
  noFieldsToShow,
  values,
  errors,
  handleChange,
  handleSubmit,
  requiredFields,
  serverSelector
}: ExtensionInputProps) {
  if (noFieldsToShow) {
    return (
      <DialogWrapper
        open={open}
        onOpenChange={onOpenChange}
        title={
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-primary" />
            <span>{extension?.name || t('extensions.run')}</span>
          </div>
        }
        description={extension?.description}
        actions={actions}
        size="md"
      >
        <div className="py-2 space-y-4">
          <p className="text-sm text-muted-foreground">No additional fields required.</p>
          {serverSelector?.isMultiServer && (
            <ServerSelector
              servers={serverSelector.servers}
              selectedServerIds={serverSelector.selectedServerIds}
              onToggleServer={serverSelector.toggleServer}
              routingStrategy={serverSelector.routingStrategy}
              onRoutingStrategyChange={serverSelector.setRoutingStrategy}
              primaryServerId={serverSelector.primaryServerId}
              onPrimaryServerIdChange={serverSelector.setPrimaryServerId}
            />
          )}
        </div>
      </DialogWrapper>
    );
  }

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title={
        <div className="flex items-center gap-2">
          <span>{extension?.name || t('extensions.run')}</span>
        </div>
      }
      description={extension?.description}
      actions={actions}
      size={isOnlyProxyDomain ? 'md' : 'lg'}
    >
      <div className="py-2 space-y-4">
        {requiredFields.map((variable) => {
          const isProxyDomain =
            variable.variable_name.toLowerCase() === 'proxy_domain' ||
            variable.variable_name.toLowerCase() === 'domain';

          if (isProxyDomain) {
            return (
              <ProxyDomainInput
                key={variable.id}
                variable={variable}
                value={values[variable.variable_name]}
                error={errors[variable.variable_name]}
                onChange={handleChange}
              />
            );
          }

          return (
            <VariableInput
              key={variable.id}
              variable={variable}
              value={values[variable.variable_name]}
              error={errors[variable.variable_name]}
              onChange={handleChange}
            />
          );
        })}
        {serverSelector?.isMultiServer && (
          <ServerSelector
            servers={serverSelector.servers}
            selectedServerIds={serverSelector.selectedServerIds}
            onToggleServer={serverSelector.toggleServer}
            routingStrategy={serverSelector.routingStrategy}
            onRoutingStrategyChange={serverSelector.setRoutingStrategy}
            primaryServerId={serverSelector.primaryServerId}
            onPrimaryServerIdChange={serverSelector.setPrimaryServerId}
          />
        )}
      </div>
    </DialogWrapper>
  );
}

function ProxyDomainInput({
  variable,
  value,
  error,
  onChange
}: {
  variable: ExtensionVariable;
  value: unknown;
  error?: string;
  onChange: (name: string, value: unknown) => void;
}) {
  const id = `var-${variable.variable_name}`;
  return (
    <div className="space-y-2">
      <Label htmlFor={id} className="text-sm font-medium flex items-center gap-2">
        Custom Domain
        {variable.is_required && <span className="text-destructive">*</span>}
      </Label>
      <Input
        id={id}
        type="text"
        value={(value as string) ?? ''}
        onChange={(e) => onChange(variable.variable_name, e.target.value)}
        placeholder={variable.description || 'app.example.com'}
        className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
        autoFocus
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
      {variable.description && (
        <p className="text-xs text-muted-foreground">{variable.description}</p>
      )}
    </div>
  );
}

function VariableInput({
  variable,
  value,
  error,
  onChange
}: {
  variable: ExtensionVariable;
  value: unknown;
  error?: string;
  onChange: (name: string, value: unknown) => void;
}) {
  const id = `var-${variable.variable_name}`;

  const renderInput = () => {
    switch (variable.variable_type) {
      case 'integer':
        return (
          <Input
            id={id}
            type="number"
            value={(value as number) ?? ''}
            onChange={(e) => {
              const num = e.target.value === '' ? '' : Number(e.target.value);
              onChange(variable.variable_name, num);
            }}
            placeholder={variable.description || String(variable.default_value ?? '')}
            className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
          />
        );
      case 'array':
        const arrayValue = Array.isArray(value) ? value.join(', ') : ((value as string) ?? '');
        return (
          <Input
            id={id}
            type="text"
            value={arrayValue}
            onChange={(e) => {
              const arr = e.target.value
                .split(',')
                .map((x) => x.trim())
                .filter((x) => x.length > 0);
              onChange(variable.variable_name, arr);
            }}
            placeholder={variable.description || 'comma-separated values'}
            className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
          />
        );
      case 'boolean':
        const checked = value === true || value === 'true' || value === 1;
        return (
          <div className="flex items-center gap-2">
            <Checkbox
              id={id}
              checked={checked}
              onCheckedChange={(checked) => onChange(variable.variable_name, checked)}
            />
            <Label htmlFor={id} className="text-sm font-normal cursor-pointer">
              {variable.description || 'Enable'}
            </Label>
          </div>
        );
      default:
        const isLongText = variable.description && variable.description.length > 100;
        if (isLongText) {
          return (
            <Textarea
              id={id}
              value={(value as string) ?? ''}
              onChange={(e) => onChange(variable.variable_name, e.target.value)}
              placeholder={variable.description || ''}
              className={cn(
                'min-h-[100px]',
                error && 'border-destructive focus-visible:ring-destructive'
              )}
            />
          );
        }
        return (
          <Input
            id={id}
            type="text"
            value={(value as string) ?? ''}
            onChange={(e) => onChange(variable.variable_name, e.target.value)}
            placeholder={variable.description || String(variable.default_value ?? '')}
            className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
          />
        );
    }
  };

  return (
    <div className="space-y-2">
      <Label htmlFor={id} className="text-sm font-medium">
        {variable.variable_name}
        {variable.is_required && <span className="text-destructive ml-1">*</span>}
        {variable.variable_type && (
          <span className="text-xs text-muted-foreground ml-2 font-normal">
            ({variable.variable_type})
          </span>
        )}
      </Label>
      {renderInput()}
      {error && <p className="text-xs text-destructive">{error}</p>}
      {variable.description && variable.variable_type !== 'boolean' && (
        <p className="text-xs text-muted-foreground">{variable.description}</p>
      )}
    </div>
  );
}
