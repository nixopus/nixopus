import React from 'react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Github } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

interface GithubInstallProps {
  appSlug: string;
  organization?: string;
  callbackUrl: string;
}

const GithubInstaller = ({ appSlug, organization, callbackUrl }: GithubInstallProps) => {
  const { t } = useTranslation();

  const handleConnectGithub = () => {
    const baseUrl = 'https://github.com';
    const installPath = organization
      ? `/organizations/${organization}/settings/apps/${appSlug}/installations/new`
      : `/apps/${appSlug}/installations/new`;

    const stateParam = encodeURIComponent(crypto.randomUUID());

    const redirectUrl = `${baseUrl}${installPath}?state=${stateParam}&redirect_uri=${encodeURIComponent(callbackUrl)}`;

    window.location.href = redirectUrl;
  };

  return (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Github size={24} />
          {t('selfHost.githubInstaller.title')}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="mb-4 text-sm text-muted-foreground">
          {t('selfHost.githubInstaller.description')}
        </p>
        <Button className="w-full" onClick={handleConnectGithub}>
          {t('selfHost.githubInstaller.connectButton')}
        </Button>
      </CardContent>
      <CardFooter className="text-xs text-muted-foreground/40">
        <p>{t('selfHost.githubInstaller.terms')}</p>
      </CardFooter>
    </Card>
  );
};

export default GithubInstaller;
