import React from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useForm } from 'react-hook-form';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import {
  useCreateDomainMutation,
  useUpdateDomainMutation
} from '@/redux/services/settings/domainsApi';
import { Domain } from '@/redux/types/domain';

const domainFormSchema = z.object({
  domainName: z
    .string()
    .min(3, { message: 'Domain name must be at least 3 characters.' })
    .refine(
      (domain) => /^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9](?:\.[a-zA-Z]{2,})+$/.test(domain),
      { message: 'Please enter a valid domain name (e.g., example.com).' }
    )
});

interface UpdateDomainDialogProps {
  open: boolean;
  setOpen: React.Dispatch<React.SetStateAction<boolean>>;
  id?: string;
  data?: Domain;
}

function UpdateDomainDialog({ open, setOpen, id, data }: UpdateDomainDialogProps) {
  const [createDomain, { isLoading }] = useCreateDomainMutation();
  const [updateDomain, { isLoading: isUpdating }] = useUpdateDomainMutation();
  const form = useForm({
    resolver: zodResolver(domainFormSchema),
    defaultValues: {
      domainName: data?.name || ''
    }
  });

  async function onSubmit(data: z.infer<typeof domainFormSchema>) {
    try {
      if (!id) {
        await createDomain({ name: data.domainName });
      } else {
        await updateDomain({ name: data.domainName, id: id });
      }
      toast.success('Domain added successfully');
    } catch (error) {
      toast.error('Failed to add domain');
    } finally {
      form.reset();
      setOpen(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {!id && (
        <DialogTrigger asChild>
          <Button variant="outline">Add Domain</Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{!id ? 'Add Domain' : 'Update Domain'}</DialogTitle>
          <DialogDescription>Domain will help to deploy your applications</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <FormField
              control={form.control}
              name="domainName"
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <Input placeholder="example.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter className="flex justify-between sm:justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  form.reset();
                  setOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading || isUpdating}>
                {isLoading || isUpdating ? 'Saving...' : 'Save'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

export default UpdateDomainDialog;
