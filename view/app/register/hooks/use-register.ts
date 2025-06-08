import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useRouter } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import { useIsAdminRegisteredQuery, useRegisterUserMutation } from '@/redux/services/users/authApi';
import { toast } from 'sonner';

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

function useRegister() {
  const { t } = useTranslation();
  const router = useRouter();
  const [registerUser, { isLoading }] = useRegisterUserMutation();
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
    try {
      await registerUser({
        email: data.email,
        password: data.password
      }).unwrap();
      toast.success(t('auth.register.success'));
      router.push('/login');
    } catch (error) {
      toast.error(t('auth.register.errors.registerFailed'));
    }
  };

  return {
    form,
    onSubmit,
    isLoading,
    isAdminRegistered,
    isAdminRegisteredLoading,
    isAdminRegisteredError
  };
}

export default useRegister;
