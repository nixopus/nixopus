'use client';

import PageLayout from '@/components/layout/page-layout';
import useExtensionDetails from '../hooks/use-extension-detail';
import { ExtensionHeader } from './components/ExtensionHeader';
import { RunButton } from './components/RunButton';
import { ExtensionTabs } from './components/ExtensionTabs';
import { ExtensionModal } from './components/ExtensionModal';

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
    setTab
  } = useExtensionDetails();

  const handleRunExtension = async (values: Record<string, unknown>) => {
    if (!extension) return;
    const exec = await runExtension({
      extensionId: extension.extension_id,
      body: { variables: values }
    }).unwrap();
    setRunModalOpen(false);
    router.push(`/extensions/${extension.id}?exec=${exec.id}&openLogs=1`);
  };

  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <div className="flex items-center justify-between">
        <ExtensionHeader extension={extension} isLoading={isLoading} />
        <RunButton
          extension={extension}
          isLoading={isLoading}
          isRunning={isRunning}
          onClick={() => setRunModalOpen(true)}
        />
      </div>

      <ExtensionTabs tab={tab} onTabChange={setTab} extension={extension} isLoading={isLoading} />

      <ExtensionModal
        open={runModalOpen}
        onOpenChange={setRunModalOpen}
        extension={extension}
        onSubmit={handleRunExtension}
      />
    </PageLayout>
  );
}
