'use client';
import React from 'react';
import RecentActivity from './components/RecentActivity';
import TeamStats from './components/TeamStats';
import useTeamSettings from '../hooks/use-team-settings';
import AddMember from './components/AddMember';
import TeamMembers from './components/TeamMembers';
import EditTeam from './components/EditTeam';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { TypographyH1, TypographyMuted } from '@/components/ui/typography';
import PageLayout from '@/components/layout/page-layout';

function Page() {
  const { t } = useTranslation();
  const {
    users,
    isAddUserDialogOpen,
    setIsAddUserDialogOpen,
    newUser,
    setNewUser,
    handleSendInvite,
    handleRemoveUser,
    getRoleBadgeVariant,
    handleUpdateTeam,
    setEditTeamDialogOpen,
    setTeamName,
    setTeamDescription,
    isEditTeamDialogOpen,
    teamName,
    teamDescription,
    isUpdating,
    handleUpdateUser,
    isInviteLoading
  } = useTeamSettings();

  return (
    <ResourceGuard resource="organization" action="read">
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        <div className={'flex items-center justify-between space-y-2'}>
          <div className="flex items-center">
            <span className="">
              <TypographyH1>{teamName}</TypographyH1>
              <TypographyMuted>{teamDescription}</TypographyMuted>
            </span>

            <ResourceGuard resource="organization" action="update">
              <EditTeam
                teamName={teamName || ''}
                teamDescription={teamDescription || ''}
                setEditTeamDialogOpen={setEditTeamDialogOpen}
                handleUpdateTeam={handleUpdateTeam}
                setTeamName={setTeamName}
                setTeamDescription={setTeamDescription}
                isEditTeamDialogOpen={isEditTeamDialogOpen}
                isUpdating={isUpdating}
              />
            </ResourceGuard>
          </div>

          <ResourceGuard resource="user" action="create">
            <AddMember
              isAddUserDialogOpen={isAddUserDialogOpen}
              setIsAddUserDialogOpen={setIsAddUserDialogOpen}
              newUser={newUser}
              setNewUser={setNewUser}
              handleSendInvite={handleSendInvite}
              isInviteLoading={isInviteLoading}
            />
          </ResourceGuard>
        </div>

        {users.length > 0 ? (
          <TeamMembers
            users={users}
            handleRemoveUser={handleRemoveUser}
            getRoleBadgeVariant={getRoleBadgeVariant}
            onUpdateUser={handleUpdateUser}
          />
        ) : (
          <div className="text-center text-muted-foreground">{t('settings.teams.noMembers')}</div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <TeamStats users={users} />
          <RecentActivity />
        </div>
      </PageLayout>
    </ResourceGuard>
  );
}

export default Page;
