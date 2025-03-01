'use client';
import React from 'react';
import { Mail } from 'lucide-react';
import NotificationChannelCard from './channel';
import useNotificationSettings from '../../hooks/use-notification-settings';

export const NotificationChannels: React.FC = () => {
    const {
        smtpConfigs,
        isLoading,
        error,
        isCreating,
        isUpdating,
        handleOnSave
    } = useNotificationSettings();
    console.log('smtpConfigs', smtpConfigs);
    return (
        <div className="grid gap-6 md:grid-cols-1">
            <NotificationChannelCard
                title="Email"
                description="Configure SMTP settings to send email notifications"
                icon={<Mail className="h-5 w-5 text-primary" />}
                configData={{
                    smtpServer: smtpConfigs?.host || '',
                    port: smtpConfigs?.port.toString() || '587',
                    username: smtpConfigs?.username || '',
                    password: smtpConfigs?.password || '',
                    fromEmail: smtpConfigs?.from_email || '',
                    fromName: smtpConfigs?.from_name || ''
                }}
                onSave={(data) => handleOnSave(data)}
                isLoading={
                    isLoading ||
                    isCreating ||
                    isUpdating
                }
            />
        </div>
    );
};

export default NotificationChannels;
