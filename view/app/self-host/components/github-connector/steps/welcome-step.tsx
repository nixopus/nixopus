import React from 'react';
import { Github, CheckCircle2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import GitHubAppManifestComponent from '../github-manifest-flow';
import { GitHubAppCredentials } from '@/redux/types/github';

interface WelcomeStepProps {
  error: string | null;
  organization?: string;
  onSuccess: (creds: GitHubAppCredentials) => void;
  onError: (error: Error) => void;
  onCreateClick?: (createFn: () => void) => void;
}

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
        <h3 className="text-2xl font-semibold">
          {t('selfHost.githubSetup.welcome.title' as any)}
        </h3>
        <p className="text-muted-foreground">
          {t('selfHost.githubSetup.welcome.description' as any)}
        </p>
      </div>
    </div>
  );
};

// const ErrorAlert: React.FC<{ error: string }> = ({ error }) => (
//   <Alert variant="destructive">
//     <AlertCircle className="h-4 w-4" />
//     <AlertDescription>{error}</AlertDescription>
//   </Alert>
// );

export const WelcomeStep: React.FC<WelcomeStepProps> = ({
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

