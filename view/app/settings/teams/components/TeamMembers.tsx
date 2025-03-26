import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { TrashIcon } from 'lucide-react';
import { DotsVerticalIcon } from '@radix-ui/react-icons';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useAppSelector } from '@/redux/hooks';
import { useResourcePermissions } from '@/lib/permission';

interface TeamMembersProps {
  users: {
    id: string;
    name: string;
    email: string;
    avatar: string;
    role: 'Owner' | 'Admin' | 'Member' | 'Viewer';
    permissions: string[];
  }[];
  handleRemoveUser: (userId: string) => void;
  getRoleBadgeVariant: (role: string) => 'destructive' | 'default' | 'secondary' | 'outline';
}

function TeamMembers({ users, handleRemoveUser, getRoleBadgeVariant }: TeamMembersProps) {
  const loggedInUser = useAppSelector((state) => state.auth.user);
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const {
    canUpdate: canUpdateUser,
    canDelete: canDeleteUser
  } = useResourcePermissions(loggedInUser, "user", activeOrganization?.id);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Team Members</CardTitle>
        <CardDescription>Manage users and their roles in your team.</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>User</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Permissions</TableHead>
              {/* {(canUpdateUser || canDeleteUser) && (
                <TableHead className="text-right">Actions</TableHead>
              )} */}
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.map((user) => (
              <TableRow key={user.id}>
                <TableCell className="font-medium">
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
                      <div className="text-sm text-muted-foreground">{user.email}</div>
                    </div>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant={getRoleBadgeVariant(user.role)}>{user.role}</Badge>
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {user.permissions.map((permission, index) => (
                      <Badge key={index} variant="outline">
                        {permission}
                      </Badge>
                    ))}
                  </div>
                </TableCell>
                {/* {(canUpdateUser || canDeleteUser) && loggedInUser.id !== user.id && (
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <DotsVerticalIcon className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                        {canUpdateUser && (
                          <>
                            <DropdownMenuItem>Edit User</DropdownMenuItem>
                            <DropdownMenuItem>Change Role</DropdownMenuItem>
                          </>
                        )}
                        {canUpdateUser && canDeleteUser && <DropdownMenuSeparator />}
                        {canDeleteUser && (
                          <DropdownMenuItem
                            className="text-destructive focus:text-destructive"
                            onClick={() => handleRemoveUser(user.id)}
                          >
                            <TrashIcon className="h-4 w-4 mr-2" />
                            Remove User
                          </DropdownMenuItem>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                )} */}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

export default TeamMembers;