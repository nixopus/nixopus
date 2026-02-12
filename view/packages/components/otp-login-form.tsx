import { cn } from '@/lib/utils';
import { Button } from '@nixopus/ui';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@nixopus/ui';
import { Input } from '@nixopus/ui';
import { Label } from '@nixopus/ui';
import { Alert, AlertDescription } from '@nixopus/ui';
import { OTPInput } from '@nixopus/ui';
import nixopusLogo from '@/public/nixopus_logo_transparent.png';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import Link from 'next/link';
import { useOtpLoginForm } from '@/packages/hooks/auth/use-otp-login-form';

export interface OtpLoginFormProps {
  email: string;
  otp: string;
  handleEmailChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handleOtpChange: (otp: string) => void;
  handleSendOtp: () => void;
  handleVerifyOtp: () => void;
  isSendingOtp: boolean;
  isVerifyingOtp: boolean;
  otpSent: boolean;
}

export function OtpLoginForm({ ...props }: OtpLoginFormProps) {
  const { t } = useTranslation();
  const { emailError, otpError, handleSendOtpClick, handleVerifyOtpClick } = useOtpLoginForm({
    email: props.email,
    otp: props.otp,
    handleSendOtp: props.handleSendOtp,
    handleVerifyOtp: props.handleVerifyOtp,
    isVerifyingOtp: props.isVerifyingOtp
  });

  return (
    <div className={cn('flex flex-col gap-6')}>
      <Card className="overflow-hidden p-0">
        <CardContent className="grid p-0 md:grid-cols-2">
          <div className="p-6 md:p-8">
            <div className="flex flex-col gap-6">
              <CardHeader className="p-0 text-center">
                <CardTitle className="text-2xl">
                  {props.otpSent ? t('auth.otpLogin.verifyTitle') : t('auth.otpLogin.title')}
                </CardTitle>
                <CardDescription>
                  {props.otpSent
                    ? t('auth.otpLogin.verifyDescription')
                    : t('auth.otpLogin.description')}
                </CardDescription>
              </CardHeader>
              {!props.otpSent && (
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
              )}
              {props.otpSent && (
                <div className="grid gap-3">
                  <Label htmlFor="otp">{t('auth.otpLogin.enterOtp')}</Label>
                  <OTPInput
                    value={props.otp}
                    onChange={props.handleOtpChange}
                    length={6}
                    disabled={props.isVerifyingOtp}
                  />
                  {otpError && (
                    <Alert variant="destructive">
                      <AlertDescription className="text-xs !text-red-600 font-medium">
                        {otpError}
                      </AlertDescription>
                    </Alert>
                  )}
                  <p className="text-muted-foreground text-sm">
                    {t('auth.otpLogin.otpSentTo')}{' '}
                    <span className="font-medium">{props.email}</span>
                  </p>
                </div>
              )}
              <Button
                type="submit"
                className="w-full"
                onClick={props.otpSent ? handleVerifyOtpClick : handleSendOtpClick}
                disabled={props.otpSent ? props.isVerifyingOtp : props.isSendingOtp}
              >
                {props.otpSent
                  ? props.isVerifyingOtp
                    ? t('auth.otpLogin.verifying')
                    : t('auth.otpLogin.verifyButton')
                  : props.isSendingOtp
                    ? t('auth.otpLogin.sending')
                    : t('auth.otpLogin.sendOtpButton')}
              </Button>
              {props.otpSent && (
                <Button
                  type="button"
                  variant="ghost"
                  className="w-full"
                  onClick={handleSendOtpClick}
                  disabled={props.isSendingOtp}
                >
                  {props.isSendingOtp ? t('auth.otpLogin.sending') : t('auth.otpLogin.resendOtp')}
                </Button>
              )}
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
