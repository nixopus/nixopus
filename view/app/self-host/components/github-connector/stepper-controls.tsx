import React from 'react';
import { Button } from '@/components/ui/button';
import { ArrowRight, ArrowLeft, Github } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

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

export const StepperControls: React.FC<StepperControlsProps> = ({
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

