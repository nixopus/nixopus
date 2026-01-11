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
import { useAppDispatch } from '@/redux/hooks';
import { initializeAuth } from '@/redux/features/users/authSlice';

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
  const dispatch = useAppDispatch();
  const [isLoading, setIsLoading] = useState(false);
  const [registrationSuccess, setRegistrationSuccess] = useState(false);
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

  const isNetworkError = (error: unknown): boolean => {
    if (error instanceof Error) {
      return (
        error.message.includes('network') ||
        error.message.includes('fetch') ||
        error.message.includes('Failed to fetch') ||
        error.name === 'NetworkError'
      );
    }
    return false;
  };

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
        toast.error(t('auth.register.errors.signUpNotAllowed.message' as any), {
          description: t('auth.register.errors.signUpNotAllowed.description' as any)
        });
      } else {
        setRegistrationSuccess(true);
        toast.success(t('auth.register.successAdmin.title' as any), {
          description: t('auth.register.successAdmin.message' as any)
        });
        await dispatch(initializeAuth() as any);
        // Note: User is already logged in after signUp, so we'll redirect to dashboard
        // The success component will handle the redirect after showing the success message
      }
    } catch (error) {
      if (isNetworkError(error)) {
        toast.error(t('auth.register.errors.networkError.title' as any), {
          description: t('auth.register.errors.networkError.description' as any)
        });
      } else if (error instanceof Error) {
        const errorMessage = error.message.toLowerCase();
        if (errorMessage.includes('server') || errorMessage.includes('500')) {
          toast.error(t('auth.register.errors.serverError.title' as any), {
            description: t('auth.register.errors.serverError.description' as any)
          });
        } else {
          toast.error(t('auth.register.errors.registerFailed'), {
            description: error.message || t('auth.register.errors.unknownError' as any)
          });
        }
      } else {
        toast.error(t('auth.register.errors.registerFailed'));
      }
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
    registrationSuccess,
    t
  };
}

export default useRegister;
