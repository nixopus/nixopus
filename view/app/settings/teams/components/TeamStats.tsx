import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

interface TeamStatsProps {
  users: {
    id: string;
    name: string;
    role: 'Admin' | 'Member' | 'Viewer' | 'Owner';
  }[];
}

function TeamStats({ users }: TeamStatsProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-4">
      <div>
        <TypographySmall className="text-sm font-medium">
          {t('settings.teams.stats.title')}
        </TypographySmall>
        <TypographyMuted className="text-xs mt-1">
          {t('settings.teams.stats.description')}
        </TypographyMuted>
      </div>
      <div className="space-y-2">
        <div className="flex justify-between items-center">
          <TypographyMuted className="text-sm">
            {t('settings.teams.stats.totalMembers')}
          </TypographyMuted>
          <span className="font-medium">{users.length}</span>
        </div>
        <div className="flex justify-between items-center">
          <TypographyMuted className="text-sm">{t('settings.teams.stats.owners')}</TypographyMuted>
          <span className="font-medium">{users.filter((u) => u.role === 'Admin').length}</span>
        </div>
        <div className="flex justify-between items-center">
          <TypographyMuted className="text-sm">{t('settings.teams.stats.members')}</TypographyMuted>
          <span className="font-medium">{users.filter((u) => u.role === 'Member').length}</span>
        </div>
        <div className="flex justify-between items-center">
          <TypographyMuted className="text-sm">{t('settings.teams.stats.viewers')}</TypographyMuted>
          <span className="font-medium">{users.filter((u) => u.role === 'Viewer').length}</span>
        </div>
      </div>
    </div>
  );
}

export default TeamStats;
