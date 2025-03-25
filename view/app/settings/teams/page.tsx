'use client';
import React from 'react';
import RecentActivity from './components/RecentActivity';
import TeamStats from './components/TeamStats';
import useTeamSettings from '../hooks/use-team-settings';
import AddMember from './components/AddMember';
import TeamMembers from './components/TeamMembers';
import EditTeam from './components/EditTeam';
import { useResourcePermissions } from '@/lib/permission';
import { useAppSelector } from '@/redux/hooks';

function Page() {
  const {
    users,
    isAddUserDialogOpen,
    setIsAddUserDialogOpen,
    newUser,
    setNewUser,
    handleAddUser,
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
  } = useTeamSettings();

  const user = useAppSelector((state) => state.auth.user);
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);

  const { canUpdate: canUpdateUser, canDelete: canDeleteUser, canCreate: canCreateUser } =
    useResourcePermissions(user, "user", activeOrganization?.id);
  const { canRead: canReadOrg, canUpdate: canUpdateOrg } =
    useResourcePermissions(user, "organization", activeOrganization?.id);

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <div className={'flex items-center justify-between space-y-2'}>
        <div className="flex items-center">
          <span className="">
            <h2 className="text-2xl font-bold tracking-tight">{teamName}</h2>
            <p className="text-muted-foreground">{teamDescription}</p>
          </span>

          {canUpdateOrg && (
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
          )}
        </div>

        {canCreateUser && (
          <AddMember
            isAddUserDialogOpen={isAddUserDialogOpen}
            setIsAddUserDialogOpen={setIsAddUserDialogOpen}
            newUser={newUser}
            setNewUser={setNewUser}
            handleAddUser={handleAddUser}
          />
        )}
      </div>

      <TeamMembers
        users={users}
        handleRemoveUser={handleRemoveUser}
        getRoleBadgeVariant={getRoleBadgeVariant}
      />

      {canReadOrg && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <TeamStats users={users} />
          <RecentActivity />
        </div>
      )}
    </div>
  );
}

export default Page;