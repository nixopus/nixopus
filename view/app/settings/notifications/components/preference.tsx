'use client';
import React from 'react';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { PreferenceType } from '@/redux/types/notification';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

export interface NotificationPreferenceCardProps {
  title: string;
  description: string;
  preferences?: PreferenceType[];
  onUpdate: (id: string, enabled: boolean) => void;
}

const NotificationPreferenceCard: React.FC<NotificationPreferenceCardProps> = ({
  title,
  description,
  preferences,
  onUpdate
}) => {
  return (
    <div className="space-y-4">
      <div>
        <TypographySmall className="text-sm font-medium">{title}</TypographySmall>
        <TypographyMuted className="text-xs mt-1">{description}</TypographyMuted>
      </div>
      <div className="space-y-4">
        {preferences?.map((pref) => (
          <div className="flex items-center justify-between" key={pref.id}>
            <div className="space-y-0.5">
              <Label htmlFor={pref.id} className="text-base">
                {pref.label}
              </Label>
              <TypographyMuted className="text-xs">{pref.description}</TypographyMuted>
            </div>
            <Switch
              id={pref.id}
              defaultChecked={pref.enabled}
              onCheckedChange={(enabled) => onUpdate?.(pref.id, enabled)}
            />
          </div>
        ))}
      </div>
    </div>
  );
};

export default NotificationPreferenceCard;
