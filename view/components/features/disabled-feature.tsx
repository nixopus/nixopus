import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';

function DisabledFeature() {
  const { t } = useTranslation();
  return (
    <div className="flex h-[calc(100vh-200px)] items-center justify-center p-4">
      <Card className="w-full max-w-md p-6 text-center border-none">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-muted">
          <AlertCircle className="h-6 w-6 text-muted-foreground" />
        </div>
        <h2 className="mb-2 text-2xl font-semibold tracking-tight">
          {t('common.featureDisabled')}
        </h2>
        <p className="mb-6 text-sm text-muted-foreground">{t('common.featureNotAvailable')}</p>
        <div className="space-y-3">
          <Button variant="outline" className="w-full" onClick={() => window.history.back()}>
            {t('common.goBack')}
          </Button>
          <Button variant="ghost" className="w-full" onClick={() => window.location.reload()}>
            {t('common.refreshPage')}
          </Button>
        </div>
      </Card>
    </div>
  );
}

export default DisabledFeature;
