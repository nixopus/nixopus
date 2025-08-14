import React from 'react';
import {
  FormControl,
  FormItem,
  FormLabel,
  FormMessage,
  FormField,
  FormDescription
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';

type FormSelectFieldProps = {
  form: any;
  label: string;
  name: string;
  description?: string;
  placeholder?: string;
  selectOptions?: { label: string; value: string }[];
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
          <Select onValueChange={field.onChange} defaultValue={field.value}>
            <FormControl>
              <SelectTrigger>
                <SelectValue placeholder={placeholder} />
              </SelectTrigger>
            </FormControl>
            <SelectContent>
              {selectOptions?.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}

export default FormSelectField;
