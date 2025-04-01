import { useAppSelector } from '@/redux/hooks';
import {
  useCreateUserMutation,
  useGetOrganizationUsersQuery,
  useRemoveUserFromOrganizationMutation,
  useUpdateOrganizationDetailsMutation,
  useUpdateUserRoleMutation
} from '@/redux/services/users/userApi';
import { UserTypes } from '@/redux/types/orgs';
import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

function useTeamSettings() {
  const { t } = useTranslation();
  const [users, setUsers] = useState<any>([]);
  const [isAddUserDialogOpen, setIsAddUserDialogOpen] = useState(false);
  const [newUser, setNewUser] = useState({ name: '', email: '', role: 'Member', password: '' });
  const [isEditTeamDialogOpen, setEditTeamDialogOpen] = useState(false);
  const [teamName, setTeamName] = useState('');
  const [teamDescription, setTeamDescription] = useState('');
  const [createUser, { isLoading: isCreating }] = useCreateUserMutation();
  const [removeUserFromOrganization] = useRemoveUserFromOrganizationMutation();
  const [updateUserRole] = useUpdateUserRoleMutation();
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
        const roleName = user.role?.name || 'Unknown';
        const permissions =
          user.role?.permissions?.map(
            (permission) => `${permission.resource.toUpperCase()}:${permission.name}`
          ) || [];
        return {
          id: user.user.id,
          name: user.user?.username || 'Unknown User',
          email: user.user?.email || '',
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

  const handleAddUser = async () => {
    const newId = crypto.randomUUID();
    const tempUser = {
      username: newUser.name || '',
      email: newUser.email || '',
      password: newUser.password || '',
      organization: activeOrganization?.id || '',
      type: newUser.role.toLowerCase() as UserTypes
    };
    const permissions = newUser.role === 'Member' ? ['READ', 'UPDATE'] : ['READ'];
    setUsers([...users, { id: newId, ...tempUser, name: newUser.name, permissions }]);
    try {
      const user = await createUser(tempUser as any);
      await refetchUsers();
      toast.success(t('settings.teams.messages.userAdded'));
    } catch (error) {
      toast.error(t('settings.teams.messages.userAddFailed'));
    }
    setNewUser({ name: '', email: '', role: 'Member', password: '' });
    setIsAddUserDialogOpen(false);
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
        role_name: role
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
    handleAddUser,
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
    isUpdating
  };
}

export default useTeamSettings;
