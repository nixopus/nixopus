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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { PlusIcon } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

interface AddMemberProps {
  isAddUserDialogOpen: boolean;
  setIsAddUserDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  newUser: {
    name: string;
    email: string;
    role: string;
    password: string;
  };
  setNewUser: React.Dispatch<
    React.SetStateAction<{
      name: string;
      email: string;
      role: string;
      password: string;
    }>
  >;
  handleAddUser: () => void;
}

function AddMember({
  isAddUserDialogOpen,
  setIsAddUserDialogOpen,
  newUser,
  setNewUser,
  handleAddUser
}: AddMemberProps) {
  const { t } = useTranslation();

  return (
    <Dialog open={isAddUserDialogOpen} onOpenChange={setIsAddUserDialogOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <PlusIcon className="h-4 w-4 mr-2" />
          {t('settings.teams.addMember.button')}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('settings.teams.addMember.dialog.title')}</DialogTitle>
          <DialogDescription>{t('settings.teams.addMember.dialog.description')}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              {t('settings.teams.addMember.dialog.fields.name.label')}
            </Label>
            <Input
              id="name"
              value={newUser.name}
              onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
              className="col-span-3"
              placeholder={t('settings.teams.addMember.dialog.fields.name.placeholder')}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="email" className="text-right">
              {t('settings.teams.addMember.dialog.fields.email.label')}
            </Label>
            <Input
              id="email"
              type="email"
              value={newUser.email}
              onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
              className="col-span-3"
              placeholder={t('settings.teams.addMember.dialog.fields.email.placeholder')}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="role" className="text-right">
              {t('settings.teams.addMember.dialog.fields.role.label')}
            </Label>
            <Select
              value={newUser.role}
              onValueChange={(value) => setNewUser({ ...newUser, role: value })}
            >
              <SelectTrigger className="col-span-3">
                <SelectValue
                  placeholder={t('settings.teams.addMember.dialog.fields.role.placeholder')}
                />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Admin">
                  {t('settings.teams.addMember.dialog.fields.role.options.admin')}
                </SelectItem>
                <SelectItem value="Member">
                  {t('settings.teams.addMember.dialog.fields.role.options.member')}
                </SelectItem>
                <SelectItem value="Viewer">
                  {t('settings.teams.addMember.dialog.fields.role.options.viewer')}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="password" className="text-right">
              {t('settings.teams.addMember.dialog.fields.password.label')}
            </Label>
            <Input
              id="password"
              type="password"
              value={newUser.password}
              onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
              className="col-span-3"
              placeholder={t('settings.teams.addMember.dialog.fields.password.placeholder')}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setIsAddUserDialogOpen(false)}>
            {t('settings.teams.addMember.dialog.buttons.cancel')}
          </Button>
          <Button onClick={handleAddUser}>
            {t('settings.teams.addMember.dialog.buttons.add')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default AddMember;
