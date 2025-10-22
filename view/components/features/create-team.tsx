import { Button } from '@/components/ui/button';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';

interface CreateTeamProps {
  open: boolean;
  setOpen: (open: boolean) => void;
  createTeam: () => void;
  teamName: string;
  teamDescription: string;
  handleTeamNameChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handleTeamDescriptionChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  isLoading?: boolean;
}

export function CreateTeam({
  open,
  setOpen,
  createTeam,
  teamName,
  teamDescription,
  handleTeamNameChange,
  handleTeamDescriptionChange,
  isLoading
}: CreateTeamProps) {
  const { t } = useTranslation();

  const actions: DialogAction[] = [
    {
      label: isLoading
        ? t('settings.teams.createTeam.buttons.creating')
        : t('settings.teams.createTeam.buttons.create'),
      onClick: createTeam,
      disabled: !teamName || !teamDescription || isLoading,
      loading: isLoading,
      variant: 'default'
    }
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={setOpen}
      title={t('settings.teams.createTeam.title')}
      description={t('settings.teams.createTeam.description')}
      actions={actions}
      size="md"
    >
      <div className="flex-col items-center space-x-2 space-y-4 justify-center">
        <div className="grid flex-1 gap-2">
          <Label htmlFor="name">{t('settings.teams.createTeam.fields.name.label')}</Label>
          <Input
            id="name"
            defaultValue={teamName}
            onChange={handleTeamNameChange}
            placeholder={t('settings.teams.createTeam.fields.name.placeholder')}
          />
        </div>
        <div className="grid flex-1 gap-2">
          <Label htmlFor="description">
            {t('settings.teams.createTeam.fields.description.label')}
          </Label>
          <Input
            id="description"
            defaultValue={teamDescription}
            onChange={handleTeamDescriptionChange}
            placeholder={t('settings.teams.createTeam.fields.description.placeholder')}
          />
        </div>
      </div>
    </DialogWrapper>
  );
}
