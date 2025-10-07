'use client';

import { useParams } from 'next/navigation';
import PageLayout from '@/components/layout/page-layout';
import { useTranslation } from '@/hooks/use-translation';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import {  Info, Terminal } from 'lucide-react';
import { useGetExtensionQuery } from '@/redux/services/extensions/extensionsApi';
import OverviewTab from './components/OverviewTab';
import ExecutionsTab from './components/LogsTab';

export default function ExtensionDetailsPage() {
  const { t } = useTranslation();
  const params = useParams();
  const id = (params?.id as string) || '';

  const { data: extension, isLoading } = useGetExtensionQuery({ id });

  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {isLoading ? (
            <Skeleton className="h-6 w-48" />
          ) : (
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded bg-accent flex items-center justify-center text-lg">
                {extension?.icon}
              </div>
              <div>
                <div className="text-xl font-semibold">{extension?.name}</div>
                <div className="text-sm text-muted-foreground">{extension?.author}</div>
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="mt-6">
        <Tabs defaultValue="overview" className="w-full">
          <TabsList>
            <TabsTrigger value="overview">
              <Info className="mr-2 h-4 w-4" />
              {t('extensions.overview') || 'Overview'}
            </TabsTrigger>
            <TabsTrigger value="executions">
              <Terminal className="mr-2 h-4 w-4" />
              {t('extensions.executions') || 'Executions'}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="mt-6">
            <OverviewTab extension={extension} isLoading={isLoading} />
          </TabsContent>

          <TabsContent value="executions" className="mt-6">
            <ExecutionsTab />
          </TabsContent>
        </Tabs>
      </div>
    </PageLayout>
  );
}


