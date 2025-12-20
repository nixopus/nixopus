'use client';
import React, { useEffect, useState, KeyboardEvent, ClipboardEvent } from 'react';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { Textarea } from '@/components/ui/textarea';
import { X, Info } from 'lucide-react';
import { parseEnvText, isMultiLineEnvPaste } from '@/lib/parse-env';

interface FormSelectTagField {
  label: string;
  name: string;
  description?: string;
  placeholder: string;
  form: any;
  required?: boolean;
  validator: (value: string) => ValidationType;
  defaultValues?: Record<string, string>;
}

interface ValidationType {
  isValid: boolean;
  error?: string;
  key?: string;
  value?: string;
}

export const FormSelectTagInputField = ({
  label,
  name,
  description,
  placeholder,
  form,
  validator,
  required = false,
  defaultValues = {}
}: FormSelectTagField) => {
  const [inputValue, setInputValue] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [value, setSelectedValue] = useState<Record<string, string>>(defaultValues);

  useEffect(() => {
    form.setValue(name, value);
  }, [value, form, name]);

  useEffect(() => {
    if (defaultValues && Object.keys(defaultValues).length > 0) {
      setSelectedValue(defaultValues);
    }
  }, [defaultValues]);

  const processInput = (input: string) => {
    const validation = validator(input);
    if (validation.isValid && 'key' in validation && 'value' in validation) {
      setSelectedValue((prev) => ({
        ...prev,
        [validation.key as string]: validation.value as string
      }));
      return true;
    }
    setError(validation.error ?? null);
    return false;
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (!inputValue.trim()) return;

      if (processInput(inputValue)) {
        setInputValue('');
        setError(null);
      }
    }
  };

  const handlePaste = (e: ClipboardEvent<HTMLTextAreaElement>) => {
    const pastedText = e.clipboardData.getData('text');
    if (isMultiLineEnvPaste(pastedText)) {
      e.preventDefault();
      const parsed = parseEnvText(pastedText);
      setSelectedValue((prev) => ({ ...prev, ...parsed }));
      setInputValue('');
      setError(null);
    }
  };

  const removeSelected = (key: string) => {
    setSelectedValue((prev) => {
      const newVars = { ...prev };
      delete newVars[key];
      return newVars;
    });
  };

  return (
    <FormField
      control={form.control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <div className="flex gap-2">
            {label && <FormLabel>{label}</FormLabel>}
            <span className="text-destructive w-3 flex-shrink-0 text-right">
              {required ? '*' : ''}
            </span>
          </div>
          <FormControl>
            <div className="space-y-2">
              <Textarea
                placeholder={placeholder}
                value={inputValue}
                onChange={(e) => {
                  setInputValue(e.target.value);
                  if (error) setError(null);
                }}
                onKeyDown={handleKeyDown}
                onPaste={handlePaste}
                className={error ? 'border-red-500' : ''}
                rows={3}
              />

              <div className="flex items-start gap-1.5 text-xs text-muted-foreground">
                <Info size={14} className="mt-0.5 flex-shrink-0" />
                <span>
                  Press <kbd className="px-1 py-0.5 bg-muted rounded text-[11px]">Enter</kbd> to add
                  one variable, or paste multiple lines (e.g., from .env file) to add all at once
                </span>
              </div>

              {error && <p className="text-sm font-medium text-red-500">{error}</p>}

              <div className="flex flex-wrap gap-2 mt-2">
                {Object.entries(value).map(([key, value]) => (
                  <div
                    key={key}
                    className="flex items-center gap-1 px-2 py-1 text-sm rounded-md bg-secondary text-secondary-foreground max-w-[200px]"
                  >
                    <span className="font-medium truncate">{key}</span>
                    <span>=</span>
                    <span className="truncate">{value}</span>
                    <button
                      type="button"
                      onClick={() => removeSelected(key)}
                      className="ml-1 text-muted-foreground shrink-0"
                    >
                      <X size={14} />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </FormControl>
          <FormDescription>{description}</FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  );
};
