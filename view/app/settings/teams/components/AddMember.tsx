'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { DialogWrapper, DialogAction } from '@/components/ui/dialog-wrapper';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';
import { Label } from '@/components/ui/label';
import { PlusIcon } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

interface AddMemberProps {
  isAddUserDialogOpen: boolean;
  setIsAddUserDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  newUser: {
    email: string;
    role: string;
  };
  setNewUser: React.Dispatch<
    React.SetStateAction<{
      email: string;
      role: string;
    }>
  >;
  handleSendInvite: () => void;
  isInviteLoading?: boolean;
}

function AddMember({
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

export default AddMember;
