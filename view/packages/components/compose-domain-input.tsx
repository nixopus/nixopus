'use client';

import React, { useCallback } from 'react';
import { Controller, UseFormReturn } from 'react-hook-form';
import { FormItem, FormLabel, FormControl, FormDescription, FormMessage } from '@nixopus/ui';
import { Input } from '@nixopus/ui';
import { Button } from '@nixopus/ui';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@nixopus/ui';
import { X, Plus } from 'lucide-react';
import { cn } from '@/lib/utils';
import { defaultValidator } from '@/packages/hooks/applications/use_multiple_domains';

interface ComposeDomainEntry {
  domain: string;
  service_name: string;
  port: number;
}

interface ServiceInfo {
  service_name: string;
  port: number;
}

interface ComposeDomainInputProps {
  form: UseFormReturn<any>;
  label: string;
  name: string;
  composeServices: ServiceInfo[];
  description?: string | React.ReactNode;
  placeholder?: string;
  required?: boolean;
  maxDomains?: number;
}

export function ComposeDomainInput({
  form,
  label,
  name,
  composeServices,
  description,
  placeholder = 'example.com',
  required = false,
  maxDomains = 5
}: ComposeDomainInputProps) {
  const entries: ComposeDomainEntry[] = form.watch(name) || [
    { domain: '', service_name: '', port: 0 }
  ];

  const setEntries = useCallback(
    (newEntries: ComposeDomainEntry[]) => {
      form.setValue(name, newEntries, { shouldValidate: true, shouldDirty: true });
    },
    [form, name]
  );

  const addEntry = useCallback(() => {
    if (entries.length < maxDomains) {
      setEntries([...entries, { domain: '', service_name: '', port: 0 }]);
    }
  }, [entries, maxDomains, setEntries]);

  const removeEntry = useCallback(
    (index: number) => {
      if (entries.length > 1) {
        setEntries(entries.filter((_, i) => i !== index));
      }
    },
    [entries, setEntries]
  );

  const updateDomain = useCallback(
    (index: number, value: string) => {
      const updated = [...entries];
      updated[index] = { ...updated[index], domain: value };
      setEntries(updated);
    },
    [entries, setEntries]
  );

  const updateService = useCallback(
    (index: number, serviceName: string) => {
      const updated = [...entries];
      const service = composeServices.find((s) => s.service_name === serviceName);
      updated[index] = {
        ...updated[index],
        service_name: serviceName,
        port: service?.port ?? 0
      };
      setEntries(updated);
    },
    [entries, composeServices, setEntries]
  );

  const updatePort = useCallback(
    (index: number, port: string) => {
      const updated = [...entries];
      updated[index] = { ...updated[index], port: parseInt(port) || 0 };
      setEntries(updated);
    },
    [entries, setEntries]
  );

  return (
    <Controller
      control={form.control}
      name={name as any}
      render={({ fieldState }: { field: any; fieldState: any }) => (
        <FormItem>
          <div className="flex gap-2">
            {label && <FormLabel>{label}</FormLabel>}
            <span className="text-destructive w-3 flex-shrink-0 text-right">
              {required ? '*' : ''}
            </span>
          </div>
          <div className="space-y-3">
            {entries.map((entry, index) => (
              <div key={index} className="flex gap-2 items-start">
                <FormControl>
                  <Input
                    placeholder={placeholder}
                    value={entry.domain}
                    onChange={(e) => updateDomain(index, e.target.value)}
                    className={cn(
                      'flex-1 min-w-0',
                      fieldState.error && !defaultValidator(entry.domain) && 'border-destructive'
                    )}
                  />
                </FormControl>

                {composeServices.length > 0 ? (
                  <Select
                    value={entry.service_name || undefined}
                    onValueChange={(val) => updateService(index, val)}
                  >
                    <SelectTrigger className="w-[180px] flex-shrink-0">
                      <SelectValue placeholder="Select service" />
                    </SelectTrigger>
                    <SelectContent>
                      {composeServices.map((svc) => (
                        <SelectItem key={svc.service_name} value={svc.service_name}>
                          {svc.service_name}
                          {svc.port > 0 ? ` :${svc.port}` : ''}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : (
                  <FormControl>
                    <Input
                      placeholder="Service name"
                      value={entry.service_name}
                      onChange={(e) => updateService(index, e.target.value)}
                      className="w-[180px] flex-shrink-0"
                    />
                  </FormControl>
                )}

                <FormControl>
                  <Input
                    type="number"
                    placeholder="Port"
                    value={entry.port || ''}
                    onChange={(e) => updatePort(index, e.target.value)}
                    className="w-[90px] flex-shrink-0"
                  />
                </FormControl>

                {entries.length > 1 && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeEntry(index)}
                    className="flex-shrink-0"
                    aria-label="Remove domain"
                  >
                    <X className="h-4 w-4" />
                  </Button>
                )}

                {index === entries.length - 1 && entries.length < maxDomains && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={addEntry}
                    className="flex-shrink-0"
                    aria-label="Add domain"
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                )}
              </div>
            ))}
          </div>
          {description && <FormDescription>{description}</FormDescription>}
          <FormDescription>
            {composeServices.length > 0
              ? "Assign each domain to a compose service. Enter the port if it wasn't auto-detected."
              : 'Enter a domain, the service name from your compose file, and the port it listens on.'}
          </FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
