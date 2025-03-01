import { useAppSelector } from '@/redux/hooks';
import {
  useGetOrganizationUsersQuery,
  useUpdateOrganizationDetailsMutation
} from '@/redux/services/users/userApi';
import { useState, useEffect } from 'react';
import { toast } from 'sonner';

function useTeamSettings() {
  const [users, setUsers] = useState<any>([]);
  const [isAddUserDialogOpen, setIsAddUserDialogOpen] = useState(false);
  const [newUser, setNewUser] = useState({ name: '', email: '', role: 'Member' });
  const [isEditTeamDialogOpen, setEditTeamDialogOpen] = useState(false);
  const [teamName, setTeamName] = useState('');
  const [teamDescription, setTeamDescription] = useState('');

  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const {
    data: apiUsers,
    isLoading,
    error
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
            (permission) => permission.resource.toUpperCase() + ':' + permission.name
          ) || [];
        return {
          id: user.user.id,
          name: user.user?.username || 'Unknown User',
          email: user.user?.email || '',
          role: roleName,
          permissions,
          avatar: user.user?.avatar
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

  const handleAddUser = () => {
    const newId = crypto.randomUUID();
    let permissions: string[] = [];

    switch (newUser.role) {
      case 'Owner':
      case 'Admin':
        permissions = ['READ', 'UPDATE', 'DELETE', 'MANAGE'];
        break;
      case 'Member':
        permissions = ['READ', 'UPDATE'];
        break;
      case 'Viewer':
        permissions = ['READ'];
        break;
    }

    const tempUser = {
      id: newId,
      name: newUser.name,
      email: newUser.email,
      role: newUser.role,
      permissions,
      avatar: ''
    };

    setUsers([...users, tempUser]);
    setNewUser({ name: '', email: '', role: 'Member' });
    setIsAddUserDialogOpen(false);
  };

  const handleRemoveUser = (userId: string) => {
    setUsers(users.filter((user: any) => user.id !== userId));
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
      toast.error('Team name and description are required');
      setTeamName(activeOrganization?.name || '');
      setTeamDescription(activeOrganization?.description || '');
      return;
    }

    if (
      teamName !== activeOrganization?.name ||
      teamDescription !== activeOrganization?.description
    ) {
      await updateOrganizationDetails({
        id: activeOrganization?.id,
        name: teamName,
        description: teamDescription
      });
      toast.success('Team details updated successfully');
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
