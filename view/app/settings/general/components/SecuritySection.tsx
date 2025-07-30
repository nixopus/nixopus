'use client';

import React from 'react';
import { Lock, CheckCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TabsContent } from '@/components/ui/tabs';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { useTranslation } from '@/hooks/use-translation';
import { TwoFactorSetup } from '@/app/settings/general/components/TwoFactorSetup';
import { RBACGuard } from '@/components/rbac/RBACGuard';

interface SecuritySectionProps {
  emailSent: boolean;
  isLoading: boolean;
  handlePasswordResetRequest: () => void;
}

function SecuritySection({
  emailSent,
  isLoading,
  handlePasswordResetRequest
}: SecuritySectionProps) {
  const { t } = useTranslation();

  return (
    <TabsContent value="security" className="space-y-4 mt-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('settings.security.password.title')}</CardTitle>
          <CardDescription>{t('settings.security.password.description')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label className="flex items-center gap-2">
              <Lock size={16} />
              {t('settings.security.password.reset.label')}
            </Label>
            <p className="text-sm text-muted-foreground">
              {t('settings.security.password.reset.description')}
            </p>
          </div>

          {emailSent ? (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertTitle>{t('settings.security.password.reset.emailSent.title')}</AlertTitle>
              <AlertDescription>
                {t('settings.security.password.reset.emailSent.description')}
              </AlertDescription>
            </Alert>
          ) : (
            <RBACGuard resource="user" action="update">
              <Button
                onClick={handlePasswordResetRequest}
                disabled={isLoading}
                variant="outline"
                className="w-full lg:w-auto"
              >
                {t('settings.security.password.reset.button')}
              </Button>
            </RBACGuard>
          )}
        </CardContent>
      </Card>

      <TwoFactorSetup />
    </TabsContent>
  );
}

export default SecuritySection;
