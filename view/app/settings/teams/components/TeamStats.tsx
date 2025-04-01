import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';

interface TeamStatsProps {
  users: {
    id: string;
    name: string;
    role: 'admin' | 'member' | 'viewer' | 'owner';
  }[];
}

function TeamStats({ users }: TeamStatsProps) {
  const { t } = useTranslation();

  return (
    <Card className="col-span-1">
      <CardHeader>
        <CardTitle>{t('settings.teams.stats.title')}</CardTitle>
        <CardDescription>{t('settings.teams.stats.description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">
              {t('settings.teams.stats.totalMembers')}
            </span>
            <span className="font-medium">{users.length}</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">
              {t('settings.teams.stats.owners')}
            </span>
            <span className="font-medium">{users.filter((u) => u.role === 'admin').length}</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">
              {t('settings.teams.stats.members')}
            </span>
            <span className="font-medium">{users.filter((u) => u.role === 'member').length}</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">
              {t('settings.teams.stats.viewers')}
            </span>
            <span className="font-medium">{users.filter((u) => u.role === 'viewer').length}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default TeamStats;
