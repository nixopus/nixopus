import { useAppSelector } from '@/redux/hooks';
import {
  useGetOrganizationUsersQuery,
  useRemoveUserFromOrganizationMutation,
  useUpdateOrganizationDetailsMutation,
  useUpdateUserRoleMutation,
  useSendInviteMutation
} from '@/redux/services/users/userApi';
import { UserTypes } from '@/redux/types/orgs';
import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

function useTeamSettings() {
  const { t } = useTranslation();
  const [users, setUsers] = useState<any>([]);
  const [isAddUserDialogOpen, setIsAddUserDialogOpen] = useState(false);
  const [newUser, setNewUser] = useState({ email: '', role: 'member' });
  const [isEditTeamDialogOpen, setEditTeamDialogOpen] = useState(false);
  const [teamName, setTeamName] = useState('');
  const [teamDescription, setTeamDescription] = useState('');
  const [removeUserFromOrganization] = useRemoveUserFromOrganizationMutation();
  const [updateUserRole] = useUpdateUserRoleMutation();
  const [sendInvite, { isLoading: isInviteLoading }] = useSendInviteMutation();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const {
    data: apiUsers,
    isLoading,
    error,
    refetch: refetchUsers
  } = useGetOrganizationUsersQuery(activeOrganization?.id, {
    skip: !activeOrganization
  });
  const [updateOrganizationDetails, { isLoading: isUpdating, error: updateError }] =
    useUpdateOrganizationDetailsMutation();

  useEffect(() => {
    if (apiUsers) {
      const transformedUsers = apiUsers.map((user) => {
        const primaryRole = user.roles?.[0] || 'Unknown';
        const roleName = primaryRole.includes('admin')
          ? 'Admin'
          : primaryRole.includes('member')
            ? 'Member'
            : primaryRole.includes('viewer')
              ? 'Viewer'
              : primaryRole.includes('owner')
                ? 'Owner'
                : 'Unknown';

        const permissions = user.permissions || [];

        return {
          id: user.user.id,
          name: user.user?.username || 'Unknown User',
          email: user.user?.email || '',
          avatar: user.user?.avatar || '',
          role: roleName,
          permissions
        };
      });

      setUsers(transformedUsers);
    }
  }, [apiUsers]);

  useEffect(() => {
    if (activeOrganization) {
      setTeamName(activeOrganization.name);
      setTeamDescription(activeOrganization.description);
    }
  }, [activeOrganization]);

  const handleSendInvite = async () => {
    if (!newUser.email || !newUser.role || !activeOrganization?.id) {
      toast.error('Please fill in all required fields');
      return;
    }

    try {
      await sendInvite({
        email: newUser.email,
        organization_id: activeOrganization.id,
        role: newUser.role.toLowerCase()
      });
      toast.success(`Invitation sent to ${newUser.email}`);
      setNewUser({ email: '', role: 'member' });
      setIsAddUserDialogOpen(false);
    } catch (error) {
      toast.error('Failed to send invitation');
    }
  };

  const handleRemoveUser = async (userId: string) => {
    try {
      await removeUserFromOrganization({
        user_id: userId,
        organization_id: activeOrganization?.id || ''
      });
      await refetchUsers();
      toast.success(t('settings.teams.messages.userRemoved'));
    } catch (error) {
      toast.error(t('settings.teams.messages.userRemoveFailed'));
    }
  };

  const handleUpdateUser = async (userId: string, role: UserTypes) => {
    try {
      await updateUserRole({
        user_id: userId,
        organization_id: activeOrganization?.id || '',
        role: role.toLowerCase()
      });
      await refetchUsers();
      toast.success(t('settings.teams.messages.userUpdated'));
    } catch (error) {
      toast.error(t('settings.teams.messages.userUpdateFailed'));
    }
  };

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case 'Owner':
        return 'default';
      case 'Admin':
        return 'destructive';
      case 'Member':
        return 'default';
      case 'Viewer':
        return 'secondary';
      default:
        return 'outline';
    }
  };

  const handleUpdateTeam = async () => {
    setEditTeamDialogOpen(false);
    if (teamName.length <= 0 || teamDescription.length <= 0) {
      toast.error(t('settings.teams.messages.requiredFields'));
      setTeamName(activeOrganization?.name || '');
      setTeamDescription(activeOrganization?.description || '');
      return;
    }

    if (
      teamName !== activeOrganization?.name ||
      teamDescription !== activeOrganization?.description
    ) {
      try {
        await updateOrganizationDetails({
          id: activeOrganization?.id || '',
          name: teamName,
          description: teamDescription
        });
        await refetchUsers();
        toast.success(t('settings.teams.messages.teamUpdated'));
      } catch (error) {
        toast.error(t('settings.teams.messages.teamUpdateFailed'));
        setTeamName(activeOrganization?.name || '');
        setTeamDescription(activeOrganization?.description || '');
      }
    }
  };

  return {
    users,
    isLoading,
    error,
    isAddUserDialogOpen,
    setIsAddUserDialogOpen,
    newUser,
    setNewUser,
    handleSendInvite,
    handleRemoveUser,
    handleUpdateUser,
    getRoleBadgeVariant,
    handleUpdateTeam,
    setEditTeamDialogOpen,
    setTeamName,
    setTeamDescription,
    isEditTeamDialogOpen,
    teamName,
    teamDescription,
    isUpdating,
    isInviteLoading
  };
}

export default useTeamSettings;
