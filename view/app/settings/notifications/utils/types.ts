import { ReactNode } from 'react';

export interface NotificationPreference {
    id: string;
    label: string;
    description: string;
    defaultValue: boolean;
}

export interface NotificationChannelField {
    id: string;
    label: string;
    placeholder: string;
    type?: string;
    required: boolean;
}

export interface NotificationChannelProps {
    title: string;
    description: string;
    icon: ReactNode;
    connected?: boolean;
    configData?: Record<string, string>;
    onConnect?: (data: Record<string, string>) => void;
    onDisconnect?: () => void;
    onSave?: (data: Record<string, string>) => void;
}

export interface NotificationPreferenceCardProps {
    title: string;
    description: string;
    preferences: NotificationPreference[];
}

export interface NotificationService {
    saveEmailConfig: (config: Record<string, string>) => Promise<{ success: boolean }>;
    saveWebhookConfig: (platform: string, config: Record<string, string>) => Promise<{ success: boolean }>;
    testEmailConnection: (config: Record<string, string>) => Promise<{ success: boolean; message: string }>;
    testWebhookConnection: (platform: string, config: Record<string, string>) => Promise<{ success: boolean; message: string }>;
    saveNotificationPreferences: (preferences: Record<string, boolean>) => Promise<{ success: boolean }>;
}