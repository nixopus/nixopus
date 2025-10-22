import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import {
  useCreateOrganizationMutation,
  useDeleteOrganizationMutation
} from '@/redux/services/users/userApi';
import React from 'react';
import { toast } from 'sonner';
import { setActiveOrganization } from '@/redux/features/users/userSlice';
import { domainsApi } from '@/redux/services/settings/domainsApi';
import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { deployApi } from '@/redux/services/deploy/applicationsApi';
import { notificationApi } from '@/redux/services/settings/notificationApi';
import { UserOrganization } from '@/redux/types/orgs';

const ACTIVE_ORGANIZATION_KEY = 'active_organization';

function useTeamSwitcher() {
  const [open, setOpen] = React.useState(false);
  const user = useAppSelector((state) => state.auth.user);
  const isAdmin = React.useMemo(() => user?.type === 'admin', [user]);
  const [teamName, setTeamName] = React.useState('');
  const [teamDescription, setTeamDescription] = React.useState('');
  const [createOrganization, { isLoading }] = useCreateOrganizationMutation();
  const [deleteOrganization] = useDeleteOrganizationMutation();
  const dispatch = useAppDispatch();
  const activeTeam = useAppSelector((state) => state.user.activeOrganization);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const teams = useAppSelector((state) => state.user.organizations);
  const displayTeam = activeTeam || (teams && teams.length > 0 ? teams[0].organization : null);

  React.useEffect(() => {
    const storedOrg = localStorage.getItem(ACTIVE_ORGANIZATION_KEY);
    if (storedOrg && teams) {
      const parsedOrg = JSON.parse(storedOrg);
      const matchingTeam = teams.find(
        (team: UserOrganization) => team.organization.id === parsedOrg.id
      );
      if (matchingTeam) {
        dispatch(setActiveOrganization(matchingTeam.organization));
      }
    }
  }, [teams, dispatch]);

  const handleTeamChange = async (team: UserOrganization) => {
    dispatch(setActiveOrganization(team.organization));
    localStorage.setItem(ACTIVE_ORGANIZATION_KEY, JSON.stringify(team.organization));
    try {
      dispatch(domainsApi.util.invalidateTags([{ type: 'Domains', id: 'LIST' }]));
      dispatch(GithubConnectorApi.util.invalidateTags([{ type: 'GithubConnector', id: 'LIST' }]));
      dispatch(deployApi.util.invalidateTags([{ type: 'Deploy', id: 'LIST' }]));
      dispatch(notificationApi.util.invalidateTags([{ type: 'Notification', id: 'LIST' }]));
    } catch (error) {
      console.error('Failed to invalidate cache:', error);
    }
  };

  const handleDeleteOrganization = async () => {
    if (teams && teams.length <= 1) {
      return;
    }

    try {
      await deleteOrganization(displayTeam.id).unwrap();
      const remainingTeams = teams?.filter(
        (team: UserOrganization) => team.organization.id !== displayTeam.id
      );
      if (remainingTeams && remainingTeams.length > 0) {
        handleTeamChange(remainingTeams[0]);
      }
      setIsDeleteDialogOpen(false);
    } catch (error) {
      console.error('Failed to delete organization:', error);
    }
  };

  const toggleAddTeamModal = () => {
    setOpen(!open);
  };

  const handleTeamNameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setTeamName(event.target.value);
  };

  const handleTeamDescriptionChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setTeamDescription(event.target.value);
  };

  const validateTeamName = (name: string) => {
    if (!name) {
      return false;
    }
    return name.length <= 50;
  };

  const validateTeamDescription = (description: string) => {
    if (!description) {
      return false;
    }
    return description.length <= 100;
  };

  const onCreateTeam = async () => {
    try {
      if (!isAdmin) {
        toast.error('You are not an admin');
        return;
      }

      if (!validateTeamName(teamName)) {
        toast.error('Team name is required and must be less than 50 characters');
        return;
      }

      if (!validateTeamDescription(teamDescription)) {
        toast.error('Team description is required and must be less than 100 characters');
        return;
      }

      const res = await createOrganization({
        name: teamName,
        description: teamDescription
      }).unwrap();

      if (!res.id) {
        toast.error('Failed to create team');
        return;
      }
      toast.success('Team created successfully');
      setOpen(false);
      setTeamName('');
      setTeamDescription('');

      // Refresh the page to ensure all roles and permissions are updated
      // This is necessary because SuperTokens roles are session based
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    } catch (error) {
      console.error('Failed to create team:', error);
      toast.error('Failed to create team');
    }
  };

  return {
    addTeamModalOpen: open,
    setAddTeamModalOpen: setOpen,
    toggleAddTeamModal,
    createTeam: onCreateTeam,
    teamName,
    teamDescription,
    handleTeamNameChange,
    handleTeamDescriptionChange,
    isLoading,
    handleTeamChange,
    handleDeleteOrganization,
    isDeleteDialogOpen,
    setIsDeleteDialogOpen,
    activeTeam,
    isAdmin,
    displayTeam
  };
}

export default useTeamSwitcher;
