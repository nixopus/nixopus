'use client';

import React from 'react';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { useTranslation } from '@/hooks/use-translation';
import { Extension, ExtensionVariable } from '@/redux/types/extension';
import { useExtensionInput } from '@/app/extensions/hooks/use-extension-input';

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

  const submit = () => {
    onSubmit?.(values);
    onOpenChange(false);
  };

  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: () => onOpenChange(false),
      variant: 'ghost'
    },
    {
      label: t('extensions.run'),
      onClick: submit,
      variant: 'default'
    }
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title={extension?.name || t('extensions.run')}
      description={extension?.description}
      actions={actions}
      size="lg"
    >
      <div className="space-y-4 py-2">
        {variables.length === 0 && (
          <div className="text-sm text-muted-foreground">{t('extensions.noVariables')}</div>
        )}
        {variables.map((v) => (
          <div key={v.id} className="space-y-1">
            <Field variable={v} value={values[v.variable_name]} onChange={handleChange} />
            {errors[v.variable_name] && (
              <div className="text-xs text-destructive">{errors[v.variable_name]}</div>
            )}
          </div>
        ))}
      </div>
    </DialogWrapper>
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
      <div className="flex items-center gap-3">
        <Checkbox
          id={id}
          checked={Boolean(value)}
          onCheckedChange={(v) => onChange(variable.variable_name, Boolean(v))}
        />
        <div className="grid gap-1">
          <Label htmlFor={id}>{variable.variable_name}</Label>
          {variable.description && (
            <span className="text-xs text-muted-foreground">{variable.description}</span>
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
      <div className="grid gap-1">
        <Label htmlFor={id}>{variable.variable_name}</Label>
        <textarea
          id={id}
          className="min-h-[100px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
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
          placeholder={variable.description || variable.variable_name}
        />
        {variable.description && (
          <span className="text-xs text-muted-foreground">{variable.description}</span>
        )}
      </div>
    );
  }
  return (
    <div className="grid gap-1">
      <Label htmlFor={id}>{variable.variable_name}</Label>
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
        placeholder={variable.description || variable.variable_name}
      />
    </div>
  );
}
