'use client';

import React from 'react';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { useTranslation } from '@/hooks/use-translation';
import { Extension, ExtensionVariable } from '@/redux/types/extension';
import { useExtensionInput } from '@/app/extensions/hooks/use-extension-input';
import { Sparkles, Globe } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ExtensionInputProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension?: Extension | null;
  onSubmit?: (values: Record<string, unknown>) => void;
}

export default function ExtensionInput({
  open,
  onOpenChange,
  extension,
  onSubmit
}: ExtensionInputProps) {
  const { t } = useTranslation();
  const { values, errors, handleChange, handleSubmit, requiredFields } = useExtensionInput({
    extension,
    open,
    onSubmit,
    onClose: () => onOpenChange(false)
  });

  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: () => onOpenChange(false),
      variant: 'ghost'
    },
    {
      label: t('extensions.run'),
      onClick: handleSubmit,
      variant: 'default'
    }
  ];

  // Check if it's just proxy_domain - show simplified view
  const isOnlyProxyDomain =
    requiredFields.length === 1 &&
    (requiredFields[0].variable_name.toLowerCase() === 'proxy_domain' ||
      requiredFields[0].variable_name.toLowerCase() === 'domain');

  const noFieldsToShow = requiredFields.length === 0;

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
      size={noFieldsToShow || isOnlyProxyDomain ? 'md' : 'lg'}
    >
      <div className="py-2">
        {noFieldsToShow && (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <Sparkles className="h-10 w-10 text-muted-foreground/40 mb-2" />
            <p className="text-sm text-muted-foreground">{t('extensions.noVariables')}</p>
            <p className="text-xs text-muted-foreground/60 mt-1">
              Ready to run without configuration
            </p>
          </div>
        )}

        {isOnlyProxyDomain && (
          <ProxyDomainInput
            variable={requiredFields[0]}
            value={values[requiredFields[0].variable_name]}
            error={errors[requiredFields[0].variable_name]}
            onChange={handleChange}
          />
        )}

        {!noFieldsToShow && !isOnlyProxyDomain && (
          <div className="space-y-3">
            {requiredFields.map((v) => (
              <FieldItem
                key={v.id}
                variable={v}
                value={values[v.variable_name]}
                error={errors[v.variable_name]}
                onChange={handleChange}
              />
            ))}
          </div>
        )}
      </div>
    </DialogWrapper>
  );
}

// Simplified input for the common proxy_domain case
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
        <Globe className="h-4 w-4 text-muted-foreground" />
        Domain
      </Label>
      <Input
        id={id}
        type="text"
        value={(value as string) ?? ''}
        onChange={(e) => onChange(variable.variable_name, e.target.value)}
        placeholder="app.example.com"
        className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
        autoFocus
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
      <p className="text-xs text-muted-foreground">
        Enter the domain where this extension will be accessible
      </p>
    </div>
  );
}

// Generic field input for other cases
function FieldItem({
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
  const displayName = variable.variable_name
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');

  if (variable.variable_type === 'boolean') {
    return (
      <div className="flex items-center gap-3 py-1">
        <Checkbox
          id={id}
          checked={Boolean(value)}
          onCheckedChange={(v) => onChange(variable.variable_name, Boolean(v))}
        />
        <Label htmlFor={id} className="text-sm cursor-pointer">
          {displayName}
        </Label>
      </div>
    );
  }

  if (variable.variable_type === 'array') {
    const textValue = Array.isArray(value)
      ? (value as unknown[]).map((v) => String(v)).join('\n')
      : String(value ?? '');
    return (
      <div className="space-y-1.5">
        <Label htmlFor={id} className="text-sm font-medium">
          {displayName}
        </Label>
        <textarea
          id={id}
          className={cn(
            'w-full min-h-[80px] rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary',
            error && 'border-destructive focus-visible:ring-destructive'
          )}
          value={textValue}
          onChange={(e) =>
            onChange(
              variable.variable_name,
              e.target.value
                .split('\n')
                .map((s) => s.trim())
                .filter((s) => s.length > 0)
            )
          }
          placeholder="One item per line..."
        />
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
    );
  }

  return (
    <div className="space-y-1.5">
      <Label htmlFor={id} className="text-sm font-medium">
        {displayName}
      </Label>
      <Input
        id={id}
        type={variable.variable_type === 'integer' ? 'number' : 'text'}
        value={(value as string) ?? ''}
        onChange={(e) =>
          onChange(
            variable.variable_name,
            variable.variable_type === 'integer' ? Number(e.target.value) : e.target.value
          )
        }
        placeholder={`Enter ${displayName.toLowerCase()}...`}
        className={cn(
          error && 'border-destructive focus-visible:ring-destructive',
          variable.variable_type === 'integer' &&
            '[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none'
        )}
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}
