'use client';
import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Loader2 } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import useGithubCallback from '@/packages/hooks/applications/use_github_callback';

const Page = () => {
  const { status, error } = useGithubCallback();
  const { t } = useTranslation();

  if (status === 'processing') {
    return (
      <Card className="w-[350px]">
        <CardContent className="flex flex-col items-center justify-center py-8">
          <Loader2 className="mb-4 h-8 w-8 animate-spin" />
          <p>{t('selfHost.githubCallback.processing')}</p>
        </CardContent>
      </Card>
    );
  }

  if (status === 'success') {
    return (
      <Card className="w-[350px]">
        <CardContent className="py-6">
          <Alert>
            <AlertDescription className="text-green-600">
              {t('selfHost.githubCallback.success')}
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-[350px]">
      <CardContent className="py-6">
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </CardContent>
    </Card>
  );
};

export default Page;
