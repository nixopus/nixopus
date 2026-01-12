'use client';

import React, { useRef } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import DialogWrapper, { DialogAction } from '@/components/ui/dialog-wrapper';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import {
  Github,
  Settings,
  Trash2,
  RefreshCw,
  Plus,
  Check,
  ExternalLink,
  AlertCircle,
  CheckCircle2,
  ArrowRight,
  ArrowLeft
} from 'lucide-react';
import { GitHubAppProps, GitHubAppCredentials, GithubConnector } from '@/redux/types/github';
import { useGitHubAppSetup } from '@/packages/hooks/applications/use-github-app-setup';
import { useGithubManifestFlow } from '@/packages/hooks/github/use_github_manifest_flow';
import useGithubConnectorSettings from '@/packages/hooks/applications/use-github-connector-settings';

const STEP_CONFIG = [
  {
    id: 'create-app',
    icon: Github,
    titleKey: 'selfHost.githubSetup.steps.createApp.title',
    descriptionKey: 'selfHost.githubSetup.steps.createApp.description'
  },
  {
    id: 'install-app',
    icon: CheckCircle2,
    titleKey: 'selfHost.githubSetup.steps.installApp.title',
    descriptionKey: 'selfHost.githubSetup.steps.installApp.description'
  }
] as const;

const BENEFITS = [
  {
    key: 'secure',
    icon: CheckCircle2
  },
  {
    key: 'automated',
    icon: CheckCircle2
  },
  {
    key: 'repositories',
    icon: CheckCircle2
  }
] as const;

const GitHubAppManifestComponent: React.FC<GitHubAppProps> = ({
  organization,
  appUrl = process.env.NEXT_PUBLIC_APP_URL,
  redirectUrl = process.env.NEXT_PUBLIC_REDIRECT_URL,
  onSuccess,
  onError,
  onCreateClick
}) => {
  const { status, loadingContent, successContent, errorContent } = useGithubManifestFlow({
    organization,
    appUrl,
    redirectUrl,
    onSuccess,
    onError,
    onCreateClick
  });

  if (status === 'redirecting' || status === 'converting') {
    return loadingContent;
  }

  if (status === 'success') {
    return successContent;
  }

  return <div className="flex flex-col items-center gap-4 w-full">{errorContent}</div>;
};

// ============================================================================
// GitHub App Installer Component
// ============================================================================

interface GithubInstallProps {
  appSlug: string;
  organization?: string;
  callbackUrl: string;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

const GithubInstaller: React.FC<GithubInstallProps> = ({
  appSlug,
  organization,
  callbackUrl,
  onSuccess,
  onError
}) => {
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
        <p className="text-sm text-muted-foreground">{t('selfHost.githubInstaller.description')}</p>
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

interface WelcomeStepProps {
  error: string | null;
  organization?: string;
  onSuccess: (creds: GitHubAppCredentials) => void;
  onError: (error: Error) => void;
  onCreateClick?: (createFn: () => void) => void;
}

const BenefitItem: React.FC<{ benefitKey: string; Icon: typeof CheckCircle2 }> = ({
  benefitKey,
  Icon
}) => {
  const { t } = useTranslation();

  return (
    <div className="flex items-start gap-3 p-4 rounded-lg bg-muted/30">
      <div className="rounded-full bg-primary/10 p-1.5 mt-0.5 shrink-0">
        <Icon size={16} className="text-primary" />
      </div>
      <div className="flex-1 space-y-1">
        <p className="text-sm font-medium">
          {t(`selfHost.githubSetup.welcome.benefits.${benefitKey}.title` as any)}
        </p>
        <p className="text-xs text-muted-foreground">
          {t(`selfHost.githubSetup.welcome.benefits.${benefitKey}.description` as any)}
        </p>
      </div>
    </div>
  );
};

const WelcomeHeader: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col items-center text-center space-y-4">
      <div className="rounded-full bg-primary/10 p-4">
        <Github size={48} className="text-primary" />
      </div>
      <div className="space-y-2">
        <h3 className="text-2xl font-semibold">{t('selfHost.githubSetup.welcome.title' as any)}</h3>
        <p className="text-muted-foreground">
          {t('selfHost.githubSetup.welcome.description' as any)}
        </p>
      </div>
    </div>
  );
};

const WelcomeStep: React.FC<WelcomeStepProps> = ({
  error,
  organization,
  onSuccess,
  onError,
  onCreateClick
}) => {
  return (
    <div className="pt-8 pb-6 px-6 space-y-6">
      <WelcomeHeader />
      <div className="space-y-3 pt-2">
        {BENEFITS.map((benefit) => (
          <BenefitItem key={benefit.key} benefitKey={benefit.key} Icon={benefit.icon} />
        ))}
      </div>
      <div className="flex justify-center pt-2">
        <GitHubAppManifestComponent
          organization={organization}
          onSuccess={onSuccess}
          onError={onError}
          onCreateClick={onCreateClick}
        />
      </div>
    </div>
  );
};

interface InstallAppStepProps {
  credentials: GitHubAppCredentials;
  organization?: string;
  onSuccess: () => void;
  onError: (error: Error) => void;
}

const InstallAppStep: React.FC<InstallAppStepProps> = ({
  credentials,
  organization,
  onSuccess,
  onError
}) => {
  return (
    <div className="pt-8 pb-6 px-6">
      <div className="flex justify-center">
        <GithubInstaller
          appSlug={credentials.slug}
          organization={organization}
          callbackUrl=""
          onSuccess={onSuccess}
          onError={onError}
        />
      </div>
    </div>
  );
};

interface StepperNavigationProps {
  Stepper: any;
}

const StepItem: React.FC<{
  Stepper: any;
  stepId: string;
  Icon: typeof Github;
  titleKey: string;
  descriptionKey: string;
}> = ({ Stepper, stepId, Icon, titleKey, descriptionKey }) => {
  const { t } = useTranslation();

  return (
    <Stepper.Step of={stepId} icon={<Icon size={20} />} className="flex-col gap-2">
      <Stepper.Title className="text-sm font-medium">{t(titleKey as any)}</Stepper.Title>
      <Stepper.Description className="text-xs text-muted-foreground">
        {t(descriptionKey as any)}
      </Stepper.Description>
    </Stepper.Step>
  );
};

const StepperNavigation: React.FC<StepperNavigationProps> = ({ Stepper }) => {
  return (
    <div className="flex justify-center w-full mb-8">
      <Stepper.Navigation aria-label="GitHub App Setup Steps" className="w-full max-w-2xl">
        {STEP_CONFIG.map((step) => (
          <StepItem
            key={step.id}
            Stepper={Stepper}
            stepId={step.id}
            Icon={step.icon}
            titleKey={step.titleKey}
            descriptionKey={step.descriptionKey}
          />
        ))}
      </Stepper.Navigation>
    </div>
  );
};

interface StepperControlsProps {
  Stepper: any;
  isFirstStep: boolean;
  isLastStep: boolean;
  canGoNext: boolean;
  onBack: () => void;
  onNext: () => void;
  onCreateApp?: () => void;
  currentStepId: string;
}

const StepperControls: React.FC<StepperControlsProps> = ({
  Stepper,
  isFirstStep,
  isLastStep,
  canGoNext,
  onBack,
  onNext,
  onCreateApp,
  currentStepId
}) => {
  const { t } = useTranslation();

  const isCreateAppStep = currentStepId === 'create-app';

  return (
    <Stepper.Controls className="px-6 pb-8 pt-6">
      <div className={`flex w-full ${isFirstStep ? 'justify-center' : 'justify-between'}`}>
        {!isFirstStep && (
          <Button variant="outline" onClick={onBack} className="gap-2">
            <ArrowLeft size={16} />
            {t('selfHost.githubSetup.buttons.back' as any)}
          </Button>
        )}

        {isCreateAppStep && onCreateApp ? (
          <Button onClick={onCreateApp} className="gap-2">
            <Github size={16} />
            {t('selfHost.githubManifest.createButton' as any)}
          </Button>
        ) : (
          !isLastStep && (
            <Button onClick={onNext} disabled={!canGoNext} className="gap-2">
              {t('selfHost.githubSetup.buttons.next' as any)}
              <ArrowRight size={16} />
            </Button>
          )
        )}
      </div>
    </Stepper.Controls>
  );
};

interface StepPanelProps {
  Stepper: any;
  currentStepId: string;
  credentials: GitHubAppCredentials | null;
  error: string | null;
  organization?: string;
  onRegistrationSuccess: (creds: GitHubAppCredentials) => void;
  onRegistrationError: (error: Error) => void;
  onInstallationSuccess: () => void;
  onInstallationError: (error: Error) => void;
  onCreateClick?: (createFn: () => void) => void;
}

const renderStepContent = ({
  currentStepId,
  credentials,
  error,
  organization,
  onRegistrationSuccess,
  onRegistrationError,
  onInstallationSuccess,
  onInstallationError,
  onCreateClick
}: Omit<StepPanelProps, 'Stepper'>): React.ReactNode => {
  if (currentStepId === 'create-app') {
    return (
      <WelcomeStep
        error={error}
        organization={organization}
        onSuccess={onRegistrationSuccess}
        onError={onRegistrationError}
        onCreateClick={onCreateClick}
      />
    );
  }

  if (currentStepId === 'install-app' && credentials) {
    return (
      <InstallAppStep
        credentials={credentials}
        organization={organization}
        onSuccess={onInstallationSuccess}
        onError={onInstallationError}
      />
    );
  }

  return null;
};

const StepPanel: React.FC<StepPanelProps> = (props) => {
  const { Stepper, ...rest } = props;

  return <Stepper.Panel>{renderStepContent(rest)}</Stepper.Panel>;
};

interface ConnectorItemProps {
  connector: GithubConnector;
  isActive: boolean;
  onSetActive: (id: string) => void;
  onDelete: (id: string) => void;
  onReset: (id: string) => void;
  isDeleting: boolean;
  isUpdating: boolean;
}

const ConnectorItem: React.FC<ConnectorItemProps> = ({
  connector,
  isActive,
  onSetActive,
  onDelete,
  isDeleting,
  isUpdating
}) => {
  const { t } = useTranslation();
  const [showDeleteDialog, setShowDeleteDialog] = React.useState(false);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowDeleteDialog(true);
  };

  const confirmDelete = () => {
    onDelete(connector.id);
    setShowDeleteDialog(false);
  };

  const handleCardClick = () => {
    if (!isActive && !isDeleting && !isUpdating) {
      onSetActive(connector.id);
    }
  };

  return (
    <>
      <div
        onClick={handleCardClick}
        className={`flex items-center justify-between p-4 rounded-lg border transition-all ${
          isActive
            ? 'border-primary bg-primary/5 cursor-default'
            : 'border-border bg-card cursor-pointer hover:border-primary/50 hover:bg-muted/50'
        } ${isDeleting || isUpdating ? 'opacity-50 cursor-not-allowed' : ''}`}
      >
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <div className="rounded-full bg-primary/10 p-2 shrink-0">
            <Github size={20} className="text-primary" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <p className="font-medium text-sm truncate">{connector.name || connector.slug}</p>
              {isActive && (
                <Badge variant="default" className="text-xs">
                  <Check size={12} className="mr-1" />
                  {t('selfHost.connectorSettings.active' as any)}
                </Badge>
              )}
            </div>
            <p className="text-xs text-muted-foreground truncate">
              {t('selfHost.connectorSettings.connector.slug' as any)}: {connector.slug || 'N/A'}
            </p>
            {connector.installation_id && (
              <p className="text-xs text-muted-foreground truncate">
                {t('selfHost.connectorSettings.connector.installationId' as any)}:{' '}
                {connector.installation_id.substring(0, 8)}...
              </p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0" onClick={(e) => e.stopPropagation()}>
          <Button
            variant="destructive"
            size="sm"
            onClick={handleDelete}
            disabled={isDeleting || isUpdating}
            title={t('selfHost.connectorSettings.actions.delete.label' as any)}
          >
            <Trash2 size={16} />
          </Button>
        </div>
      </div>

      <DeleteDialog
        open={showDeleteDialog}
        onOpenChange={setShowDeleteDialog}
        title={t('selfHost.connectorSettings.actions.delete.dialog.title' as any)}
        description={t(
          'selfHost.connectorSettings.actions.delete.dialog.description' as any
        ).replace('{name}', connector.name || connector.slug)}
        onConfirm={confirmDelete}
        confirmText={
          isDeleting
            ? t('selfHost.connectorSettings.actions.delete.dialog.deleting' as any)
            : t('selfHost.connectorSettings.actions.delete.dialog.confirm' as any)
        }
        cancelText={t('selfHost.connectorSettings.actions.delete.dialog.cancel' as any)}
        isDeleting={isDeleting}
        variant="destructive"
        icon={Trash2}
      />
    </>
  );
};

interface GitHubConnectorSettingsModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAddNew?: () => void;
}

const GitHubConnectorSettingsModal: React.FC<GitHubConnectorSettingsModalProps> = ({
  open,
  onOpenChange,
  onAddNew
}) => {
  const { t } = useTranslation();
  const {
    connectors,
    activeConnector,
    activeConnectorId,
    isLoadingConnectors,
    isDeleting,
    isUpdating,
    handleSetActiveConnector,
    handleDeleteConnector,
    handleResetConnector,
    setIsSettingsModalOpen
  } = useGithubConnectorSettings();

  const handleClose = () => {
    setIsSettingsModalOpen(false);
    onOpenChange(false);
  };

  const handleAddNew = () => {
    handleClose();
    onAddNew?.();
  };

  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: handleClose,
      variant: 'outline'
    },
    ...(onAddNew
      ? [
          {
            label: t('selfHost.connectorSettings.actions.addNew' as any),
            onClick: handleAddNew,
            variant: 'default' as const,
            icon: Plus
          }
        ]
      : [])
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={handleClose}
      title={t('selfHost.connectorSettings.title' as any)}
      description={t('selfHost.connectorSettings.description' as any)}
      actions={actions}
      size="lg"
    >
      <div className="space-y-4">
        {isLoadingConnectors ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-20 w-full" />
            ))}
          </div>
        ) : connectors.length === 0 ? (
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {t('selfHost.connectorSettings.noConnectors' as any)}
            </AlertDescription>
          </Alert>
        ) : (
          <>
            <div className="space-y-3">
              {connectors.map((connector) => (
                <ConnectorItem
                  key={connector.id}
                  connector={connector}
                  isActive={connector.id === activeConnectorId}
                  onSetActive={handleSetActiveConnector}
                  onDelete={handleDeleteConnector}
                  onReset={handleResetConnector}
                  isDeleting={isDeleting === connector.id}
                  isUpdating={isUpdating === connector.id}
                />
              ))}
            </div>
            {(() => {
              const currentActiveConnector = connectors.find((c) => c.id === activeConnectorId);

              if (!currentActiveConnector) {
                return null;
              }

              const installationUrl = `https://github.com/settings/installations/${currentActiveConnector.installation_id}`;

              return (
                <div className="pt-4 border-t">
                  <div className="flex items-start gap-3">
                    <Settings size={18} className="text-muted-foreground mt-0.5 shrink-0" />
                    <div className="flex-1 space-y-2">
                      <div>
                        <p className="text-sm font-medium mb-1">
                          {t('selfHost.connectorSettings.currentConnector.title' as any)}
                        </p>
                        <div className="flex items-center gap-2 mb-2">
                          <Badge variant="secondary" className="text-xs">
                            {currentActiveConnector.name || currentActiveConnector.slug}
                          </Badge>
                        </div>
                      </div>
                      <p className="text-xs text-muted-foreground leading-relaxed">
                        {t('selfHost.connectorSettings.currentConnector.description' as any)}
                      </p>
                      {currentActiveConnector.installation_id && (
                        <a
                          href={installationUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-1.5 text-xs text-primary hover:underline font-medium"
                        >
                          {t('selfHost.connectorSettings.currentConnector.viewOnGithub' as any)}
                          <ExternalLink size={12} />
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              );
            })()}
          </>
        )}
      </div>
    </DialogWrapper>
  );
};

interface GitHubAppSetupProps {
  organization?: string;
  GetGithubConnectors: () => void;
}

const GitHubAppSetup: React.FC<GitHubAppSetupProps> = ({ organization, GetGithubConnectors }) => {
  const handleGetGithubConnectors = async () => {
    await GetGithubConnectors();
  };

  const createAppRef = useRef<(() => void) | null>(null);

  const {
    Stepper,
    utils,
    credentials,
    error,
    setStepperMethods,
    handleRegistrationSuccess,
    handleRegistrationError,
    handleInstallationSuccess,
    handleInstallationError,
    handleNext,
    handleBack,
    getCanGoNext
  } = useGitHubAppSetup(handleGetGithubConnectors);

  const handleCreateClick = (createFn: () => void) => {
    createAppRef.current = createFn;
  };

  const handleCreateApp = () => {
    if (createAppRef.current) {
      createAppRef.current();
    }
  };

  return (
    <div className="flex flex-col items-center w-full max-w-4xl mx-auto p-4 space-y-6">
      <Stepper.Provider
        variant="horizontal"
        labelOrientation="vertical"
        initialStep="create-app"
        className="w-full"
      >
        {({ methods }) => {
          setStepperMethods(methods);
          const { current } = methods;
          const currentStepId = current.id;
          const isFirstStep = utils.getIndex(currentStepId) === 0;
          const isLastStep = utils.getLast().id === currentStepId;
          const canGoNext = Boolean(getCanGoNext(currentStepId));

          return (
            <>
              <StepperNavigation Stepper={Stepper} />
              <div className="flex justify-center w-full">
                <div className="w-full max-w-2xl">
                  <StepPanel
                    Stepper={Stepper}
                    currentStepId={currentStepId}
                    credentials={credentials}
                    error={error}
                    organization={organization}
                    onRegistrationSuccess={handleRegistrationSuccess}
                    onRegistrationError={handleRegistrationError}
                    onInstallationSuccess={handleInstallationSuccess}
                    onInstallationError={handleInstallationError}
                    onCreateClick={handleCreateClick}
                  />
                  <StepperControls
                    Stepper={Stepper}
                    isFirstStep={isFirstStep}
                    isLastStep={isLastStep}
                    canGoNext={canGoNext}
                    onBack={handleBack}
                    onNext={handleNext}
                    onCreateApp={handleCreateApp}
                    currentStepId={currentStepId}
                  />
                </div>
              </div>
            </>
          );
        }}
      </Stepper.Provider>
    </div>
  );
};

export default GitHubAppSetup;
export { GitHubConnectorSettingsModal, GitHubAppManifestComponent, GithubInstaller };
