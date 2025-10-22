import React from 'react';
import {
  FormControl,
  FormItem,
  FormLabel,
  FormMessage,
  FormField,
  FormDescription
} from '@/components/ui/form';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';

type FormSelectFieldProps = {
  form: any;
  label: string;
  name: string;
  description?: string;
  placeholder?: string;
  selectOptions?: SelectOption[];
  required?: boolean;
};

function FormSelectField({
  form,
  label,
  name,
  description,
  placeholder,
  selectOptions,
  required = true
}: FormSelectFieldProps) {
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
            <SelectWrapper
              value={field.value}
              onValueChange={field.onChange}
              options={selectOptions || []}
              placeholder={placeholder}
            />
          </FormControl>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}

export default FormSelectField;
