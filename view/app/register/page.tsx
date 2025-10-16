'use client';

import { useTranslation } from '@/hooks/use-translation';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import nixopusLogo from '@/public/nixopus_logo_transparent.png';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { PasswordInputField } from '@/components/ui/password-input-field';
import { signUp } from 'supertokens-auth-react/recipe/emailpassword';
import { toast } from 'sonner';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { LogIn } from 'lucide-react';
import { useIsAdminRegisteredQuery } from '@/redux/services/users/authApi';
import { Skeleton } from '@/components/ui/skeleton';

const registerSchema = (t: (key: string) => string) =>
  z
    .object({
      email: z.string().email(t('auth.register.errors.invalidEmail')),
      password: z
        .string()
        .min(8, t('auth.register.errors.passwordRequirements.minLength'))
        .regex(/[A-Z]/, t('auth.register.errors.passwordRequirements.uppercase'))
        .regex(/[a-z]/, t('auth.register.errors.passwordRequirements.lowercase'))
        .regex(/[0-9]/, t('auth.register.errors.passwordRequirements.number'))
        .regex(
          /[!@#$%^&*(),.?":{}|<>]/,
          t('auth.register.errors.passwordRequirements.specialChar')
        ),
      confirmPassword: z.string()
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t('auth.register.errors.passwordMismatch'),
      path: ['confirmPassword']
    });

type RegisterForm = z.infer<ReturnType<typeof registerSchema>>;

export default function RegisterPage() {
  const { t } = useTranslation();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const { data: isAdminRegistered, isLoading: isAdminRegisteredLoading, isError: isAdminRegisteredError } = useIsAdminRegisteredQuery();
  const form = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema(t)),
    defaultValues: {
      email: '',
      password: '',
      confirmPassword: ''
    }
  });

  const onSubmit = async (data: RegisterForm) => {
    setIsLoading(true);
    try {
      const response = await signUp({
        formFields: [
          { id: 'email', value: data.email },
          { id: 'password', value: data.password }
        ]
      });

      if (response.status === 'FIELD_ERROR') {
        response.formFields.forEach(field => {
          toast.error(field.error);
        });
      } else if (response.status === 'SIGN_UP_NOT_ALLOWED') {
        toast.error('Sign up is not allowed');
      } else {
        toast.success(t('auth.register.success'));
        router.push('/auth');
      }
    } catch (error) {
      toast.error(t('auth.register.errors.registerFailed'));
    } finally {
      setIsLoading(false);
    }
  };


  if (isAdminRegisteredLoading) {
    return <AdminRegisteredSkeleton />;
  }

  if (isAdminRegisteredError) {
    return <AdminRegisteredError />;
  }

  if (isAdminRegistered) {
    return <AdminRegistered />;
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <div className={cn('flex flex-col gap-6')}>
          <Card className="overflow-hidden p-0">
            <CardContent className="grid p-0 md:grid-cols-2">
              <div className="p-6 md:p-8">
                <div className="flex flex-col gap-6">
                  <div className="flex flex-col items-center text-center">
                    <h1 className="text-2xl font-bold">{t('auth.register.title')}</h1>
                    <p className="text-muted-foreground text-balance">
                      {t('auth.register.description')}
                    </p>
                  </div>
                  <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                    <div className="grid gap-3">
                      <Label htmlFor="email">{t('auth.email')}</Label>
                      <Input
                        id="email"
                        type="email"
                        placeholder={t('auth.register.emailPlaceholder')}
                        {...form.register('email')}
                      />
                      {form.formState.errors.email && (
                        <p className="text-sm text-destructive">
                          {form.formState.errors.email.message}
                        </p>
                      )}
                    </div>
                    <div className="grid gap-3">
                      <Label htmlFor="password">{t('auth.password')}</Label>
                      <PasswordInputField id="password" type="password" {...form.register('password')} />
                      {form.formState.errors.password && (
                        <p className="text-sm text-destructive">
                          {form.formState.errors.password.message}
                        </p>
                      )}
                    </div>
                    <div className="grid gap-3">
                      <Label htmlFor="confirmPassword">{t('auth.register.confirmPassword')}</Label>
                      <PasswordInputField                       
                        id="confirmPassword"
                        type="password"
                        {...form.register('confirmPassword')}
                      />

                      {form.formState.errors.confirmPassword && (
                        <p className="text-sm text-destructive">
                          {form.formState.errors.confirmPassword.message}
                        </p>
                      )}
                    </div>
                    <Button type="submit" className="w-full" disabled={isLoading}>
                      {isLoading ? t('auth.register.loading') : t('auth.register.submit')}
                    </Button>
                    <div className="text-center text-sm">
                      {t('auth.register.alreadyHaveAccount')}{' '}
                      <Link href="/auth" className="underline underline-offset-4">
                        {t('auth.login.title')}
                      </Link>
                    </div>
                  </form>
                </div>
              </div>
              <div className="bg-muted relative hidden md:block">
                <img
                  src={nixopusLogo.src}
                  alt="Nixopus Logo"
                  className="absolute inset-0 h-full w-full object-cover"
                />
              </div>
            </CardContent>
          </Card>
          <div className="text-muted-foreground text-center text-xs text-balance">
            {t('auth.register.terms')}{' '}
            <a
              href="https://docs.nixopus.com/license"
              className="underline underline-offset-4 hover:text-primary"
              target="_blank"
              rel="noopener noreferrer"
            >
              {t('auth.register.termsOfService')}
            </a>{' '}
            {t('auth.register.and')}{' '}
            <a
              href="https://docs.nixopus.com/privacy-policy"
              className="underline underline-offset-4 hover:text-primary"
              target="_blank"
              rel="noopener noreferrer"
            >
              {t('auth.register.privacyPolicy')}
            </a>
            .
          </div>
        </div>
      </div>
    </div>
  );
}


const AdminRegisteredError = () => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <div className={cn('flex flex-col gap-6')}>
          <Card className="overflow-hidden p-0">
            <CardContent className="p-0">
              <div className="p-6 md:p-8">
                <div className="flex flex-col gap-6">
                  <div className="flex flex-col items-center text-center">
                    <h1 className="text-2xl font-bold">
                      {t('auth.register.errors.somethingWentWrong')}
                    </h1>
                    <p className="text-muted-foreground text-balance mt-4">
                      {t('auth.register.errors.loadingAdminRegistration')}
                    </p>
                  </div>
                  <div className="flex justify-center gap-4 mt-4 mb-4">
                    <Button variant="outline" onClick={() => window.location.reload()}>
                      {t('auth.register.errors.tryAgain')}
                    </Button>
                    <Button variant="outline" onClick={() => router.push('/auth')}>
                      {t('auth.register.errors.loginButton')}
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};


const AdminRegistered = () => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <Card>
        <CardContent>
          <div className="flex flex-col items-center text-center">
            <h1 className="text-2xl font-bold">
              {t('auth.register.errors.adminAlreadyRegistered')}
            </h1>
            <p className="text-muted-foreground text-balance mt-4">
              {t('auth.register.errors.adminAlreadyRegisteredDescription')}
            </p>
          </div>
        </CardContent>
        <CardFooter className="flex justify-center">
          <Button onClick={() => router.push('/auth')}>
            <LogIn className="mr-2 h-4 w-4" />
            {t('auth.register.errors.loginButton')}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
};


const AdminRegisteredSkeleton = () => {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <div className="flex flex-col gap-6">
          <Card className="overflow-hidden p-0">
            <CardContent className="grid p-0 md:grid-cols-2">
              <div className="p-6 md:p-8">
                <div className="flex flex-col gap-6">
                  <div className="flex flex-col items-center text-center">
                    <Skeleton className="h-8 w-48" />
                    <Skeleton className="mt-4 h-4 w-64" />
                  </div>
                  <div className="space-y-4">
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-16" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-20" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-32" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <Skeleton className="h-10 w-full" />
                    <Skeleton className="mx-auto h-4 w-48" />
                  </div>
                </div>
              </div>
              <div className="bg-muted relative hidden md:block">
                <Skeleton className="absolute inset-0 h-full w-full" />
              </div>
            </CardContent>
          </Card>
          <Skeleton className="mx-auto h-4 w-64" />
        </div>
      </div>
    </div>
  );
};
