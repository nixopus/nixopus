import React from 'react';
import { Github, CheckCircle2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

interface StepperNavigationProps {
  Stepper: any;
}

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

export const StepperNavigation: React.FC<StepperNavigationProps> = ({ Stepper }) => {
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

