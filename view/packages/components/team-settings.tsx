'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { Label } from '@/components/ui/label';
import { PlusIcon, PencilIcon, ArrowRight, Loader2 } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { UserTypes } from '@/redux/types/orgs';
import { DataTable } from '@/components/ui/data-table';
import { TrashIcon } from 'lucide-react';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { useRouter } from 'next/navigation';
import { useGetRecentAuditLogsQuery } from '@/redux/services/audit';
import { formatDistanceToNow } from 'date-fns';
import {
  AddMemberProps,
  AVAILABLE_ROLEs,
  EditTeamProps,
  EditUserDialogProps,
  TeamMembersProps,
  TeamStatsProps
} from '../types/settings';
import { useTeamMembers, useTeamStats } from '../hooks/settings/use-team-members';
import { useEditUserDialog } from '../hooks/settings/use-edit-user-dialog';

export function AddMember({
  isAddUserDialogOpen,
  setIsAddUserDialogOpen,
  newUser,
  setNewUser,
  handleSendInvite,
  isInviteLoading = false
}: AddMemberProps) {
  const { t } = useTranslation();

  const actions: DialogAction[] = [
    {
      label: 'Cancel',
      onClick: () => setIsAddUserDialogOpen(false),
      variant: 'outline'
    },
    {
      label: isInviteLoading ? 'Sending...' : 'Send Invite',
      onClick: handleSendInvite,
      disabled: isInviteLoading,
      loading: isInviteLoading,
      variant: 'default'
    }
  ];

  const trigger = (
    <Button size="sm">
      <PlusIcon className="h-4 w-4 mr-2" />
      Invite Member
    </Button>
  );

  return (
    <DialogWrapper
      open={isAddUserDialogOpen}
      onOpenChange={setIsAddUserDialogOpen}
      title="Invite Team Member"
      description="Send a magic link invitation to add a new member to your team"
      trigger={trigger}
      actions={actions}
      size="md"
    >
      <div className="grid gap-4 py-4">
        <div className="grid grid-cols-4 items-center gap-4">
          <Label htmlFor="email" className="text-right">
            Email
          </Label>
          <Input
            id="email"
            type="email"
            value={newUser.email}
            onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
            className="col-span-3"
            placeholder="Enter email address"
          />
        </div>
        <div className="grid grid-cols-4 items-center gap-4">
          <Label htmlFor="role" className="text-right">
            Role
          </Label>
          <SelectWrapper
            value={newUser.role}
            onValueChange={(value) => setNewUser({ ...newUser, role: value })}
            options={[
              { value: 'admin', label: 'Admin' },
              { value: 'member', label: 'Member' },
              { value: 'viewer', label: 'Viewer' }
            ]}
            placeholder="Select role"
            className="col-span-3"
          />
        </div>
      </div>
    </DialogWrapper>
  );
}

export function EditTeam({
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

export function EditUserDialog({ isOpen, onClose, user, onSave }: EditUserDialogProps) {
  const { t } = useTranslation();
  const { selectedRole, handleRoleChange, actions } = useEditUserDialog({
    isOpen,
    onClose,
    user,
    onSave
  });

  return (
    <DialogWrapper
      open={isOpen}
      onOpenChange={onClose}
      title={t('settings.teams.editUser.dialog.title')}
      description={t('settings.teams.editUser.dialog.description').replace('{name}', user.name)}
      actions={actions}
      size="sm"
    >
      <div className="space-y-6 py-4 ">
        <Label>{t('settings.teams.editUser.dialog.fields.role.label')}</Label>
        <SelectWrapper
          value={selectedRole}
          onValueChange={handleRoleChange}
          options={AVAILABLE_ROLEs}
          placeholder={t('settings.teams.editUser.dialog.fields.role.placeholder')}
        />
      </div>
    </DialogWrapper>
  );
}

export const getActionColor = (actionColor: string) => {
  switch (actionColor) {
    case 'green':
      return 'bg-green-500';
    case 'blue':
      return 'bg-blue-500';
    case 'red':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
};

export function RecentActivity() {
  const { t } = useTranslation();
  const router = useRouter();
  const { data: activities, isLoading, error } = useGetRecentAuditLogsQuery();

  const handleViewAll = () => {
    router.push('/activities');
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <TypographySmall className="text-sm font-medium">
            {t('settings.teams.recentActivity.title')}
          </TypographySmall>
          <TypographyMuted className="text-xs mt-1">
            {t('settings.teams.recentActivity.description')}
          </TypographyMuted>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleViewAll}
          className="flex items-center gap-1"
        >
          {t('settings.teams.recentActivity.viewAll')}
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
      {isLoading ? (
        <div className="flex items-center justify-center p-4">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : error ? (
        <div className="p-4 text-red-600">{t('settings.teams.recentActivity.error')}</div>
      ) : activities && activities.length > 0 ? (
        <div className="space-y-4">
          {activities.map((activity) => (
            <div key={activity.id} className="flex items-start gap-4">
              <div
                className={`h-2 w-2 mt-2 rounded-full ${getActionColor(activity.action_color)}`}
              />
              <div>
                <TypographySmall>{activity.message}</TypographySmall>
                <TypographyMuted>
                  {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
                </TypographyMuted>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center text-muted-foreground">
          {t('settings.teams.recentActivity.noActivities')}
        </div>
      )}
    </div>
  );
}

export function TeamMembers({
  users,
  handleRemoveUser,
  getRoleBadgeVariant,
  onUpdateUser
}: TeamMembersProps) {
  const { t } = useTranslation();
  const {
    columns,
    editingUser,
    userToRemove,
    isDeleteDialogOpen,
    handleDeleteConfirm,
    handleEditDialogClose,
    handleDeleteDialogOpenChange
  } = useTeamMembers({
    users,
    handleRemoveUser,
    getRoleBadgeVariant,
    onUpdateUser
  });

  const handleSaveUser = (userId: string, role: UserTypes) => {
    onUpdateUser(userId, role);
    handleEditDialogClose();
  };

  return (
    <>
      <div className="space-y-4">
        <div>
          <TypographySmall className="text-sm font-medium">
            {t('settings.teams.members.title')}
          </TypographySmall>
          <TypographyMuted className="text-xs mt-1">
            {t('settings.teams.members.description')}
          </TypographyMuted>
        </div>
        <DataTable data={users} columns={columns} showBorder={false} hoverable={false} />
      </div>

      {editingUser && (
        <EditUserDialog
          isOpen={!!editingUser}
          onClose={handleEditDialogClose}
          user={editingUser}
          onSave={handleSaveUser}
        />
      )}

      {userToRemove && (
        <DeleteDialog
          title={t('settings.teams.members.deleteDialog.title').replace(
            '{name}',
            userToRemove.name
          )}
          description={t('settings.teams.members.deleteDialog.description').replace(
            '{name}',
            userToRemove.name
          )}
          onConfirm={handleDeleteConfirm}
          confirmText={t('settings.teams.members.deleteDialog.confirm')}
          variant="destructive"
          icon={TrashIcon}
          open={isDeleteDialogOpen}
          onOpenChange={handleDeleteDialogOpenChange}
        />
      )}
    </>
  );
}

export function TeamStats({ users }: TeamStatsProps) {
  const { t } = useTranslation();
  const stats = useTeamStats(users);
  return (
    <div className="space-y-4">
      <div>
        <TypographySmall className="text-sm font-medium">
          {t('settings.teams.stats.title')}
        </TypographySmall>
        <TypographyMuted className="text-xs mt-1">
          {t('settings.teams.stats.description')}
        </TypographyMuted>
      </div>
      <div className="space-y-2">
        {stats.map((stat, index) => (
          <div key={index} className="flex justify-between items-center">
            <TypographyMuted className="text-sm">{stat.label}</TypographyMuted>
            <span className="font-medium">{stat.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

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
