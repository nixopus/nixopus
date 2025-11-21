'use client';

import React from 'react';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { useTranslation } from '@/hooks/use-translation';
import { Extension, ExtensionVariable } from '@/redux/types/extension';
import { useExtensionInput } from '@/app/extensions/hooks/use-extension-input';
import { Info, Sparkles } from 'lucide-react';
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
  const { variables, values, errors, handleChange, handleSubmit } = useExtensionInput({
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

  const requiredFields = variables.filter((v) => v.is_required);
  const optionalFields = variables.filter((v) => !v.is_required);

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
      <div className="space-y-6 py-2 max-h-[65vh] overflow-y-auto px-1">
        {variables.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Sparkles className="h-12 w-12 text-muted-foreground/40 mb-3" />
            <p className="text-sm text-muted-foreground">{t('extensions.noVariables')}</p>
            <p className="text-xs text-muted-foreground/70 mt-1">
              This extension is ready to run without configuration
            </p>
          </div>
        )}

        {requiredFields.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b">
              <div className="h-1.5 w-1.5 rounded-full bg-primary" />
              <h3 className="text-sm font-semibold text-foreground">Required Configuration</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {requiredFields.map((v) => (
                <FieldWrapper
                  key={v.id}
                  error={errors[v.variable_name]}
                  fullWidth={v.variable_type === 'array'}
                >
                  <Field variable={v} value={values[v.variable_name]} onChange={handleChange} />
                </FieldWrapper>
              ))}
            </div>
          </div>
        )}

        {optionalFields.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b">
              <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
              <h3 className="text-sm font-medium text-muted-foreground">
                Optional Configuration
              </h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {optionalFields.map((v) => (
                <FieldWrapper
                  key={v.id}
                  error={errors[v.variable_name]}
                  fullWidth={v.variable_type === 'array'}
                >
                  <Field variable={v} value={values[v.variable_name]} onChange={handleChange} />
                </FieldWrapper>
              ))}
            </div>
          </div>
        )}
      </div>
    </DialogWrapper>
  );
}

function FieldWrapper({
  error,
  children,
  fullWidth = false
}: {
  error?: string;
  children: React.ReactNode;
  fullWidth?: boolean;
}) {
  return (
    <div
      className={cn(
        'relative rounded-lg border p-4 transition-all duration-200',
        error
          ? 'border-destructive bg-destructive/5'
          : 'border-border bg-card hover:border-primary/50 hover:shadow-sm',
        fullWidth && 'md:col-span-2'
      )}
    >
      {children}
      {error && (
        <div className="flex items-center gap-1.5 mt-2 text-xs text-destructive font-medium">
          <Info className="h-3.5 w-3.5" />
          {error}
        </div>
      )}
    </div>
  );
}

function Field({
  variable,
  value,
  onChange
}: {
  variable: ExtensionVariable;
  value: unknown;
  onChange: (name: string, value: unknown) => void;
}) {
  const id = `var-${variable.variable_name}`;

  if (variable.variable_type === 'boolean') {
    return (
      <div className="flex items-start gap-3">
        <Checkbox
          id={id}
          checked={Boolean(value)}
          onCheckedChange={(v) => onChange(variable.variable_name, Boolean(v))}
          className="mt-0.5"
        />
        <div className="flex-1 grid gap-1.5">
          <Label htmlFor={id} className="font-medium text-sm cursor-pointer">
            {formatVariableName(variable.variable_name)}
            {variable.is_required && <span className="text-destructive ml-1">*</span>}
          </Label>
          {variable.description && (
            <span className="text-xs text-muted-foreground leading-relaxed">
              {variable.description}
            </span>
          )}
        </div>
      </div>
    );
  }

  if (variable.variable_type === 'array') {
    const textValue = Array.isArray(value)
      ? (value as unknown[]).map((v) => String(v)).join('\n')
      : String(value ?? '');
    return (
      <div className="grid gap-2">
        <Label htmlFor={id} className="font-medium text-sm">
          {formatVariableName(variable.variable_name)}
          {variable.is_required && <span className="text-destructive ml-1">*</span>}
        </Label>
        {variable.description && (
          <span className="text-xs text-muted-foreground -mt-1">{variable.description}</span>
        )}
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
          placeholder={`Enter each item on a new line...`}
        />
        <span className="text-xs text-muted-foreground/70">One item per line</span>
      </div>
    );
  }

  return (
    <div className="grid gap-2">
      <Label htmlFor={id} className="font-medium text-sm">
        {formatVariableName(variable.variable_name)}
        {variable.is_required && <span className="text-destructive ml-1">*</span>}
      </Label>
      {variable.description && (
        <span className="text-xs text-muted-foreground -mt-1">{variable.description}</span>
      )}
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
        placeholder={`Enter ${formatVariableName(variable.variable_name).toLowerCase()}...`}
        className="focus-visible:ring-primary"
      />
    </div>
  );
}

function formatVariableName(name: string): string {
  return name
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}
