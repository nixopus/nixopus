import React, { useState, useEffect, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Github, Loader2 } from 'lucide-react';
import {
  GitHubAppCredentials,
  GitHubAppManifest,
  GitHubAppProps,
  GitHubAppStatus
} from '@/redux/types/github';
import { useCreateGithubConnectorMutation } from '@/redux/services/connector/githubConnectorApi';
import { useTranslation } from '@/hooks/use-translation';
import { getWebhookUrl } from '@/redux/conf';

const adjectives = [
  'cosmic',
  'quantum',
  'stellar',
  'neural',
  'cyber',
  'atomic',
  'digital',
  'nebula',
  'phoenix',
  'shadow'
] as const;

const nouns = [
  'nexus',
  'pulse',
  'matrix',
  'cipher',
  'orbit',
  'nova',
  'core',
  'forge',
  'prism',
  'vector'
] as const;

const generateRandomName = (): string => {
  const adjective = adjectives[Math.floor(Math.random() * adjectives.length)];
  const noun = nouns[Math.floor(Math.random() * nouns.length)];
  return `${adjective}-${noun}`;
};

const GitHubAppManifestComponent: React.FC<GitHubAppProps> = ({
  organization,
  appUrl = process.env.NEXT_PUBLIC_APP_URL,
  redirectUrl = process.env.NEXT_PUBLIC_REDIRECT_URL,
  onSuccess,
  onError,
  onCreateClick
}) => {
  const { t } = useTranslation();
  const appName = useMemo(() => generateRandomName(), []);
  const [status, setStatus] = useState<GitHubAppStatus>('initial');
  const [error, setError] = useState<string | null>(null);
  const [createGithubConnector, { isLoading, error: registerGithubAppError }] =
    useCreateGithubConnectorMutation();
  const [webhookUrl, setWebhookUrl] = useState<string | null>(null);

  useEffect(() => {
    const fetchWebHookUrl = async () => {
      const url = await getWebhookUrl();
      setWebhookUrl(url);
    };
    fetchWebHookUrl();
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const stateParam = params.get('state');
    if (code) {
      handleGitHubCallback(code, stateParam);
    }
  }, []);

  useEffect(() => {
    if (onCreateClick && status === 'initial') {
      onCreateClick(createManifestForm);
    }
  }, [onCreateClick, status]);

  const generateState = (): string => {
    return crypto
      .getRandomValues(new Uint8Array(16))
      .reduce((acc, val) => acc + val.toString(16).padStart(2, '0'), '');
  };

  const createManifestForm = (): void => {
    const state = generateState();
    const manifest: GitHubAppManifest = {
      name: appName,
      url: appUrl || window.location.origin,
      hook_attributes: {
        url: webhookUrl || `${window.location.origin}/github/webhook`,
        active: true
      },
      redirect_url: redirectUrl || `${window.location.origin}/self-host`,
      callback_urls: [redirectUrl || `${window.location.origin}/self-host`],
      public: true,
      default_permissions: {
        contents: 'read',
        issues: 'write',
        metadata: 'read',
        pull_requests: 'write'
      },
      default_events: ['issues', 'issue_comment', 'pull_request', 'push'],
      setup_url: `${window.location.origin}/self-host`,
      setup_on_update: true
    };

    const form = document.createElement('form');
    form.method = 'post';
    form.action = organization
      ? `https://github.com/organizations/${organization}/settings/apps/new?state=${state}`
      : `https://github.com/settings/apps/new?state=${state}`;

    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = 'manifest';
    input.value = JSON.stringify(manifest);
    form.appendChild(input);

    document.body.appendChild(form);
    form.submit();
    document.body.removeChild(form);

    setStatus('redirecting');
  };

  const handleGitHubCallback = async (code: string, stateParam: string | null): Promise<void> => {
    setStatus('converting');
    try {
      const response = await fetch(`https://api.github.com/app-manifests/${code}/conversions`, {
        method: 'POST',
        headers: {
          Accept: 'application/vnd.github.v3+json'
        }
      });

      if (!response.ok) {
        console.log('response', response);
        throw new Error('Failed to convert manifest');
      }

      const credentials: GitHubAppCredentials = await response.json();

      await createGithubConnector({
        app_id: credentials.id.toString(),
        slug: credentials.slug,
        pem: credentials.pem,
        client_id: credentials.client_id,
        client_secret: credentials.client_secret,
        webhook_secret: credentials.webhook_secret
      });

      setStatus('success');
      onSuccess?.(credentials);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred';
      setError(errorMessage);
      setStatus('error');
      onError?.(err instanceof Error ? err : new Error(errorMessage));
    }
  };

  if (status === 'redirecting' || status === 'converting') {
    return (
      <div className="flex flex-col items-center gap-4 py-8">
        <Loader2 className="h-8 w-8 animate-spin" />
        <p>
          {status === 'redirecting'
            ? t('selfHost.githubManifest.status.redirecting')
            : t('selfHost.githubManifest.status.converting')}
        </p>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <Alert>
        <AlertDescription className="text-green-600">
          {t('selfHost.githubManifest.status.success')}
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="flex flex-col items-center gap-4 w-full">
      {error && (
        <Alert variant="destructive" className="w-full">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
    </div>
  );
};

export default GitHubAppManifestComponent;
