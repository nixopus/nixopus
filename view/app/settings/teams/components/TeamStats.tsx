import React from 'react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
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
    <Card className="col-span-1">
      <CardHeader>
        <TypographySmall>{t('settings.teams.stats.title')}</TypographySmall>
        <TypographyMuted>{t('settings.teams.stats.description')}</TypographyMuted>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <TypographyMuted>{t('settings.teams.stats.totalMembers')}</TypographyMuted>
            <span className="font-medium">{users.length}</span>
          </div>
          <div className="flex justify-between items-center">
            <TypographyMuted>{t('settings.teams.stats.owners')}</TypographyMuted>
            <span className="font-medium">{users.filter((u) => u.role === 'Admin').length}</span>
          </div>
          <div className="flex justify-between items-center">
            <TypographyMuted>{t('settings.teams.stats.members')}</TypographyMuted>
            <span className="font-medium">{users.filter((u) => u.role === 'Member').length}</span>
          </div>
          <div className="flex justify-between items-center">
            <TypographyMuted>{t('settings.teams.stats.viewers')}</TypographyMuted>
            <span className="font-medium">{users.filter((u) => u.role === 'Viewer').length}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default TeamStats;
