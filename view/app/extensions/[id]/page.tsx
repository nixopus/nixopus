'use client';

import PageLayout from '@/packages/layouts/page-layout';
import useExtensionDetails from '../../../packages/hooks/extensions/use-extension-detail';
import { ExtensionInput } from '@/packages/components/extension';
import { Button } from '@nixopus/ui';
import { OverviewTab } from '@/packages/components/extension-tabs';
import { SubPageHeader } from '@nixopus/ui';

export default function ExtensionDetailsPage() {
  const {
    runModalOpen,
    isLoading,
    extension,
    setRunModalOpen,
    t,
    variableColumns,
    handleRunExtension,
    handleChange,
    handleSubmit,
    requiredFields,
    values,
    errors,
    buttonText,
    isOnlyProxyDomain,
    noFieldsToShow,
    actions
  } = useExtensionDetails();

  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <SubPageHeader
        title={extension?.name || ''}
        actions={
          <Button
            className="min-w-[112px]"
            onClick={() => setRunModalOpen(true)}
            disabled={!extension}
          >
            {buttonText}
          </Button>
        }
      />

      <OverviewTab extension={extension} isLoading={isLoading} variableColumns={variableColumns} />

      <ExtensionInput
        open={runModalOpen}
        onOpenChange={setRunModalOpen}
        extension={extension}
        onSubmit={handleRunExtension}
        t={t}
        actions={actions}
        isOnlyProxyDomain={isOnlyProxyDomain}
        noFieldsToShow={noFieldsToShow}
        values={values}
        errors={errors}
        handleChange={handleChange}
        handleSubmit={handleSubmit}
        requiredFields={requiredFields}
      />
    </PageLayout>
  );
}
