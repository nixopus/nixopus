import { cn } from '@/lib/utils';
import { Button } from '@nixopus/ui';
import { Card, CardContent } from '@nixopus/ui';
import { Input } from '@nixopus/ui';
import { Alert, AlertDescription } from '@nixopus/ui';
import { OTPInput } from '@nixopus/ui';
import { useTranslation, type translationKey } from '@/packages/hooks/shared/use-translation';
import { useOtpLoginForm } from '@/packages/hooks/auth/use-otp-login-form';

export interface OtpLoginFormProps {
  email: string;
  otp: string;
  handleEmailChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handleOtpChange: (otp: string) => void;
  handleSendOtp: () => void;
  handleVerifyOtp: () => void;
  handleChangeEmail: () => void;
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
    <div className={cn('flex flex-col gap-4')}>
      <div className="flex justify-center">
        <img
          src="/logo_black.png"
          alt="Nixopus Logo"
          className="max-h-16 max-w-16 object-contain dark:hidden"
        />
        <img
          src="/logo_white.png"
          alt="Nixopus Logo"
          className="max-h-16 max-w-16 object-contain hidden dark:block"
        />
      </div>
      <Card className="overflow-hidden p-0 flex flex-col justify-center border-0 shadow-none">
        <CardContent className="p-6 md:p-8">
          <div className="flex flex-col gap-4">
            {!props.otpSent && (
              <div className="grid gap-2">
                <Input
                  id="email"
                  type="email"
                  placeholder={t('auth.login.emailPlaceholder')}
                  required
                  value={props.email}
                  onChange={props.handleEmailChange}
                  className="h-11 placeholder:text-gray-500"
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
              <>
                <p className="text-center text-sm text-muted-foreground">
                  {t('auth.otpLogin.otpSentTo')} {props.email}
                </p>
                <div className="grid gap-2">
                  <OTPInput
                    value={props.otp}
                    onChange={props.handleOtpChange}
                    length={6}
                    disabled={props.isVerifyingOtp}
                    className="[&_input]:h-11"
                  />
                  {otpError && (
                    <Alert variant="destructive">
                      <AlertDescription className="text-xs !text-red-600 font-medium">
                        {otpError}
                      </AlertDescription>
                    </Alert>
                  )}
                </div>
              </>
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
              <div className="flex items-center justify-between">
                <button
                  type="button"
                  onClick={handleSendOtpClick}
                  disabled={props.isSendingOtp}
                  className="text-sm text-muted-foreground hover:text-primary hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {props.isSendingOtp ? t('auth.otpLogin.sending') : t('auth.otpLogin.resendOtp')}
                </button>
                <Button
                  type="button"
                  variant="link"
                  className="h-auto p-0 text-sm text-muted-foreground hover:text-primary hover:underline"
                  onClick={props.handleChangeEmail}
                >
                  {t('auth.otpLogin.changeEmail' as translationKey)}
                </Button>
              </div>
            )}
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
