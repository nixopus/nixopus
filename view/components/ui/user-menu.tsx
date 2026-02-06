'use client';

import * as React from 'react';
import { useTheme } from 'next-themes';
import { Check, LogOut, Palette } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@nixopus/ui';
import { Avatar, AvatarFallback, AvatarImage } from '@nixopus/ui';
import { Button } from '@nixopus/ui';
import { palette, themeColors } from '@/packages/utils/colors';
import { User } from '@/redux/types/user';
import { cn } from '@/lib/utils';

interface UserMenuProps {
  user: User | null;
  onLogout: () => void;
}

export function UserMenu({ user, onLogout }: UserMenuProps) {
  const { setTheme, theme } = useTheme();

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  const displayName = user?.username || user?.email || 'User';
  const initials = getInitials(displayName);

  const getCardColor = (themeName: string) => {
    if (themeName && themeColors[themeName]?.card) {
      return `hsl(${themeColors[themeName].card})`;
    }
    return undefined;
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="relative h-9 w-9 rounded-full">
          <Avatar className="h-9 w-9">
            {user?.avatar && <AvatarImage src={user.avatar} alt={displayName} />}
            <AvatarFallback className="bg-primary/10 text-primary text-sm font-medium">
              {initials}
            </AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-64" align="end" sideOffset={8}>
        <DropdownMenuLabel className="font-normal">
          <div className="flex items-center gap-3 py-1">
            <Avatar className="h-10 w-10">
              {user?.avatar && <AvatarImage src={user.avatar} alt={displayName} />}
              <AvatarFallback className="bg-primary/10 text-primary text-sm font-medium">
                {initials}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col space-y-0.5 leading-none">
              <p className="text-sm font-medium">{displayName}</p>
              {user?.email && user.email !== displayName && (
                <p className="text-xs text-muted-foreground truncate max-w-[160px]">{user.email}</p>
              )}
            </div>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuLabel className="text-xs text-muted-foreground flex items-center gap-2 py-1.5">
          <Palette className="h-3.5 w-3.5" />
          Theme
        </DropdownMenuLabel>
        {palette.map((themeName) => {
          const isActive = theme === themeName;
          const cardColor = getCardColor(themeName);
          return (
            <DropdownMenuItem
              key={themeName}
              onClick={() => setTheme(themeName)}
              className={cn('flex items-center gap-2 cursor-pointer', isActive && 'bg-card')}
              style={
                isActive && cardColor
                  ? {
                      borderLeft: `3px solid ${cardColor}`
                    }
                  : undefined
              }
            >
              {isActive ? <Check className="h-4 w-4 text-primary" /> : <div className="h-4 w-4" />}
              <span className="capitalize">{themeName}</span>
            </DropdownMenuItem>
          );
        })}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={onLogout}
          className="text-destructive focus:text-destructive cursor-pointer"
        >
          <LogOut className="mr-2 h-4 w-4" />
          Log out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
