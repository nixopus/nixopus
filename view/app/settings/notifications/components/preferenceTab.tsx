'use client';
import React from 'react';
import NotificationPreferenceCard from './preference';
import { PreferenceType } from '@/redux/types/notification';
import { useTranslation } from '@/hooks/use-translation';

interface NotificationPreferencesTabProps {
  activityPreferences?: PreferenceType[];
  securityPreferences?: PreferenceType[];
  onUpdatePreference: (id: string, enabled: boolean) => void;
}

export const NotificationPreferencesTab: React.FC<NotificationPreferencesTabProps> = ({
  activityPreferences,
  securityPreferences,
  onUpdatePreference
}) => {
  const { t } = useTranslation();

  return (
    <div className="grid gap-6 md:grid-cols-1">
      <NotificationPreferenceCard
        title={t('settings.notifications.preferences.activity.title')}
        description={t('settings.notifications.preferences.activity.description')}
        preferences={activityPreferences}
        onUpdate={onUpdatePreference}
      />

      <NotificationPreferenceCard
        title={t('settings.notifications.preferences.security.title')}
        description={t('settings.notifications.preferences.security.description')}
        preferences={securityPreferences}
        onUpdate={onUpdatePreference}
      />
    </div>
  );
};

export default NotificationPreferencesTab;
