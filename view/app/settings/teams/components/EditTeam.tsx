'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { Label } from '@/components/ui/label';
import { PencilIcon } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

interface EditTeamProps {
  isEditTeamDialogOpen: boolean;
  setEditTeamDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  handleUpdateTeam: () => void;
  teamName: string;
  setTeamName: React.Dispatch<React.SetStateAction<string>>;
  teamDescription: string;
  setTeamDescription: React.Dispatch<React.SetStateAction<string>>;
  isUpdating: boolean;
}

function EditTeam({
  isEditTeamDialogOpen,
  setEditTeamDialogOpen,
  handleUpdateTeam,
  teamName,
  setTeamName,
  teamDescription,
  setTeamDescription,
  isUpdating
}: EditTeamProps) {
  const { t } = useTranslation();

  const actions: DialogAction[] = [
    {
      label: t('settings.teams.editTeam.dialog.buttons.cancel'),
      onClick: () => setEditTeamDialogOpen(false),
      variant: 'outline'
    },
    {
      label: isUpdating
        ? t('settings.teams.editTeam.dialog.buttons.updating')
        : t('settings.teams.editTeam.dialog.buttons.update'),
      onClick: handleUpdateTeam,
      disabled: isUpdating,
      loading: isUpdating,
      variant: 'default'
    }
  ];

  const trigger = (
    <Button variant={'ghost'} size={'icon'} className="ml-12">
      <PencilIcon className="w-4 h-4" />
    </Button>
  );

  return (
    <DialogWrapper
      open={isEditTeamDialogOpen}
      onOpenChange={setEditTeamDialogOpen}
      title={t('settings.teams.editTeam.dialog.title')}
      description={t('settings.teams.editTeam.dialog.description')}
      trigger={trigger}
      actions={actions}
      size="md"
    >
      <div className="grid gap-4 py-4">
        <div className="grid grid-cols-4 items-center gap-4">
          <Label htmlFor="name" className="text-right">
            {t('settings.teams.editTeam.dialog.fields.name.label')}
          </Label>
          <Input
            id="name"
            value={teamName}
            onChange={(e) => setTeamName(e.target.value)}
            className="col-span-3"
            placeholder={t('settings.teams.editTeam.dialog.fields.name.placeholder')}
          />
        </div>
        <div className="grid grid-cols-4 items-center gap-4">
          <Label htmlFor="description" className="text-right">
            {t('settings.teams.editTeam.dialog.fields.description.label')}
          </Label>
          <Input
            id="description"
            value={teamDescription}
            onChange={(e) => setTeamDescription(e.target.value)}
            className="col-span-3"
            placeholder={t('settings.teams.editTeam.dialog.fields.description.placeholder')}
          />
        </div>
      </div>
    </DialogWrapper>
  );
}

export default EditTeam;
