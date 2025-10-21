'use client';

import { useTranslation, type translationKey } from '@/hooks/use-translation';
import { useRouter } from 'next/navigation';
import { signUp } from 'supertokens-auth-react/recipe/emailpassword';
import { toast } from 'sonner';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useIsAdminRegisteredQuery } from '@/redux/services/users/authApi';

const registerSchema = (t: (key: translationKey, params?: Record<string, string>) => string) =>
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

function useRegister() {
  const { t } = useTranslation();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const {
    data: isAdminRegistered,
    isLoading: isAdminRegisteredLoading,
    isError: isAdminRegisteredError
  } = useIsAdminRegisteredQuery();
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
        response.formFields.forEach((field) => {
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

  return {
    form,
    onSubmit,
    isLoading,
    isAdminRegistered,
    isAdminRegisteredLoading,
    isAdminRegisteredError,
    t
  };
}

export default useRegister;
