'use client';

import React, { useState, useEffect } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { UserTypes } from '@/redux/types/orgs';
import { toast } from 'sonner';
import { DialogAction } from '@/components/ui/dialog-wrapper';
import { EditUserDialogProps } from '../../types/settings';

export function useEditUserDialog({ isOpen, onClose, user, onSave }: EditUserDialogProps) {
  const { t } = useTranslation();
  const [selectedRole, setSelectedRole] = useState<UserTypes>('admin');

  useEffect(() => {
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

  return {
    selectedRole,
    handleRoleChange,
    actions
  };
}
