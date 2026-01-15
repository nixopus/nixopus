import { SelectOption } from '@/components/ui/select-wrapper';
import { SettingsCategory } from '@/packages/hooks/shared/use-settings-categories';
import { UserTypes } from '@/redux/types/orgs';
import { UserSettings, User as UserType } from '@/redux/types/user';
import {
  SMTPConfig,
  WebhookConfig,
  SMTPFormData,
  PreferenceType
} from '@/redux/types/notification';

export interface TeamStatsProps {
  users: {
    id: string;
    name: string;
    role: 'Admin' | 'Member' | 'Viewer' | 'Owner';
  }[];
}

export type EditUser = {
  id: string;
  name: string;
  email: string;
  avatar: string;
  role: 'Owner' | 'Admin' | 'Member' | 'Viewer';
  permissions: string[];
};

export type RoleType = 'owner' | 'admin' | 'member' | 'viewer';

export interface TeamMembersProps {
  users: EditUser[];
  handleRemoveUser: (userId: string) => void;
  getRoleBadgeVariant: (role: string) => 'default' | 'secondary' | 'destructive' | 'outline';
  onUpdateUser: (userId: string, role: UserTypes) => Promise<void>;
}

export const MAX_VISIBLE_PERMISSIONS = 3;

export interface EditUserDialogProps {
  isOpen: boolean;
  onClose: () => void;
  user: {
    id: string;
    name: string;
    email: string;
    avatar: string;
    role: 'Owner' | 'Admin' | 'Member' | 'Viewer';
  };
  onSave: (userId: string, role: UserTypes) => void;
}

export const AVAILABLE_ROLEs: SelectOption[] = [
  {
    value: 'admin',
    label: 'Admin'
  },
  {
    value: 'member',
    label: 'Member'
  },
  {
    value: 'viewer',
    label: 'Viewer'
  }
];

export interface EditTeamProps {
  isEditTeamDialogOpen: boolean;
  setEditTeamDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  handleUpdateTeam: () => void;
  teamName: string;
  setTeamName: React.Dispatch<React.SetStateAction<string>>;
  teamDescription: string;
  setTeamDescription: React.Dispatch<React.SetStateAction<string>>;
  isUpdating: boolean;
}

export interface AddMemberProps {
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

export interface AccountSectionProps {
  username: string;
  setUsername: (username: string) => void;
  usernameError: string;
  usernameSuccess: boolean;
  setUsernameError: (error: string) => void;
  email: string;
  isLoading: boolean;
  handleUsernameChange: () => void;
  user: UserType;
  userSettings: UserSettings;
  isGettingUserSettings: boolean;
  isUpdatingFont: boolean;
  isUpdatingTheme: boolean;
  isUpdatingLanguage: boolean;
  isUpdatingAutoUpdate: boolean;
  handleThemeChange: (theme: string) => void;
  handleLanguageChange: (language: string) => void;
  handleAutoUpdateChange: (autoUpdate: boolean) => void;
  handleFontUpdate: (fontFamily: string, fontSize: number) => Promise<void>;
}

export interface AvatarSectionProps {
  onImageChange: (imageUrl: string | null) => void;
  user: UserType;
}

export interface SecuritySectionProps {}

export interface SettingsSidebarProps {
  categories: SettingsCategory[];
  activeCategory: string;
  onCategoryChange: (category: string) => void;
}

export interface ChannelTabProps {
  smtpConfigs?: SMTPConfig;
  slackConfig?: WebhookConfig;
  discordConfig?: WebhookConfig;
  isLoading: boolean;
  handleOnSave: (data: SMTPFormData) => void;
  handleOnSaveSlack: (data: Record<string, string>) => void;
  handleOnSaveDiscord: (data: Record<string, string>) => void;
}

export interface NotificationPreferenceCardProps {
  title: string;
  description: string;
  preferences?: PreferenceType[];
  onUpdate: (id: string, enabled: boolean) => void;
}

export interface NotificationPreferencesTabProps {
  activityPreferences?: PreferenceType[];
  securityPreferences?: PreferenceType[];
  onUpdatePreference: (id: string, enabled: boolean) => void;
}
