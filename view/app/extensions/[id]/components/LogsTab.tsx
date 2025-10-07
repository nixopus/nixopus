'use client';

import { useTranslation } from '@/hooks/use-translation';
import { useParams } from 'next/navigation';
import { useListExecutionsQuery } from '@/redux/services/extensions/extensionsApi';
import { Skeleton } from '@/components/ui/skeleton';

export default function ExecutionsTab() {
  const { t } = useTranslation();
  const params = useParams();
  const id = (params?.id as string) || '';
  const { data: executions, isLoading } = useListExecutionsQuery({ extensionId: id }, { skip: !id });

  return (
    <div className="space-y-3">
      <div className="rounded-md border overflow-hidden">
        <div className="grid grid-cols-12 bg-muted/50 px-3 py-2 text-xs font-medium text-muted-foreground">
          <div className="col-span-4">{t('extensions.executionId') || 'Execution ID'}</div>
          <div className="col-span-2">{t('extensions.status') || 'Status'}</div>
          <div className="col-span-3">{t('extensions.startedAt') || 'Started At'}</div>
          <div className="col-span-3">{t('extensions.completedAt') || 'Completed At'}</div>
        </div>
        <div className="divide-y">
          {isLoading && Array.from({ length: 5 }).map((_, i) => (
            <div key={`s-${i}`} className="grid grid-cols-12 px-3 py-3 text-sm">
              <div className="col-span-4"><Skeleton className="h-4 w-40" /></div>
              <div className="col-span-2"><Skeleton className="h-4 w-20" /></div>
              <div className="col-span-3"><Skeleton className="h-4 w-28" /></div>
              <div className="col-span-3"><Skeleton className="h-4 w-28" /></div>
            </div>
          ))}
          {!isLoading && (executions || []).map((e) => (
            <div key={e.id} className="grid grid-cols-12 px-3 py-3 text-sm">
              <div className="col-span-4 truncate">{e.id}</div>
              <div className="col-span-2 text-muted-foreground capitalize">{e.status}</div>
              <div className="col-span-3 text-muted-foreground">{new Date(e.started_at).toLocaleString()}</div>
              <div className="col-span-3 text-muted-foreground">
                {e.completed_at ? new Date(e.completed_at).toLocaleString() : '-'}
              </div>
            </div>
          ))}
          {!isLoading && (!executions || executions.length === 0) && (
            <div className="px-3 py-6 text-sm text-muted-foreground">
              {t('extensions.noExecutions') || 'No executions yet.'}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
