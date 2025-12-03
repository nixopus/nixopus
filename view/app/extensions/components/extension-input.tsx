'use client';

import React from 'react';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { useTranslation } from '@/hooks/use-translation';
import { Extension, ExtensionVariable } from '@/redux/types/extension';
import { useExtensionInput } from '@/app/extensions/hooks/use-extension-input';
import { Info, Sparkles, ChevronDown, Search, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip';

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
  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    searchQuery,
    setSearchQuery,
    showOptional,
    setShowOptional,
    requiredFields,
    optionalFields,
    showSearch,
    hasVariables,
    hasSearchResults
  } = useExtensionInput({
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
      size="xl"
    >
      <div className="space-y-4 py-2 max-h-[65vh] overflow-y-auto px-1">
        {!hasVariables && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Sparkles className="h-12 w-12 text-muted-foreground/40 mb-3" />
            <p className="text-sm text-muted-foreground">{t('extensions.noVariables')}</p>
            <p className="text-xs text-muted-foreground/70 mt-1">
              This extension is ready to run without configuration
            </p>
          </div>
        )}

        {showSearch && (
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              type="text"
              placeholder="Search variables..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 pr-9"
            />
            {searchQuery && (
              <Button
                variant="ghost"
                size="sm"
                className="absolute right-1 top-1/2 -translate-y-1/2 h-7 w-7 p-0"
                onClick={() => setSearchQuery('')}
              >
                <X className="h-3.5 w-3.5" />
              </Button>
            )}
          </div>
        )}

        {requiredFields.length > 0 && (
          <div className="space-y-3">
            <div className="flex items-center gap-2 pb-1.5">
              <div className="h-1.5 w-1.5 rounded-full bg-primary" />
              <h3 className="text-sm font-semibold text-foreground">
                Required ({requiredFields.length})
              </h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
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
          </div>
        )}

        {optionalFields.length > 0 && (
          <Collapsible open={showOptional} onOpenChange={setShowOptional}>
            <CollapsibleTrigger asChild>
              <Button
                variant="ghost"
                className="w-full justify-between p-3 h-auto hover:bg-muted/50"
              >
                <div className="flex items-center gap-2">
                  <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
                  <h3 className="text-sm font-medium text-muted-foreground">
                    Optional Configuration ({optionalFields.length})
                  </h3>
                </div>
                <ChevronDown
                  className={cn(
                    'h-4 w-4 text-muted-foreground transition-transform duration-200',
                    showOptional && 'rotate-180'
                  )}
                />
              </Button>
            </CollapsibleTrigger>
            <CollapsibleContent className="space-y-3 pt-2">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {optionalFields.map((v) => (
                  <FieldItem
                    key={v.id}
                    variable={v}
                    value={values[v.variable_name]}
                    error={errors[v.variable_name]}
                    onChange={handleChange}
                  />
                ))}
              </div>
            </CollapsibleContent>
          </Collapsible>
        )}

        {searchQuery && !hasSearchResults && (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <p className="text-sm text-muted-foreground">No variables found</p>
            <p className="text-xs text-muted-foreground/70 mt-1">
              Try adjusting your search query
            </p>
          </div>
        )}
      </div>
    </DialogWrapper>
  );
}

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

  const labelWithTooltip = (
    <div className="flex items-center gap-1.5">
      <Label
        htmlFor={id}
        className={cn('font-medium text-sm', variable.variable_type === 'boolean' && 'cursor-pointer')}
      >
        {displayName}
      </Label>
      {variable.description && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Info className="h-3.5 w-3.5 text-muted-foreground cursor-help" />
          </TooltipTrigger>
          <TooltipContent side="right" className="max-w-xs">
            {variable.description}
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );

  const renderInput = () => {
    if (variable.variable_type === 'boolean') {
      return (
        <div className="flex items-start gap-3">
          <Checkbox
            id={id}
            checked={Boolean(value)}
            onCheckedChange={(v) => onChange(variable.variable_name, Boolean(v))}
            className="mt-0.5"
          />
          <div className="flex-1">{labelWithTooltip}</div>
        </div>
      );
    }

    if (variable.variable_type === 'array') {
      const textValue = Array.isArray(value)
        ? (value as unknown[]).map((v) => String(v)).join('\n')
        : String(value ?? '');
      return (
        <>
          {labelWithTooltip}
          <textarea
            id={id}
            className="min-h-[100px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-0 disabled:cursor-not-allowed disabled:opacity-50 transition-shadow"
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
            placeholder="Enter each item on a new line..."
          />
          <span className="text-xs text-muted-foreground/70">One item per line</span>
        </>
      );
    }

    return (
      <>
        {labelWithTooltip}
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
            'focus-visible:ring-primary',
            variable.variable_type === 'integer' &&
              '[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none'
          )}
        />
      </>
    );
  };

  return (
    <div
      className={cn(
        'relative rounded-lg p-3 transition-all duration-200',
        error
          ? 'bg-destructive/5 border border-destructive'
          : 'bg-muted/30 hover:bg-muted/50',
        variable.variable_type === 'array' && 'md:col-span-2',
        variable.variable_type !== 'boolean' && 'grid gap-2'
      )}
    >
      {renderInput()}
      {error && (
        <div className="flex items-center gap-1.5 mt-2 text-xs text-destructive font-medium">
          <Info className="h-3.5 w-3.5" />
          {error}
        </div>
      )}
    </div>
  );
}
