'use client';

import PageLayout from '@/components/layout/page-layout';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { Info, Terminal } from 'lucide-react';
import OverviewTab from './components/OverviewTab';
import ExecutionsTab from './components/LogsTab';
import { Button } from '@/components/ui/button';
import ExtensionInput from '@/app/extensions/components/extension-input';
import useExtensionDetails from '../hooks/use-extension-detail';

export default function ExtensionDetailsPage() {
  const {
    runModalOpen,
    runExtension,
    isRunning,
    isLoading,
    tab,
    extension,
    router,
    setRunModalOpen,
    t,
    setTab
  } = useExtensionDetails();

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
        <div>
          {isLoading ? (
            <Skeleton className="h-9 w-28" />
          ) : (
            <Button
              className="min-w-[112px]"
              onClick={() => setRunModalOpen(true)}
              disabled={!extension || isRunning}
            >
              {extension?.extension_type === 'install'
                ? t('extensions.install') || 'Install'
                : t('extensions.run') || 'Run'}
            </Button>
          )}
        </div>
      </div>

      <div className="mt-6">
        <Tabs value={tab} onValueChange={setTab} className="w-full">
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
      <ExtensionInput
        open={runModalOpen}
        onOpenChange={setRunModalOpen}
        extension={extension}
        onSubmit={async (values) => {
          if (!extension) return;
          const exec = await runExtension({
            extensionId: extension.extension_id,
            body: { variables: values }
          }).unwrap();
          setRunModalOpen(false);
          router.push(`/extensions/${extension.id}?exec=${exec.id}&openLogs=1`);
        }}
      />
    </PageLayout>
  );
}
