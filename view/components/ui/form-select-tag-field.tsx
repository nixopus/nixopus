'use client';
import React, { useEffect, useState, KeyboardEvent } from 'react';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { X } from 'lucide-react';

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

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();

      if (!inputValue.trim()) return;

      const validation = validator(inputValue);

      if (validation.isValid && 'key' in validation && 'value' in validation) {
        setSelectedValue((prev) => ({
          ...prev,
          [validation.key as string]: validation.value as string
        }));
        setInputValue('');
        setError(null);
      } else {
        setError(validation.error ?? null);
      }
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
              <Input
                placeholder={placeholder}
                value={inputValue}
                onChange={(e) => {
                  setInputValue(e.target.value);
                  if (error) setError(null);
                }}
                onKeyDown={handleKeyDown}
                className={error ? 'border-red-500' : ''}
              />

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
