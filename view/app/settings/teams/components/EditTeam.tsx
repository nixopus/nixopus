'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { PencilIcon } from 'lucide-react';

interface EditTeamProps {
  isEditTeamDialogOpen: boolean;
  setEditTeamDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  handleUpdateTeam: () => void;
  teamName: string;
  setTeamName: React.Dispatch<React.SetStateAction<string>>;
  teamDescription: string;
  setTeamDescription: React.Dispatch<React.SetStateAction<string>>;
  isUpdating: boolean;
}

function EditTeam({
  isEditTeamDialogOpen,
  setEditTeamDialogOpen,
  handleUpdateTeam,
  teamName,
  setTeamName,
  teamDescription,
  setTeamDescription,
  isUpdating
}: EditTeamProps) {
  return (
    <Dialog open={isEditTeamDialogOpen} onOpenChange={setEditTeamDialogOpen}>
      <DialogTrigger asChild>
        <Button variant={'ghost'} size={'icon'} className="ml-12">
          <PencilIcon className="w-4 h-4" />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Team Details</DialogTitle>
          <DialogDescription>Team details helps you identify your team.</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              Name
            </Label>
            <Input
              id="name"
              value={teamName}
              onChange={(e) => setTeamName(e.target.value)}
              className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="description" className="text-right">
              Description
            </Label>
            <Input
              id="name"
              value={teamDescription}
              onChange={(e) => setTeamDescription(e.target.value)}
              className="col-span-3"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setEditTeamDialogOpen(false)}>
            Cancel
          </Button>
          <Button onClick={handleUpdateTeam} disabled={isUpdating}>
            {isUpdating ? 'Updating...' : 'Update'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default EditTeam;
