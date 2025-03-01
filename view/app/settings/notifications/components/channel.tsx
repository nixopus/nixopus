'use client';
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { NotificationChannelField, NotificationChannelProps } from '../utils/types';

const NotificationChannelCard: React.FC<NotificationChannelProps> = ({
  title,
  description,
  icon,
  configData = {},
  onSave,
  isLoading
}) => {
  const [formData, setFormData] = useState<Record<string, string>>(configData);
  const [isFormValid, setIsFormValid] = useState<boolean>(false);
  
  // Add useEffect to update formData when configData changes
  useEffect(() => {
    setFormData(configData);
  }, [configData]);

  const getChannelFields = (): NotificationChannelField[] => {
    switch (title) {
      case 'Email':
        return [
          {
            id: 'smtpServer',
            label: 'SMTP Server',
            placeholder: 'smtp.example.com',
            required: true
          },
          { id: 'port', label: 'Port', placeholder: '587', required: true },
          { id: 'username', label: 'Username', placeholder: 'your@email.com', required: true },
          {
            id: 'password',
            label: 'Password',
            placeholder: '••••••••',
            type: 'password',
            required: true
          },
          {
            id: 'fromEmail',
            label: 'From Email',
            placeholder: 'notifications@yourdomain.com',
            required: true
          },
          { id: 'fromName', label: 'From Name', placeholder: 'Your App Name', required: true }
        ];
      default:
        return [];
    }
  };
  
  const fields = getChannelFields();
  
  useEffect(() => {
    const valid = fields
      .filter((field) => field.required)
      .every((field) => formData[field.id] && formData[field.id].trim() !== '');
    setIsFormValid(valid);
  }, [formData, fields]);
  
  const handleInputChange = (id: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      [id]: value
    }));
  };
  
  const handleSave = () => {
    if (isFormValid) {
      onSave && onSave(formData);
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
      </CardHeader>
      <CardContent>
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
              />
            </div>
          ))}
          
          <div className="pt-2 flex space-x-2 justify-end">
            <Button onClick={handleSave} disabled={!isFormValid || isLoading}>
              Save Configuration
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default NotificationChannelCard;