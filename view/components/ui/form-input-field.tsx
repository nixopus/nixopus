import React from 'react';
import {
  FormControl,
  FormDescription,
  FormItem,
  FormLabel,
  FormMessage,
  FormField
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';

interface FormInputFieldProps {
  form: any;
  label: string;
  name: string;
  description: string;
  placeholder?: string;
  required?: boolean;
}

function FormInputField({
  form,
  label,
  name,
  description,
  placeholder,
  required = true
}: FormInputFieldProps) {
  return (
    <div>
      <FormField
        control={form.control}
        name={name}
        render={({ field }) => (
          <FormItem>
            <div className="flex gap-2">
              {label && <FormLabel>{label}</FormLabel>}{' '}
              {required && <span className="text-destructive">*</span>}
            </div>
            <FormControl>
              <Input placeholder={placeholder} {...field} />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
}

export default FormInputField;
