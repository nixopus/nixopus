'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Terminal, Github, GitBranch, Copy, Check, ArrowRight, LogOut } from 'lucide-react';
import { Separator } from '@/components/ui/separator';
import { TypographyH1, TypographyMuted } from '@/components/ui/typography';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useCreateAPIKeyMutation } from '@/redux/services/users/authApi';
import { toast } from 'sonner';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import { logout, logoutUser } from '@/redux/features/users/authSlice';
import { LogoutDialog } from '@/components/ui/logout-dialog';
import {
  useCreateGithubConnectorMutation,
  useUpdateGithubConnectorMutation,
  useGetAllGithubConnectorQuery
} from '@/redux/services/connector/githubConnectorApi';
import { setActiveConnectorId } from '@/redux/features/github-connector/githubConnectorSlice';

export default function OnboardPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const dispatch = useAppDispatch();
  const { t } = useTranslation();
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [showLogoutDialog, setShowLogoutDialog] = useState(false);
  const [createAPIKey, { isLoading: isGenerating }] = useCreateAPIKeyMutation();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const [createGithubConnector] = useCreateGithubConnectorMutation();
  const [updateGithubConnector] = useUpdateGithubConnectorMutation();
  const { data: connectors, refetch: refetchConnectors } = useGetAllGithubConnectorQuery();

  // Handle GitHub installation callback if installation_id is present
  useEffect(() => {
    const handleGitHubCallback = async () => {
      const installationId = searchParams.get('installation_id');
      const setupAction = searchParams.get('setup_action');

      if (installationId && setupAction === 'install') {
        try {
          // Check if connector exists
          let connectorId: string | null = null;

          if (connectors && connectors.length > 0) {
            // Find connector without installation_id or use first one
            const connectorWithoutInstallation = connectors.find(
              (c) => !c.installation_id || c.installation_id.trim() === ''
            );
            connectorId = connectorWithoutInstallation?.id || connectors[0].id;
          } else {
            // Create connector with empty credentials (will use shared config from backend)
            await createGithubConnector({
              app_id: '',
              slug: '',
              pem: '',
              client_id: '',
              client_secret: '',
              webhook_secret: ''
            }).unwrap();

            // Refetch to get the new connector
            const updatedConnectors = await refetchConnectors().unwrap();
            if (updatedConnectors && updatedConnectors.length > 0) {
              connectorId = updatedConnectors[0].id;
            }
          }

          if (connectorId) {
            // Update connector with installation_id (stored in github_connectors table)
            await updateGithubConnector({
              installation_id: installationId,
              connector_id: connectorId
            }).unwrap();

            dispatch(setActiveConnectorId(connectorId));
            await refetchConnectors();

            // Clean up URL and redirect to dashboard
            window.history.replaceState({}, document.title, '/onboard');
            router.push('/dashboard');
          }
        } catch (err: any) {
          console.error('Failed to handle GitHub callback:', err);
          toast.error(err?.data?.message || 'Failed to connect GitHub');
        }
      }
    };

    handleGitHubCallback();
  }, [
    searchParams,
    connectors,
    createGithubConnector,
    updateGithubConnector,
    refetchConnectors,
    dispatch,
    router
  ]);

  const handleGenerateApiKey = async () => {
    if (apiKey) {
      return;
    }

    if (!activeOrganization?.id) {
      toast.error(t('deploy.cli.noOrganization') || 'Please select an organization first');
      return;
    }

    try {
      const result = await createAPIKey({
        name: 'Deployment API Key'
      }).unwrap();
      setApiKey(result.key);
      toast.success(t('deploy.cli.apiKeyGenerated') || 'API key generated successfully!');
    } catch (error: any) {
      toast.error(
        error?.data?.message || t('deploy.cli.apiKeyError') || 'Failed to generate API key'
      );
    }
  };

  const handleCopyAPIKey = async () => {
    if (!apiKey) return;
    try {
      await navigator.clipboard.writeText(apiKey);
      setCopied(true);
      toast.success(t('deploy.cli.apiKeyCopied') || 'API key copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      toast.error(t('deploy.cli.copyError') || 'Failed to copy API key');
    }
  };

  const handleGitDeploy = () => {
    // Direct redirect to GitHub App installation
    const githubAppSlug = process.env.NEXT_PUBLIC_GITHUB_APP_SLUG || 'nixopus-deploy';
    const callbackUrl = `${window.location.origin}/dashboard/github-callback`;
    const installUrl = `https://github.com/apps/${githubAppSlug}/installations/new?redirect_uri=${encodeURIComponent(
      callbackUrl
    )}`;
    window.location.href = installUrl;
  };

  const handleLogoutClick = () => {
    setShowLogoutDialog(true);
  };

  const handleLogoutConfirm = async () => {
    setShowLogoutDialog(false);

    try {
      const keys = [
        'COLLAPSIBLE_STATE_KEY',
        'LAST_ACTIVE_NAV_KEY',
        'SIDEBAR_STORAGE_KEY',
        'terminalOpen',
        'persist:root',
        'active_organization'
      ];
      keys.forEach((key) => localStorage.removeItem(key));

      dispatch({ type: 'RESET_STATE' });
      await dispatch(logoutUser() as any);
      router.push('/auth');
    } catch (error) {
      console.error('Logout failed:', error);
      const keys = [
        'COLLAPSIBLE_STATE_KEY',
        'LAST_ACTIVE_NAV_KEY',
        'SIDEBAR_STORAGE_KEY',
        'terminalOpen',
        'persist:root',
        'active_organization'
      ];
      keys.forEach((key) => localStorage.removeItem(key));

      dispatch({ type: 'RESET_STATE' });
      dispatch(logout());
      router.push('/auth');
    }
  };

  const handleLogoutCancel = () => {
    setShowLogoutDialog(false);
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-background p-6 relative">
      <Button
        variant="ghost"
        size="sm"
        onClick={handleLogoutClick}
        className="absolute top-4 right-4 gap-2"
      >
        <LogOut className="h-4 w-4" />
        {t('user.menu.logout') || 'Logout'}
      </Button>
      <div className="flex flex-col items-center space-y-4 w-full max-w-2xl mx-auto">
        <div className="text-center">
          <TypographyH1 className="text-4xl font-bold tracking-tight">
            {t('deploy.page.title') || 'Choose Your Deployment Method'}
          </TypographyH1>
        </div>

        <Card className="w-full">
          <CardContent className="p-6 space-y-4">
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-muted">
                  <Terminal className="h-5 w-5 text-muted-foreground" />
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold text-base">
                    {t('deploy.cli.title') || 'CLI / API Deployment'}
                  </h3>
                  <TypographyMuted className="text-xs">
                    {t('deploy.cli.description') ||
                      'Deploy from your terminal using an API key.'}
                  </TypographyMuted>
                </div>
              </div>

              {!apiKey ? (
                <Button
                  className="w-full"
                  size="lg"
                  variant="outline"
                  onClick={handleGenerateApiKey}
                  disabled={isGenerating}
                >
                  {isGenerating
                    ? t('deploy.cli.generating') || 'Generating...'
                    : t('deploy.cli.button') || 'Generate API Key'}
                </Button>
              ) : (
                <div className="space-y-3">
                  <div className="space-y-2">
                    <div className="flex gap-2">
                      <Input value={apiKey} readOnly className="font-mono text-sm flex-1" />
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={handleCopyAPIKey}
                        className="flex-shrink-0"
                      >
                        {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                    <TypographyMuted className="text-xs text-destructive/80">
                      {t('deploy.cli.apiKeyWarning') ||
                        "Save this API key now. You won't be able to see it again."}
                    </TypographyMuted>
                  </div>
                </div>
              )}
            </div>

            <div className="relative flex items-center py-2">
              <Separator className="flex-1" />
              <span className="px-4 text-sm text-muted-foreground">{t('deploy.or') || 'Or'}</span>
              <Separator className="flex-1" />
            </div>

            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-primary/10">
                  <Github className="h-5 w-5 text-primary" />
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold text-base">
                    {t('deploy.git.title') || 'Connect Git Repository'}
                  </h3>
                  <TypographyMuted className="text-xs">
                    {t('deploy.git.description') ||
                      'Automatic deployments on every push. Best for teams and continuous delivery.'}
                  </TypographyMuted>
                </div>
              </div>
              <Button className="w-full" size="lg" onClick={handleGitDeploy}>
                <GitBranch className="h-4 w-4 mr-2" />
                {t('deploy.git.button') || 'Connect Repository'}
                <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
      <LogoutDialog
        open={showLogoutDialog}
        onConfirm={handleLogoutConfirm}
        onCancel={handleLogoutCancel}
      />
    </div>
  );
}
