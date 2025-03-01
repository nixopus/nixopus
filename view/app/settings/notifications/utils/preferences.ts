import { NotificationPreference } from './types';

export const activityPreferences: NotificationPreference[] = [
  {
    id: 'team-updates',
    label: 'Team Updates',
    description: 'When team members join or leave your team',
    defaultValue: true
  }
];

export const securityPreferences: NotificationPreference[] = [
  {
    id: 'login-alerts',
    label: 'Login Alerts',
    description: 'When a new device logs into your account',
    defaultValue: true
  },
  {
    id: 'password-changes',
    label: 'Password Changes',
    description: 'When your password is changed',
    defaultValue: true
  },
  {
    id: 'security-alerts',
    label: 'Security Alerts',
    description: 'Important security notifications',
    defaultValue: true
  }
];

export const updatePreferences: NotificationPreference[] = [
  {
    id: 'product-updates',
    label: 'Product Updates',
    description: 'New features and improvements',
    defaultValue: true
  },
  {
    id: 'newsletter',
    label: 'Newsletter',
    description: 'Our monthly newsletter with tips and updates',
    defaultValue: false
  },
  {
    id: 'marketing',
    label: 'Marketing',
    description: 'Promotions and special offers',
    defaultValue: false
  }
];
