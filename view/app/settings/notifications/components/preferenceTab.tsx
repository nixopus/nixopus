'use client';
import React from 'react';
import { NotificationPreference } from '../utils/types';
import NotificationPreferenceCard from './preference';

interface NotificationPreferencesTabProps {
  activityPreferences: NotificationPreference[];
  securityPreferences: NotificationPreference[];
  updatePreferences: NotificationPreference[];
}

export const NotificationPreferencesTab: React.FC<NotificationPreferencesTabProps> = ({
  activityPreferences,
  securityPreferences,
  updatePreferences
}) => {
  return (
    <div className="grid gap-6 md:grid-cols-1">
      <NotificationPreferenceCard
        title="Activity Notifications"
        description="Notifications about activity in your team"
        preferences={activityPreferences}
      />

      <NotificationPreferenceCard
        title="Security Notifications"
        description="Important alerts about deployed applications and errors"
        preferences={securityPreferences}
      />

      <NotificationPreferenceCard
        title="Updates & Marketing"
        description="Stay informed about our product and services"
        preferences={updatePreferences}
      />
    </div>
  );
};

export default NotificationPreferencesTab;
