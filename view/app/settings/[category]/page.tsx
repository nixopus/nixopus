'use client';

import { useParams } from 'next/navigation';
import { SettingsContent } from '@/packages/components/settings';

const VALID_CATEGORIES = [
  'general',
  'notifications',
  'keyboard-shortcuts',
  'network',
  'terminal',
  'container',
  'agent',
  'troubleshooting'
];

export default function SettingsPage() {
  const params = useParams();
  const category = params.category as string;

  if (!VALID_CATEGORIES.includes(category)) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Settings page not found.</p>
      </div>
    );
  }

  return <SettingsContent activeCategory={category} />;
}
