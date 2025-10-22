import React from 'react';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';
import { UserTypes } from '@/redux/types/orgs';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

interface EditUserDialogProps {
  isOpen: boolean;
  onClose: () => void;
  user: {
    id: string;
    name: string;
    email: string;
    avatar: string;
    role: 'Owner' | 'Admin' | 'Member' | 'Viewer';
  };
  onSave: (userId: string, role: UserTypes) => void;
}

const AVAILABLE_ROLES: SelectOption[] = [
  { value: 'admin', label: 'Admin' },
  { value: 'member', label: 'Member' },
  { value: 'viewer', label: 'Viewer' }
];

function EditUserDialog({ isOpen, onClose, user, onSave }: EditUserDialogProps) {
  const { t } = useTranslation();
  const [selectedRole, setSelectedRole] = React.useState<UserTypes>('admin');

  React.useEffect(() => {
    if (isOpen) {
      const role = user.role.toLowerCase() as UserTypes;
      setSelectedRole(role);
    }
  }, [isOpen, user]);

  const handleRoleChange = (value: string) => {
    const newRole = value as UserTypes;
    setSelectedRole(newRole);
  };

  const handleSave = () => {
    if (!selectedRole) {
      toast.error(t('settings.teams.editUser.dialog.errors.noRole'));
      return;
    }
    onSave(user.id, selectedRole);
    onClose();
  };

  const actions: DialogAction[] = [
    {
      label: t('settings.teams.editUser.dialog.buttons.cancel'),
      onClick: onClose,
      variant: 'outline'
    },
    {
      label: t('settings.teams.editUser.dialog.buttons.save'),
      onClick: handleSave,
      variant: 'default'
    }
  ];

  return (
    <DialogWrapper
      open={isOpen}
      onOpenChange={onClose}
      title={t('settings.teams.editUser.dialog.title')}
      description={t('settings.teams.editUser.dialog.description').replace('{name}', user.name)}
      actions={actions}
      size="sm"
    >
      <div className="space-y-6 py-4">
        <div className="space-y-2">
          <Label>{t('settings.teams.editUser.dialog.fields.role.label')}</Label>
          <SelectWrapper
            value={selectedRole}
            onValueChange={handleRoleChange}
            options={AVAILABLE_ROLES}
            placeholder={t('settings.teams.editUser.dialog.fields.role.placeholder')}
          />
        </div>
      </div>
    </DialogWrapper>
  );
}

export default EditUserDialog;
