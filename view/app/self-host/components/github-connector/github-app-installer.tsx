import React from 'react';
import { Button } from '@/components/ui/button';
import { Github } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

interface GithubInstallProps {
  appSlug: string;
  organization?: string;
  callbackUrl: string;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

const GithubInstaller = ({ appSlug, organization, callbackUrl, onSuccess, onError }: GithubInstallProps) => {
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
    <div className="w-full max-w-md space-y-6">
      <div className="text-center space-y-2">
        <div className="flex items-center justify-center gap-2">
          <Github size={24} />
          <h3 className="text-lg font-semibold">{t('selfHost.githubInstaller.title')}</h3>
        </div>
        <p className="text-sm text-muted-foreground">
          {t('selfHost.githubInstaller.description')}
        </p>
      </div>
      <Button className="w-full" onClick={handleConnectGithub} size="lg">
        {t('selfHost.githubInstaller.connectButton')}
      </Button>
      <p className="text-xs text-center text-muted-foreground/60">
        {t('selfHost.githubInstaller.terms')}
      </p>
    </div>
  );
};

export default GithubInstaller;
