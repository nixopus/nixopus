"use client"
import React from 'react'
import { Mail, MessageSquare } from 'lucide-react'
import notificationService from '../utils/notification-service';
import NotificationChannelCard from './channel';


export const NotificationChannelsTab: React.FC = () => {
    const handleConnectEmail = (data: Record<string, string>) => {
        notificationService.saveEmailConfig(data)
            .then(response => {
                console.log('Connected email', response);
            })
            .catch(error => {
                console.error('Error connecting email', error);
            });
    };

    const handleConnectSlack = (data: Record<string, string>) => {
        notificationService.saveWebhookConfig('Slack', data)
            .then(response => {
                console.log('Connected Slack', response);
            })
            .catch(error => {
                console.error('Error connecting Slack', error);
            });
    };

    const handleConnectDiscord = (data: Record<string, string>) => {
        notificationService.saveWebhookConfig('Discord', data)
            .then(response => {
                console.log('Connected Discord', response);
            })
            .catch(error => {
                console.error('Error connecting Discord', error);
            });
    };

    return (
        <div className="grid gap-6 md:grid-cols-1">
            <NotificationChannelCard
                title="Email"
                description="Configure SMTP settings to send email notifications"
                icon={<Mail className="h-5 w-5 text-primary" />}
                connected={false}
                configData={{
                    smtpServer: '',
                    port: '587',
                    username: '',
                    password: '',
                    fromEmail: '',
                    fromName: 'Your App Name'
                }}
                onConnect={handleConnectEmail}
                onDisconnect={() => console.log('Disconnect email')}
                onSave={(data) => console.log('Save email config', data)}
            />

            <NotificationChannelCard
                title="Slack"
                description="Use Slack webhooks to send notifications to your workspace"
                icon={<MessageSquare className="h-5 w-5 text-indigo-500" />}
                connected={false}
                configData={{
                    webhookUrl: '',
                    channel: '#notifications',
                    username: 'NotificationBot'
                }}
                onConnect={handleConnectSlack}
                onDisconnect={() => console.log('Disconnect Slack')}
                onSave={(data) => console.log('Save Slack config', data)}
            />

            <NotificationChannelCard
                title="Discord"
                description="Use Discord webhooks to send notifications to your server"
                icon={<MessageSquare className="h-5 w-5 text-purple-500" />}
                connected={false}
                configData={{
                    webhookUrl: '',
                    username: 'NotificationBot'
                }}
                onConnect={handleConnectDiscord}
                onDisconnect={() => console.log('Disconnect Discord')}
                onSave={(data) => console.log('Save Discord config', data)}
            />
        </div>
    );
};

export default NotificationChannelsTab;