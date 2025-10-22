import React, { useState } from 'react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { TrashIcon, ChevronDownIcon, ChevronUpIcon, PencilIcon } from 'lucide-react';
import { DotsVerticalIcon } from '@radix-ui/react-icons';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useAppSelector } from '@/redux/hooks';
import EditUserDialog from './EditUserDialog';
import { UserTypes } from '@/redux/types/orgs';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { useTranslation } from '@/hooks/use-translation';
import { User } from '@/redux/types/user';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { useRBAC } from '@/lib/rbac';

type EditUser = {
  id: string;
  name: string;
  email: string;
  avatar: string;
  role: 'Owner' | 'Admin' | 'Member' | 'Viewer';
  permissions: string[];
};

type RoleType = 'owner' | 'admin' | 'member' | 'viewer';

interface TeamMembersProps {
  users: EditUser[];
  handleRemoveUser: (userId: string) => void;
  getRoleBadgeVariant: (role: string) => 'default' | 'secondary' | 'destructive' | 'outline';
  onUpdateUser: (userId: string, role: UserTypes) => Promise<void>;
}

const MAX_VISIBLE_PERMISSIONS = 3;

function TeamMembers({
  users,
  handleRemoveUser,
  getRoleBadgeVariant,
  onUpdateUser
}: TeamMembersProps) {
  const { t } = useTranslation();
  const { canAccessResource, isAdmin, isLoading: rbacLoading } = useRBAC();
  const loggedInUser = useAppSelector((state) => state.auth.user) as User;
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const [expandedUsers, setExpandedUsers] = useState<Set<string>>(new Set());
  const [editingUser, setEditingUser] = useState<EditUser | null>(null);
  const [userToRemove, setUserToRemove] = useState<EditUser | null>(null);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  const canModifyUser = (targetUser: EditUser) => {
    if (!loggedInUser || !targetUser || !activeOrganization) {
      return false;
    }

    if (loggedInUser.id === targetUser.id) {
      return false;
    }

    if (rbacLoading) {
      return false;
    }

    return isAdmin;
  };

  const toggleUserPermissions = (userId: string) => {
    setExpandedUsers((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(userId)) {
        newSet.delete(userId);
      } else {
        newSet.add(userId);
      }
      return newSet;
    });
  };

  const handleEditUser = (user: EditUser) => {
    if (!canModifyUser(user)) {
      return;
    }
    setEditingUser({
      ...user,
      permissions: user.permissions
    });
  };

  const handleSaveUser = (userId: string, role: UserTypes) => {
    const user = users.find((u) => u.id === userId);
    if (!user || !canModifyUser(user)) {
      return;
    }
    onUpdateUser(userId, role);
    setEditingUser(null);
  };

  const renderPermissions = (permissions: string[], userId: string) => {
    const isExpanded = expandedUsers.has(userId);
    const visiblePermissions = isExpanded
      ? permissions
      : permissions.slice(0, MAX_VISIBLE_PERMISSIONS);
    const hasMore = permissions.length > MAX_VISIBLE_PERMISSIONS;

    return (
      <div className="flex items-center gap-2">
        <div className="flex flex-wrap gap-1.5">
          {visiblePermissions.map((permission, index) => (
            <Badge
              key={index}
              variant="outline"
              className="bg-primary/10 text-primary rounded-full"
            >
              {permission}
            </Badge>
          ))}
        </div>
        {hasMore && (
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2 text-xs font-medium text-primary hover:text-primary/80"
            onClick={() => toggleUserPermissions(userId)}
          >
            {isExpanded ? (
              <>
                {t('settings.teams.members.actions.showLess')}{' '}
                <ChevronUpIcon className="ml-1 h-3 w-3" />
              </>
            ) : (
              <>
                {t('settings.teams.members.actions.showMore')}{' '}
                <ChevronDownIcon className="ml-1 h-3 w-3" />
              </>
            )}
          </Button>
        )}
      </div>
    );
  };

  const hasEditableActions = (user: EditUser) => {
    if (loggedInUser.id === user.id) {
      return false;
    }
    return canModifyUser(user);
  };

  const hasAnyEditableActions = users.some((user) => hasEditableActions(user));

  const columns: TableColumn<EditUser>[] = [
    {
      key: 'user',
      title: t('settings.teams.members.table.headers.user'),
      render: (_, user) => (
        <div className="flex items-center space-x-3">
          <Avatar className="w-12 h-12 shadow-md">
            {user.avatar ? (
              <AvatarImage src={user.avatar} alt="Profile avatar" />
            ) : (
              <AvatarFallback className="bg-secondary text-foreground text-xl font-medium">
                {user.name.slice(0, 2).toUpperCase()}
              </AvatarFallback>
            )}
          </Avatar>
          <div>
            <div className="font-medium">{user.name}</div>
            <TypographyMuted>{user.email}</TypographyMuted>
          </div>
        </div>
      )
    },
    {
      key: 'role',
      title: t('settings.teams.members.table.headers.role'),
      render: (_, user) => (
        <Badge variant={getRoleBadgeVariant(user.role)}>{user.role}</Badge>
      )
    },
    {
      key: 'permissions',
      title: t('settings.teams.members.table.headers.permissions'),
      render: (_, user) => renderPermissions(user.permissions, user.id)
    }
  ];

  if (hasAnyEditableActions) {
    columns.push({
      key: 'actions',
      title: t('settings.teams.members.table.headers.actions'),
      render: (_, user) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <DotsVerticalIcon className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <ResourceGuard resource="user" action="update">
              {canModifyUser(user) && (
                <DropdownMenuItem onClick={() => handleEditUser(user)}>
                  <PencilIcon className="h-4 w-4 mr-2" />
                  {t('settings.teams.members.actions.edit')}
                </DropdownMenuItem>
              )}
            </ResourceGuard>
            <ResourceGuard resource="user" action="update">
              <ResourceGuard resource="user" action="delete">
                {canModifyUser(user) && <DropdownMenuSeparator />}
              </ResourceGuard>
            </ResourceGuard>
            <ResourceGuard resource="user" action="delete">
              {canModifyUser(user) && (
                <DropdownMenuItem
                  className="text-destructive focus:text-destructive"
                  onClick={() => {
                    setUserToRemove(user);
                    setIsDeleteDialogOpen(true);
                  }}
                >
                  <TrashIcon className="h-4 w-4 mr-2" />
                  {t('settings.teams.members.actions.remove')}
                </DropdownMenuItem>
              )}
            </ResourceGuard>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
      align: 'right'
    });
  }

  return (
    <>
      <Card>
        <CardHeader>
          <TypographySmall>{t('settings.teams.members.title')}</TypographySmall>
          <TypographyMuted>{t('settings.teams.members.description')}</TypographyMuted>
        </CardHeader>
        <CardContent>
          <DataTable
            data={users}
            columns={columns}
            showBorder={false}
            hoverable={false}
          />
        </CardContent>
      </Card>

      {editingUser && (
        <EditUserDialog
          isOpen={!!editingUser}
          onClose={() => setEditingUser(null)}
          user={editingUser}
          onSave={handleSaveUser}
        />
      )}

      {userToRemove && (
        <DeleteDialog
          title={t('settings.teams.members.deleteDialog.title').replace(
            '{name}',
            userToRemove.name
          )}
          description={t('settings.teams.members.deleteDialog.description').replace(
            '{name}',
            userToRemove.name
          )}
          onConfirm={() => {
            handleRemoveUser(userToRemove.id);
            setUserToRemove(null);
            setIsDeleteDialogOpen(false);
          }}
          confirmText={t('settings.teams.members.deleteDialog.confirm')}
          variant="destructive"
          icon={TrashIcon}
          open={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
        />
      )}
    </>
  );
}

export default TeamMembers;
