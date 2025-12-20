'use client';

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Info, Terminal } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import OverviewTab from './OverviewTab';
import ExecutionsTab from './LogsTab';
import type { Extension } from '@/redux/types/extension';
import { useParams } from 'next/navigation';
import { useListExecutionsQuery } from '@/redux/services/extensions/extensionsApi';
import { useEffect, useMemo } from 'react';

interface ExtensionTabsProps {
  tab: string;
  onTabChange: (value: string) => void;
  extension?: Extension;
  isLoading: boolean;
}

export function ExtensionTabs({ tab, onTabChange, extension, isLoading }: ExtensionTabsProps) {
  const { t } = useTranslation();
  const params = useParams();
  const id = (params?.id as string) || '';

  const { data: executions, isLoading: isExecsLoading } = useListExecutionsQuery(
    { extensionId: id },
    { skip: !id }
  );

  const hasExecutions = useMemo(() => (executions || []).length > 0, [executions]);

  useEffect(() => {
    if (tab === 'executions' && !isExecsLoading && !hasExecutions) {
      onTabChange('overview');
    }
  }, [tab, hasExecutions, isExecsLoading, onTabChange]);

  if (!isExecsLoading && !hasExecutions) {
    return (
      <div className="mt-6">
        <OverviewTab extension={extension} isLoading={isLoading} />
      </div>
    );
  }

  return (
    <div className="mt-6">
      <Tabs value={tab} onValueChange={onTabChange} className="w-full">
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
  );
}
