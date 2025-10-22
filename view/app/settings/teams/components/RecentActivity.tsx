import React from 'react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useGetRecentAuditLogsQuery } from '@/redux/services/audit';
import { formatDistanceToNow } from 'date-fns';
import { Loader2, ArrowRight } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { useRouter } from 'next/navigation';

const getActionColor = (actionColor: string) => {
  switch (actionColor) {
    case 'green':
      return 'bg-green-500';
    case 'blue':
      return 'bg-blue-500';
    case 'red':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
};

function RecentActivity() {
  const { t } = useTranslation();
  const router = useRouter();
  const { data: activities, isLoading, error } = useGetRecentAuditLogsQuery();

  const handleViewAll = () => {
    router.push('/activities');
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <TypographySmall>{t('settings.teams.recentActivity.title')}</TypographySmall>
            <TypographyMuted>{t('settings.teams.recentActivity.description')}</TypographyMuted>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleViewAll}
            className="flex items-center gap-1"
          >
            {t('settings.teams.recentActivity.viewAll')}
            <ArrowRight className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex items-center justify-center p-4">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : error ? (
          <div className="p-4 text-red-600">{t('settings.teams.recentActivity.error')}</div>
        ) : activities && activities.length > 0 ? (
          <div className="space-y-4">
            {activities.map((activity) => (
              <div key={activity.id} className="flex items-start gap-4">
                <div
                  className={`h-2 w-2 mt-2 rounded-full ${getActionColor(activity.action_color)}`}
                ></div>
                <div>
                  <TypographySmall>{activity.message}</TypographySmall>
                  <TypographyMuted>
                    {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
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
