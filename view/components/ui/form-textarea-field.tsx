import React from 'react';
import {
  FormControl,
  FormDescription,
  FormItem,
  FormLabel,
  FormMessage,
  FormField
} from '@/components/ui/form';
import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

interface FormTextareaFieldProps {
  form: any;
  label: string;
  name: string;
  description?: string;
  placeholder?: string;
  required?: boolean;
  validator?: (value: string) => boolean;
  rows?: number;
  className?: string;
}

function FormTextareaField({
  form,
  label,
  name,
  description,
  placeholder,
  required = true,
  validator,
  rows = 4,
  className
}: FormTextareaFieldProps) {
  return (
    <div>
      <FormField
        control={form.control}
        name={name}
        rules={{
          validate: validator
            ? {
                custom: (value: string) => {
                  if (!value) return true;
                  return validator(value) || `Invalid ${name} format`;
                }
              }
            : undefined
        }}
        render={({ field }) => (
          <FormItem>
            <div className="flex gap-2">
              {label && <FormLabel>{label}</FormLabel>}
              <span className="text-destructive w-3 flex-shrink-0 text-right">
                {required ? '*' : ''}
              </span>
            </div>
            <FormControl>
              <Textarea
                placeholder={placeholder}
                rows={rows}
                className={cn(className)}
                {...field}
              />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
}

export default FormTextareaField;
