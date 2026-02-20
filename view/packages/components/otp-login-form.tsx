import { cn } from '@/lib/utils';
import { Button } from '@nixopus/ui';
import { Card, CardContent } from '@nixopus/ui';
import { Input } from '@nixopus/ui';
import { Label } from '@nixopus/ui';
import { Alert, AlertDescription } from '@nixopus/ui';
import { OTPInput } from '@nixopus/ui';
import nixopusLogo from '@/public/logo_white.png';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
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
      <Card className="overflow-hidden p-0 min-h-[500px] flex flex-col justify-center">
        <CardContent className="grid p-0 md:grid-cols-2">
          <div className="p-6 md:p-8">
            <div className="flex flex-col gap-6">
              {props.otpSent && (
                <div className="flex flex-col items-center text-center">
                  <h1 className="text-2xl font-bold">{t('auth.otpLogin.verifyTitle')}</h1>
                  <p className="text-muted-foreground text-balance">
                    {t('auth.otpLogin.verifyDescription')}
                  </p>
                </div>
              )}
              {!props.otpSent && (
                <div className="flex flex-col items-center text-center">
                  <h1 className="text-2xl font-bold">Login with OTP</h1>
                </div>
              )}
              {!props.otpSent && (
                <div className="grid gap-3">
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
            </div>
          </div>
          <div className="bg-muted relative hidden md:block">
            <img
              src={nixopusLogo.src}
              alt="Nixopus Logo"
              className="absolute inset-0 h-full w-full object-contain p-8"
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
