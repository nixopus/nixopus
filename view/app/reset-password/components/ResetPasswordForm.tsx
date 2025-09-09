import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { useResetPasswordMutation } from '@/redux/services/users/authApi';
import { toast } from 'sonner';
import { useRouter } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import { PasswordInputField } from '@/components/ui/password-input-field';

const resetPasswordSchema = (t: (key: string) => string) =>
  z
    .object({
      password: z
        .string()
        .min(8, t('auth.resetPassword.errors.passwordRequirements.minLength'))
        .regex(/[A-Z]/, t('auth.resetPassword.errors.passwordRequirements.uppercase'))
        .regex(/[a-z]/, t('auth.resetPassword.errors.passwordRequirements.lowercase'))
        .regex(/[0-9]/, t('auth.resetPassword.errors.passwordRequirements.number'))
        .regex(
          /[!@#$%^&*(),.?":{}|<>]/,
          t('auth.resetPassword.errors.passwordRequirements.specialChar')
        ),
      confirmPassword: z.string()
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t('auth.resetPassword.errors.passwordMismatch'),
      path: ['confirmPassword']
    });

type ResetPasswordForm = z.infer<ReturnType<typeof resetPasswordSchema>>;

interface ResetPasswordFormProps {
  token: string | null;
}

export function ResetPasswordForm({ token }: ResetPasswordFormProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const [resetPassword, { isLoading }] = useResetPasswordMutation();

  const form = useForm<ResetPasswordForm>({
    resolver: zodResolver(resetPasswordSchema(t)),
    defaultValues: {
      password: '',
      confirmPassword: ''
    }
  });

  const onSubmit = async (data: ResetPasswordForm) => {
    if (!token) {
      toast.error(t('auth.resetPassword.errors.invalidLink'));
      return;
    }

    try {
      await resetPassword({ token, password: data.password }).unwrap();
      toast.success(t('auth.resetPassword.success'));
      router.push('/login');
    } catch (error) {
      toast.error(t('auth.resetPassword.errors.resetFailed'));
    }
  };

  if (!token) {
    return (
      <Card className="w-full max-w-md">
        <CardContent className="p-6">
          <h1 className="text-2xl font-bold text-center mb-4">
            {t('auth.resetPassword.errors.invalidLink')}
          </h1>
          <p className="text-center text-muted-foreground">{t('auth.resetPassword.description')}</p>
          <Button className="w-full mt-4" onClick={() => router.push('/login')}>
            {t('auth.login.title')}
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md">
      <CardContent className="p-6">
        <h1 className="text-2xl font-bold text-center mb-4">{t('auth.resetPassword.title')}</h1>
        <p className="text-center text-muted-foreground mb-6">
          {t('auth.resetPassword.description')}
        </p>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('auth.resetPassword.newPassword')}</FormLabel>
                  <FormControl>
                    <PasswordInputField
                      type="password"
                      placeholder={t('auth.resetPassword.newPassword')}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="confirmPassword"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('auth.resetPassword.confirmPassword')}</FormLabel>
                  <FormControl>
                    <PasswordInputField
                      type="password"
                      placeholder={t('auth.resetPassword.confirmPassword')}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? t('auth.resetPassword.submitting') : t('auth.resetPassword.submit')}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
