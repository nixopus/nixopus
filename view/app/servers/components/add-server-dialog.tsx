'use client';

import React from 'react';
import { Server, Tag, Network } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import { useAddServerDialog, type ServerFormValues } from '../hooks/use-add-server-dialog';
import type { ControllerRenderProps } from 'react-hook-form';
import { cn } from '@/lib/utils';
import type {
  AddServerDialogProps,
  ServerFormFieldProps,
  BasicFieldsSectionProps,
  AuthMethodTabsProps,
  AdvancedFieldsSectionProps,
  AuthMethod
} from '../types';

function ServerFormField({ config, form }: ServerFormFieldProps<ServerFormValues>) {
  const renderInput = (field: ControllerRenderProps<ServerFormValues, keyof ServerFormValues>) => {
    const baseProps = {
      placeholder: config.placeholder,
      ...field
    };

    switch (config.type) {
      case 'textarea':
        return <Textarea {...baseProps} className="font-mono text-xs h-32" />;
      case 'password':
        return <Input {...baseProps} type="password" />;
      case 'number':
        return <Input {...baseProps} type="number" />;
      default:
        if (config.name === 'host') {
          return (
            <div className="relative">
              <Network className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input {...baseProps} className="pl-10" />
            </div>
          );
        }
        return <Input {...baseProps} />;
    }
  };

  return (
    <FormField
      control={form.control}
      name={config.name as keyof ServerFormValues}
      render={({ field }) => (
        <FormItem>
          <FormLabel className={config.name === 'labels' ? 'flex items-center gap-2' : undefined}>
            {config.name === 'labels' && <Tag className="h-4 w-4" />}
            {config.label}
          </FormLabel>
          <FormControl>{renderInput(field)}</FormControl>
          {config.description && <FormDescription>{config.description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}

function BasicFieldsSection({ fields, form }: BasicFieldsSectionProps<ServerFormValues>) {
  return (
    <div className="grid grid-cols-3 gap-4">
      {fields.map((field) => (
        <div
          key={field.name}
          className={cn(field.colSpan === 2 && 'col-span-2', field.colSpan === 3 && 'col-span-3')}
        >
          <ServerFormField config={field} form={form} />
        </div>
      ))}
    </div>
  );
}

function AuthMethodTabs({
  authMethod,
  setAuthMethod,
  authTabs,
  authMethodLabel,
  form
}: AuthMethodTabsProps<ServerFormValues>) {
  return (
    <div className="space-y-4">
      <Label className="text-sm font-medium">{authMethodLabel}</Label>
      <Tabs value={authMethod} onValueChange={(v) => setAuthMethod(v as AuthMethod)}>
        <TabsList className="grid w-full grid-cols-2">
          {authTabs.map((tab) => (
            <TabsTrigger key={tab.value} value={tab.value} className="gap-2">
              <tab.icon className="h-4 w-4" />
              {tab.label}
            </TabsTrigger>
          ))}
        </TabsList>
        {authTabs.map((tab) => (
          <TabsContent key={tab.value} value={tab.value} className="pt-4">
            <ServerFormField config={tab.field} form={form} />
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}

function AdvancedFieldsSection({ fields, form }: AdvancedFieldsSectionProps<ServerFormValues>) {
  return (
    <div className="space-y-4">
      {fields.map((field) => (
        <ServerFormField key={field.name} config={field} form={form} />
      ))}
    </div>
  );
}

export function AddServerDialog({ open, onOpenChange, onSubmit }: AddServerDialogProps) {
  const {
    form,
    authMethod,
    setAuthMethod,
    actions,
    basicFields,
    advancedFields,
    authTabs,
    authMethodLabel,
    t
  } = useAddServerDialog({ onSubmit, onOpenChange });

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title={
        <span className="flex items-center gap-2">
          <Server className="h-5 w-5" />
          {t('servers.dialog.add.title')}
        </span>
      }
      description={t('servers.dialog.add.description')}
      actions={actions}
      size="lg"
    >
      <Form {...form}>
        <form className="space-y-6">
          <BasicFieldsSection fields={basicFields} form={form} />

          <AuthMethodTabs
            authMethod={authMethod}
            setAuthMethod={setAuthMethod}
            authTabs={authTabs}
            authMethodLabel={authMethodLabel}
            form={form}
          />

          <AdvancedFieldsSection fields={advancedFields} form={form} />
        </form>
      </Form>
    </DialogWrapper>
  );
}
