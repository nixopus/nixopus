import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { PasswordInputField } from '@/components/ui/password-input-field';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import nixopusLogo from '@/public/nixopus_logo_transparent.png';
import { useTranslation } from '@/hooks/use-translation';
import Link from 'next/link';
import { useState } from 'react';
import { z } from 'zod';

export interface LoginFormProps {
  email: string;
  password: string;
  handleEmailChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handlePasswordChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handleLogin: () => void;
  isLoading: boolean;
  twoFactorCode?: string;
  handleTwoFactorCodeChange?: (event: React.ChangeEvent<HTMLInputElement>) => void;
  showTwoFactor?: boolean;
  handleTwoFactorLogin?: () => void;
  isTwoFactorLoading?: boolean;
}

export function LoginForm({ ...props }: LoginFormProps) {
  const { t } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const [emailError, setEmailError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  // Zod schema for login validation
  const loginSchema = z.object({
    email: z.string().min(1, 'Email is required').email('Please enter a valid Email'),
    password: z.string().min(1, 'Password is required')
  });

  const handleLoginClick = (): void => {
    // Reset errors first
    setEmailError('');
    setPasswordError('');

    const result = loginSchema.safeParse({
      email: props.email ?? '',
      password: props.password ?? ''
    });

    if (!result.success) {
      const fieldErrors = result.error.flatten().fieldErrors;
      if (fieldErrors.email && fieldErrors.email[0]) {
        setEmailError(fieldErrors.email[0]);
      }
      if (fieldErrors.password && fieldErrors.password[0]) {
        setPasswordError(fieldErrors.password[0]);
      }
      return;
    }

    props.handleLogin();
  };

  return (
    <div className={cn('flex flex-col gap-6')}>
      <Card className="overflow-hidden p-0">
        <CardContent className="grid p-0 md:grid-cols-2">
          <div className="p-6 md:p-8">
            <div className="flex flex-col gap-6">
              <div className="flex flex-col items-center text-center">
                <h1 className="text-2xl font-bold">
                  {props.showTwoFactor ? t('auth.login.2fa.title') : t('auth.login.title')}
                </h1>
                <p className="text-muted-foreground text-balance">
                  {props.showTwoFactor
                    ? t('auth.login.2fa.description')
                    : t('auth.login.description')}
                </p>
              </div>
              {!props.showTwoFactor && (
                <>
                  <div className="grid gap-3">
                    <Label htmlFor="email">{t('auth.email')}</Label>
                    <Input
                      id="email"
                      type="email"
                      placeholder={t('auth.login.emailPlaceholder')}
                      required
                      value={props.email}
                      onChange={props.handleEmailChange}
                    />
                    {emailError && (
                      <Alert variant="destructive">
                        <AlertDescription className="text-xs !text-red-600 font-medium">
                          {emailError}
                        </AlertDescription>
                      </Alert>
                    )}
                  </div>
                  <div className="grid gap-3">
                    {/* <div className="flex items-center">
                      <Label htmlFor="password">Password</Label>
                      <a href="#" className="ml-auto text-sm underline-offset-2 hover:underline">
                        Forgot your password?
                      </a>
                    </div> */}
                    <Label htmlFor="password">{t('auth.password')}</Label>
                    <PasswordInputField
                      id="password"
                      required
                      value={props.password}
                      onChange={props.handlePasswordChange}
                    />
                    {passwordError && (
                      <Alert variant="destructive">
                        <AlertDescription className="text-xs !text-red-600 font-medium">
                          {passwordError}
                        </AlertDescription>
                      </Alert>
                    )}
                  </div>
                </>
              )}
              {props.showTwoFactor && (
                <div className="grid gap-3">
                  <Label htmlFor="twoFactorCode">{t('auth.login.2fa.enterCode')}</Label>
                  <Input
                    id="twoFactorCode"
                    type="text"
                    placeholder={t('auth.login.2fa.codePlaceholder')}
                    required
                    value={props.twoFactorCode}
                    onChange={props.handleTwoFactorCodeChange}
                  />
                </div>
              )}
              <Button
                type="submit"
                className="w-full"
                onClick={props.showTwoFactor ? props.handleTwoFactorLogin : handleLoginClick}
                disabled={props.showTwoFactor ? props.isTwoFactorLoading : props.isLoading}
              >
                {props.showTwoFactor
                  ? props.isTwoFactorLoading
                    ? t('auth.login.2fa.verifying')
                    : t('auth.login.2fa.verifyButton')
                  : props.isLoading
                    ? t('auth.login.loading')
                    : t('auth.login.submit')}
              </Button>
              <div className="text-center text-sm">
                Don&apos;t have an account?{' '}
                <Link href="/register" className="underline underline-offset-4">
                  {t('auth.register.title')}
                </Link>
              </div>
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
        {t('auth.login.terms')}{' '}
        <a
          href="https://docs.nixopus.com/license"
          className="underline underline-offset-4 hover:text-primary"
        >
          {t('auth.login.termsOfService')}
        </a>{' '}
        {t('auth.login.and')}{' '}
        <a
          href="https://docs.nixopus.com/privacy-policy"
          className="underline underline-offset-4 hover:text-primary"
        >
          {t('auth.login.privacyPolicy')}
        </a>
        .
      </div>
    </div>
  );
}
