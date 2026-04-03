import { cn } from '@/lib/utils';
import { Button } from '@nixopus/ui';
import { Card, CardContent } from '@nixopus/ui';
import { Input } from '@nixopus/ui';
import { Alert, AlertDescription } from '@nixopus/ui';
import { OTPInput } from '@nixopus/ui';
import { Label } from '@nixopus/ui';
import { useTranslation, type translationKey } from '@/packages/hooks/shared/use-translation';
import { useOtpLoginForm } from '@/packages/hooks/auth/use-otp-login-form';
import { useState, useCallback, useEffect, useRef, type ComponentType } from 'react';
import { Mail } from 'lucide-react';
import Image from 'next/image';
import { useTheme } from 'next-themes';
import type { CaptchaComponentProps } from '@/plugins/registry';

export interface OtpLoginFormProps {
  email: string;
  otp: string;
  handleEmailChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handleOtpChange: (otp: string) => void;
  handleSendOtp: (captchaToken?: string) => void;
  handleVerifyOtp: (captchaToken?: string) => void;
  handleChangeEmail: () => void;
  isSendingOtp: boolean;
  isVerifyingOtp: boolean;
  otpSent: boolean;
  timer?: number;
  formatTimer?: (seconds: number) => string;
  CaptchaComponent?: ComponentType<CaptchaComponentProps>;
  captchaSiteKey?: string;
}

export function OtpLoginForm({ CaptchaComponent, captchaSiteKey, ...props }: OtpLoginFormProps) {
  const { t } = useTranslation();
  const { resolvedTheme } = useTheme();
  const [captchaToken, setCaptchaToken] = useState<string | null>(null);
  const captchaEnabled = !!CaptchaComponent && !!captchaSiteKey;
  const hasVerifiedRef = useRef(false);
  const otpContainerRef = useRef<HTMLDivElement>(null);

  const handleToken = useCallback((token: string | null) => {
    setCaptchaToken(token);
  }, []);

  const { emailError, otpError, handleSendOtpClick, handleVerifyOtpClick } = useOtpLoginForm({
    email: props.email,
    otp: props.otp,
    handleSendOtp: () => {
      props.handleSendOtp(captchaToken ?? undefined);
      setCaptchaToken(null);
    },
    handleVerifyOtp: () => {
      props.handleVerifyOtp(captchaToken ?? undefined);
      setCaptchaToken(null);
    },
    isVerifyingOtp: props.isVerifyingOtp
  });

  useEffect(() => {
    if (!props.otpSent || !props.otp || props.otp.length === 0) {
      hasVerifiedRef.current = false;
    }
  }, [props.otp, props.otpSent]);

  const handleOtpVerifyWithToken = useCallback(() => {
    props.handleVerifyOtp(captchaToken ?? undefined);
    setCaptchaToken(null);
  }, [props.handleVerifyOtp, captchaToken]);

  useEffect(() => {
    if (
      props.otpSent &&
      props.otp &&
      props.otp.length === 6 &&
      !props.isVerifyingOtp &&
      !hasVerifiedRef.current &&
      (!captchaEnabled || captchaToken)
    ) {
      hasVerifiedRef.current = true;
      handleOtpVerifyWithToken();
    }
  }, [
    props.otp,
    props.otpSent,
    props.isVerifyingOtp,
    captchaEnabled,
    captchaToken,
    handleOtpVerifyWithToken
  ]);

  useEffect(() => {
    if (!props.otpSent) return;
    const timer = setTimeout(() => {
      const input = otpContainerRef.current?.querySelector<HTMLInputElement>('input');
      input?.focus();
    }, 0);
    return () => clearTimeout(timer);
  }, [props.otpSent]);

  const handleFormSubmit = useCallback(
    (e: React.FormEvent<HTMLFormElement>): void => {
      e.preventDefault();
      if (!props.otpSent) {
        handleSendOtpClick();
      }
    },
    [props.otpSent, handleSendOtpClick]
  );

  return (
    <div className={cn('flex flex-col gap-6')}>
      <div className="flex justify-center w-full">
        <Card className="w-full max-w-md">
          <CardContent className="p-6 md:p-8">
            <div className="flex justify-center mb-4">
              <Image
                src={resolvedTheme === 'dark' ? '/logo_white.png' : '/logo_black.png'}
                alt="Nixopus"
                width={96}
                height={96}
                priority
                className="shrink-0"
              />
            </div>
            <form onSubmit={handleFormSubmit} className="flex flex-col gap-6">
              {props.otpSent ? (
                <div className="grid gap-3">
                  <div className="flex items-center justify-between">
                    <Label className="text-center flex-1">
                      {t('auth.otpLogin.otpSentTo')} {props.email}
                    </Label>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={props.handleChangeEmail}
                      className="text-xs h-auto py-1 px-2 text-muted-foreground hover:text-foreground"
                    >
                      {t('auth.otpLogin.changeEmail' as translationKey)}
                    </Button>
                  </div>
                  <div ref={otpContainerRef}>
                    <OTPInput
                      value={props.otp}
                      onChange={props.handleOtpChange}
                      length={6}
                      disabled={props.isVerifyingOtp}
                    />
                  </div>
                  {props.timer !== undefined && props.timer > 0 && (
                    <div className="flex items-center justify-center text-sm text-muted-foreground">
                      <span>
                        {t('auth.otpLogin.expiresIn' as translationKey)}{' '}
                        <span className="font-semibold text-foreground">
                          {props.formatTimer
                            ? props.formatTimer(props.timer)
                            : `${Math.floor(props.timer / 60)}:${(props.timer % 60).toString().padStart(2, '0')}`}
                        </span>
                      </span>
                    </div>
                  )}
                  {props.timer === 0 && props.otpSent && (
                    <div className="text-center text-sm text-muted-foreground">
                      {t('auth.otpLogin.expired' as translationKey)}
                    </div>
                  )}
                  {otpError && (
                    <Alert variant="destructive">
                      <AlertDescription className="text-xs !text-red-600 font-medium">
                        {otpError}
                      </AlertDescription>
                    </Alert>
                  )}
                  {captchaEnabled && CaptchaComponent && (
                    <div className="flex justify-center">
                      <CaptchaComponent siteKey={captchaSiteKey!} onToken={handleToken} />
                    </div>
                  )}
                </div>
              ) : (
                <div className="grid gap-3">
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      id="email"
                      type="email"
                      placeholder={t('auth.login.emailPlaceholder')}
                      required
                      value={props.email}
                      onChange={props.handleEmailChange}
                      className="pl-10 h-12"
                    />
                  </div>
                  {emailError && (
                    <Alert variant="destructive">
                      <AlertDescription className="text-xs !text-red-600 font-medium">
                        {emailError}
                      </AlertDescription>
                    </Alert>
                  )}
                  {captchaEnabled && CaptchaComponent && (
                    <div className="flex justify-center">
                      <CaptchaComponent siteKey={captchaSiteKey!} onToken={handleToken} />
                    </div>
                  )}
                </div>
              )}
              <Button
                type="submit"
                className="w-full"
                onClick={props.otpSent ? handleOtpVerifyWithToken : handleSendOtpClick}
                disabled={
                  props.otpSent
                    ? props.isVerifyingOtp ||
                      !props.otp ||
                      props.otp.length !== 6 ||
                      props.timer === 0 ||
                      (captchaEnabled && !captchaToken)
                    : props.isSendingOtp || (captchaEnabled && !captchaToken)
                }
              >
                {props.otpSent
                  ? props.isVerifyingOtp
                    ? t('auth.otpLogin.verifying')
                    : t('auth.otpLogin.verifyButton')
                  : props.isSendingOtp
                    ? t('auth.otpLogin.sending')
                    : t('auth.otpLogin.sendOtpButton')}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
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
