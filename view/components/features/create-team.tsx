import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
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

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('settings.teams.createTeam.title')}</DialogTitle>
          <DialogDescription>{t('settings.teams.createTeam.description')}</DialogDescription>
        </DialogHeader>
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
          <Button
            type="submit"
            disabled={!teamName || !teamDescription || isLoading}
            className="px-10 flex self-end justify-self-center"
            onClick={createTeam}
          >
            <span>
              {isLoading
                ? t('settings.teams.createTeam.buttons.creating')
                : t('settings.teams.createTeam.buttons.create')}
            </span>
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
