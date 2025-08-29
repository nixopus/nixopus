import React from 'react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { useGetRecentAuditLogsQuery } from '@/redux/services/audit';
import { formatDistanceToNow } from 'date-fns';
import { Loader2 } from 'lucide-react';
import { AuditAction, AuditLog } from '@/redux/types/audit';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

const getActionColor = (action: AuditAction) => {
  switch (action) {
    case 'create':
      return 'bg-green-500';
    case 'update':
      return 'bg-blue-500';
    case 'delete':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
};

const getActionMessage = (log: AuditLog, t: (key: string) => string) => {
  const username = log.user?.username || t('settings.teams.recentActivity.actions.defaultUser');
  const resource = log.resource_type
    .split('_')
    .map((word: string) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
    .toLowerCase();

  const actionKey = `settings.teams.recentActivity.actions.${log.action}`;
  return t(actionKey).replace('{username}', username).replace('{resource}', resource);
};

function RecentActivity() {
  const { t } = useTranslation();
  const { data: auditLogs, isLoading, error } = useGetRecentAuditLogsQuery();

  return (
    <Card>
      <CardHeader>
        <TypographySmall>{t('settings.teams.recentActivity.title')}</TypographySmall>
        <TypographyMuted>{t('settings.teams.recentActivity.description')}</TypographyMuted>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex items-center justify-center p-4">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : error ? (
          <div className="p-4 text-red-600">{t('settings.teams.recentActivity.error')}</div>
        ) : auditLogs && auditLogs.length > 0 ? (
          <div className="space-y-4">
            {auditLogs.map((log) => (
              <div key={log.id} className="flex items-start gap-4">
                <div className={`h-2 w-2 mt-2 rounded-full ${getActionColor(log.action)}`}></div>
                <div>
                  <TypographySmall>{getActionMessage(log, t)}</TypographySmall>
                  <TypographyMuted>
                    {formatDistanceToNow(new Date(log.created_at), { addSuffix: true })}
                  </TypographyMuted>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center text-muted-foreground">
            {t('settings.teams.recentActivity.noActivities')}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default RecentActivity;
