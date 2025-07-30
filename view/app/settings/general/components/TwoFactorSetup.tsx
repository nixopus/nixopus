import { useState } from 'react';
import {
  useSetupTwoFactorMutation,
  useVerifyTwoFactorMutation,
  useDisableTwoFactorMutation
} from '@/redux/services/users/authApi';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';
import { QRCodeSVG } from 'qrcode.react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Shield, Smartphone, CheckCircle2 } from 'lucide-react';
import { useAppSelector } from '@/redux/hooks';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { setTwoFactorEnabled } from '@/redux/features/users/authSlice';
import { useAppDispatch } from '@/redux/hooks';
import { RBACGuard } from '@/components/rbac/RBACGuard';

export function TwoFactorSetup() {
  const { t } = useTranslation();
  const [code, setCode] = useState('');
  const [setupTwoFactor, { data: setupData }] = useSetupTwoFactorMutation();
  const [verifyTwoFactor] = useVerifyTwoFactorMutation();
  const [disableTwoFactor] = useDisableTwoFactorMutation();
  const user = useAppSelector((state) => state.auth.user);
  const dispatch = useAppDispatch();
  const handleSetup = async () => {
    try {
      const response = await setupTwoFactor().unwrap();
      dispatch(setTwoFactorEnabled(true));
      toast.success(t('settings.2fa.setupSuccess'));
    } catch (error) {
      toast.error(t('settings.2fa.setupError'));
    }
  };

  const handleVerify = async () => {
    try {
      await verifyTwoFactor({ code }).unwrap();
      toast.success(t('settings.2fa.verifySuccess'));
    } catch (error) {
      toast.error(t('settings.2fa.verifyError'));
    }
  };

  const handleDisable = async () => {
    try {
      await disableTwoFactor().unwrap();
      toast.success(t('settings.2fa.disableSuccess'));
    } catch (error) {
      toast.error(t('settings.2fa.disableError'));
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          {t('settings.2fa.title')}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {user?.two_factor_enabled ? (
          <div className="space-y-6">
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertTitle>{t('settings.2fa.enabledTitle')}</AlertTitle>
              <AlertDescription>{t('settings.2fa.enabledDescription')}</AlertDescription>
            </Alert>

            <RBACGuard resource="user" action="update">
              <Button onClick={handleDisable} variant="destructive" className="w-full">
                {t('settings.2fa.disableButton')}
              </Button>
            </RBACGuard>
          </div>
        ) : !setupData ? (
          <div className="space-y-4">
            <Alert>
              <Smartphone className="h-4 w-4" />
              <AlertTitle>{t('settings.2fa.title')}</AlertTitle>
              <AlertDescription>{t('settings.2fa.description')}</AlertDescription>
            </Alert>
            <RBACGuard resource="user" action="update">
              <Button onClick={handleSetup} className="w-full">
                {t('settings.2fa.setupButton')}
              </Button>
            </RBACGuard>
          </div>
        ) : (
          <div className="space-y-6">
            <div className="flex flex-col items-center space-y-4">
              <div className="rounded-lg border p-4">
                <QRCodeSVG value={setupData.qr_code} size={200} level="H" includeMargin={true} />
              </div>
              <p className="text-sm text-muted-foreground text-center">
                {t('settings.2fa.description')}
              </p>
            </div>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label>{t('settings.2fa.enterCode')}</Label>
                <Input
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder={t('settings.2fa.codePlaceholder')}
                  className="w-full"
                />
              </div>
              <RBACGuard resource="user" action="update">
                <Button
                  onClick={handleVerify}
                  className="w-full"
                  disabled={!code || code.length !== 6}
                >
                  {t('settings.2fa.verifyButton')}
                </Button>
              </RBACGuard>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
