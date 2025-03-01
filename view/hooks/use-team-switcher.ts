import { useAppSelector } from '@/redux/hooks'
import { useAddUserToOrganizationMutation, useCreateOrganizationMutation } from '@/redux/services/users/userApi'
import { UserOrganization } from '@/redux/types/orgs'
import React from 'react'
import { toast } from 'sonner'

function useTeamSwitcher() {
    const [open, setOpen] = React.useState(false)
    const user = useAppSelector(state => state.auth.user)
    const isAdmin = React.useMemo(() => user?.type === "admin", [user])
    const [teamName, setTeamName] = React.useState("")
    const [teamDescription, setTeamDescription] = React.useState("")
    const [createOrganization, { isLoading }] = useCreateOrganizationMutation()
    const [addUserToOrganization, { isLoading: isAddingUser }] = useAddUserToOrganizationMutation()
    const organizations = useAppSelector(state => state.user.organizations)

    const toggleAddTeamModal = () => {
        setOpen(!open)
    }

    const handleTeamNameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setTeamName(event.target.value)
    }

    const handleTeamDescriptionChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setTeamDescription(event.target.value)
    }

    const validateTeamName = (name: string) => {
        return name.length > 0
    }

    const validateTeamDescription = (description: string) => {
        return description.length > 0
    }

    const getOwnerRoleId = () => {
        const org = organizations.find((org: UserOrganization) => org.role.name === "Owner")
        return org?.role.id
    }

    const onCreateTeam = async () => {
        try {
            if (!isAdmin) {
                toast.error("You are not an admin")
                return
            }

            if (!validateTeamName(teamName)) {
                toast.error("Team name is required")
                return
            }

            if (!validateTeamDescription(teamDescription)) {
                toast.error("Team description is required")
                return
            }

            const res = await createOrganization({
                name: teamName,
                description: teamDescription
            }).unwrap()

            if (!res.id) {
                toast.error("Failed to create team")
                return
            }
            const ownerRoleId = getOwnerRoleId()
            await addUserToOrganization({
                organization_id: res.id,
                user_id: user?.id,
                role_id: ownerRoleId
            }).unwrap()
            toast.success("Team created successfully")
            setOpen(false)
        } catch (error) {
            console.log(error)
            toast.error("Failed to create team")
        }
    }

    return {
        addTeamModalOpen: open,
        setAddTeamModalOpen: setOpen,
        toggleAddTeamModal,
        createTeam: onCreateTeam,
        teamName,
        teamDescription,
        handleTeamNameChange,
        handleTeamDescriptionChange,
        isLoading: isLoading || isAddingUser
    }
}

export default useTeamSwitcher