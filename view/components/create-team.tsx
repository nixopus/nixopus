import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface CreateTeamProps {
  open: boolean;
  setOpen: (open: boolean) => void;
  createTeam: () => void;
  teamName: string;
  teamDescription: string;
  handleTeamNameChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  handleTeamDescriptionChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  isLoading?: boolean;
}

export function CreateTeam({
  open,
  setOpen,
  createTeam,
  teamName,
  teamDescription,
  handleTeamNameChange,
  handleTeamDescriptionChange,
  isLoading
}: CreateTeamProps) {
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create Team</DialogTitle>
          <DialogDescription>Create a new team to collaborate with others.</DialogDescription>
        </DialogHeader>
        <div className="flex-col items-center space-x-2 space-y-4 justify-center">
          <div className="grid flex-1 gap-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" defaultValue={teamName} onChange={handleTeamNameChange} />
          </div>
          <div className="grid flex-1 gap-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              defaultValue={teamDescription}
              onChange={handleTeamDescriptionChange}
            />
          </div>
          <Button
            type="submit"
            disabled={!teamName || !teamDescription || isLoading}
            className="px-10 flex self-end justify-self-center"
            onClick={createTeam}
          >
            <span>{isLoading ? 'Creating...' : 'Create'}</span>
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
