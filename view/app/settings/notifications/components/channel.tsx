"use client"
import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { NotificationChannelField, NotificationChannelProps } from '../utils/types'

const NotificationChannelCard: React.FC<NotificationChannelProps> = ({
    title,
    description,
    icon,
    connected = false,
    configData = {},
    onConnect,
    onDisconnect,
    onSave
}) => {
    const [isConnected, setIsConnected] = useState<boolean>(connected)
    const [formData, setFormData] = useState<Record<string, string>>(configData)
    const [isFormValid, setIsFormValid] = useState<boolean>(false)
    const [isTesting, setIsTesting] = useState<boolean>(false)
    const [isExpanded, setIsExpanded] = useState<boolean>(false)

    const getChannelFields = (): NotificationChannelField[] => {
        switch (title) {
            case 'Email':
                return [
                    { id: 'smtpServer', label: 'SMTP Server', placeholder: 'smtp.example.com', required: true },
                    { id: 'port', label: 'Port', placeholder: '587', required: true },
                    { id: 'username', label: 'Username', placeholder: 'your@email.com', required: true },
                    { id: 'password', label: 'Password', placeholder: '••••••••', type: 'password', required: true },
                    { id: 'fromEmail', label: 'From Email', placeholder: 'notifications@yourdomain.com', required: true },
                    { id: 'fromName', label: 'From Name', placeholder: 'Your App Name', required: true }
                ];
            case 'Slack':
                return [
                    { id: 'webhookUrl', label: 'Webhook URL', placeholder: 'https://hooks.slack.com/services/...', required: true },
                    { id: 'channel', label: 'Default Channel', placeholder: '#general', required: true },
                    { id: 'username', label: 'Bot Username', placeholder: 'NotificationBot', required: false },
                    { id: 'iconUrl', label: 'Bot Icon URL', placeholder: 'https://example.com/icon.png', required: false }
                ];
            case 'Discord':
                return [
                    { id: 'webhookUrl', label: 'Webhook URL', placeholder: 'https://discord.com/api/webhooks/...', required: true },
                    { id: 'username', label: 'Bot Username', placeholder: 'NotificationBot', required: false },
                    { id: 'avatarUrl', label: 'Avatar URL', placeholder: 'https://example.com/avatar.png', required: false }
                ];
            default:
                return [];
        }
    };

    const fields = getChannelFields();

    useEffect(() => {
        const valid = fields
            .filter(field => field.required)
            .every(field => formData[field.id] && formData[field.id].trim() !== '');
        setIsFormValid(valid);
    }, [formData, fields]);

    const handleInputChange = (id: string, value: string) => {
        setFormData(prev => ({
            ...prev,
            [id]: value
        }));
    };

    const handleToggle = () => {
        if (isConnected) {
            setIsConnected(false);
            onDisconnect && onDisconnect();
        } else {
            if (isFormValid) {
                setIsConnected(true);
                onConnect && onConnect(formData);
            } else {
                setIsExpanded(true);
            }
        }
    };

    const handleSave = () => {
        if (isFormValid) {
            onSave && onSave(formData);
            setIsExpanded(false);
        }
    };

    const handleTest = async () => {
        setIsTesting(true);
        try {
            await new Promise(resolve => setTimeout(resolve, 1500));
            alert(`Test ${title} notification sent successfully!`);
        } catch (error) {
            alert(`Failed to send test ${title} notification`);
        } finally {
            setIsTesting(false);
        }
    };

    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <div className="flex items-center space-x-3">
                    {icon}
                    <div>
                        <CardTitle className="text-lg">{title}</CardTitle>
                        <CardDescription>{description}</CardDescription>
                    </div>
                </div>
                <div className="flex items-center space-x-2">
                    <Badge variant={isConnected ? "default" : "outline"}>
                        {isConnected ? "Connected" : "Disconnected"}
                    </Badge>
                    <Switch checked={isConnected} onCheckedChange={handleToggle} />
                </div>
            </CardHeader>
            <CardContent>
                <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
                    <CollapsibleTrigger asChild>
                        <Button variant="ghost" size="sm" className="flex items-center w-full justify-between">
                            <span>Configuration</span>
                            {isExpanded ?
                                <ChevronUp className="h-4 w-4" /> :
                                <ChevronDown className="h-4 w-4" />
                            }
                        </Button>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                        <div className="space-y-4 pt-4">
                            {fields.map((field) => (
                                <div className="space-y-2" key={field.id}>
                                    <Label htmlFor={`${title.toLowerCase()}-${field.id}`}>
                                        {field.label} {field.required && <span className="text-red-500">*</span>}
                                    </Label>
                                    <Input
                                        id={`${title.toLowerCase()}-${field.id}`}
                                        type={field.type || 'text'}
                                        value={formData[field.id] || ''}
                                        onChange={(e) => handleInputChange(field.id, e.target.value)}
                                        placeholder={field.placeholder}
                                        disabled={isConnected}
                                    />
                                </div>
                            ))}

                            <div className="pt-2 flex space-x-2 justify-end">
                                {isConnected ? (
                                    <>
                                        <Button
                                            variant="outline"
                                            onClick={handleToggle}
                                        >
                                            Disconnect
                                        </Button>
                                        <Button
                                            onClick={handleTest}
                                            disabled={isTesting}
                                        >
                                            {isTesting ? "Sending..." : "Test Connection"}
                                        </Button>
                                    </>
                                ) : (
                                    <Button
                                        onClick={handleSave}
                                        disabled={!isFormValid}
                                    >
                                        Save Configuration
                                    </Button>
                                )}
                            </div>
                        </div>
                    </CollapsibleContent>
                </Collapsible>
            </CardContent>
        </Card>
    )
}

export default NotificationChannelCard;