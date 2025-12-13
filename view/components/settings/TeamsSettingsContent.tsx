'use client';

import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import useTeamSettings from '@/app/settings/hooks/use-team-settings';
import AddMember from '@/app/settings/teams/components/AddMember';
import TeamMembers from '@/app/settings/teams/components/TeamMembers';
import EditTeam from '@/app/settings/teams/components/EditTeam';
import TeamStats from '@/app/settings/teams/components/TeamStats';
import RecentActivity from '@/app/settings/teams/components/RecentActivity';
import { TypographyH1, TypographyMuted } from '@/components/ui/typography';

export function TeamsSettingsContent() {
  const { t } = useTranslation();
  const settings = useTeamSettings();

  return (
    <ResourceGuard resource="organization" action="read">
      <div className="space-y-6">
        <h2 className="text-2xl font-semibold">Teams</h2>
        <div className="flex items-center justify-between">
          <div>
            <TypographyH1>{settings.teamName}</TypographyH1>
            <TypographyMuted>{settings.teamDescription}</TypographyMuted>
          </div>
          <div className="flex gap-2">
            <ResourceGuard resource="organization" action="update">
              <EditTeam
                teamName={settings.teamName || ''}
                teamDescription={settings.teamDescription || ''}
                setEditTeamDialogOpen={settings.setEditTeamDialogOpen}
                handleUpdateTeam={settings.handleUpdateTeam}
                setTeamName={settings.setTeamName}
                setTeamDescription={settings.setTeamDescription}
                isEditTeamDialogOpen={settings.isEditTeamDialogOpen}
                isUpdating={settings.isUpdating}
              />
            </ResourceGuard>
            <ResourceGuard resource="user" action="create">
              <AddMember
                isAddUserDialogOpen={settings.isAddUserDialogOpen}
                setIsAddUserDialogOpen={settings.setIsAddUserDialogOpen}
                newUser={settings.newUser}
                setNewUser={settings.setNewUser}
                handleSendInvite={settings.handleSendInvite}
                isInviteLoading={settings.isInviteLoading}
              />
            </ResourceGuard>
          </div>
        </div>
        {settings.users.length > 0 ? (
          <TeamMembers
            users={settings.users}
            handleRemoveUser={settings.handleRemoveUser}
            getRoleBadgeVariant={settings.getRoleBadgeVariant}
            onUpdateUser={settings.handleUpdateUser}
          />
        ) : (
          <div className="text-center text-muted-foreground">{t('settings.teams.noMembers')}</div>
        )}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <TeamStats users={settings.users} />
          <RecentActivity />
        </div>
      </div>
    </ResourceGuard>
  );
}
