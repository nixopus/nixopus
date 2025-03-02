'use client';
import React from 'react';
import NotificationPreferenceCard from './preference';
import { PreferenceType } from '@/redux/types/notification';

interface NotificationPreferencesTabProps {
  activityPreferences?: PreferenceType[];
  securityPreferences?: PreferenceType[];
  updatePreferences?: PreferenceType[];
  onUpdatePreference: (id: string, enabled: boolean) => void;
}

export const NotificationPreferencesTab: React.FC<NotificationPreferencesTabProps> = ({
  activityPreferences,
  securityPreferences,
  updatePreferences,
  onUpdatePreference
}) => {
  return (
    <div className="grid gap-6 md:grid-cols-1">
      <NotificationPreferenceCard
        title="Activity Notifications"
        description="Notifications about activity in your team"
        preferences={activityPreferences}
        onUpdate={onUpdatePreference}
      />

      <NotificationPreferenceCard
        title="Security Notifications"
        description="Important alerts about deployed applications and errors"
        preferences={securityPreferences}
        onUpdate={onUpdatePreference}
      />

      <NotificationPreferenceCard
        title="Updates & Marketing"
        description="Stay informed about our product and services"
        preferences={updatePreferences}
        onUpdate={onUpdatePreference}
      />
    </div>
  );
};

export default NotificationPreferencesTab;
