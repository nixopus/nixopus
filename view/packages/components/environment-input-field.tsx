'use client';

import React from 'react';
import {
  FormControl,
  FormDescription,
  FormItem,
  FormLabel,
  FormMessage,
  FormField,
  Input
} from '@nixopus/ui';
import { formatEnvironmentName } from '@/packages/utils/environment';
import { SUGGESTED_ENVIRONMENTS } from '@/redux/types/deploy-form';

interface EnvironmentInputFieldProps {
  form: any;
  name: string;
  label: string;
  required?: boolean;
}

export function EnvironmentInputField({
  form,
  name,
  label,
  required = false
}: EnvironmentInputFieldProps) {
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
            <Input
              placeholder="e.g. production, staging, qa"
              {...field}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                field.onChange(formatEnvironmentName(e.target.value));
              }}
              onBlur={(e: React.FocusEvent<HTMLInputElement>) => {
                field.onBlur();
                const formatted = formatEnvironmentName(e.target.value);
                if (formatted !== e.target.value) {
                  field.onChange(formatted);
                }
              }}
            />
          </FormControl>
          <FormDescription>
            <span className="text-xs text-muted-foreground">
              Lowercase letters, numbers, and hyphens.{' '}
              {SUGGESTED_ENVIRONMENTS.map((env, i) => (
                <React.Fragment key={env}>
                  <button
                    type="button"
                    className="text-primary hover:underline cursor-pointer"
                    onClick={() => field.onChange(env)}
                  >
                    {env}
                  </button>
                  {i < SUGGESTED_ENVIRONMENTS.length - 1 ? ', ' : ''}
                </React.Fragment>
              ))}
            </span>
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
