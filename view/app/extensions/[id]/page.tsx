'use client';

import PageLayout from '@/components/layout/page-layout';
import PageHeader from '@/components/ui/page-header';
import TabsWrapper, { TabsWrapperList } from '@/components/ui/tabs-wrapper';
import useExtensionDetails from '../../../packages/hooks/extensions/use-extension-detail';
import { ExtensionInput } from '@/packages/components/extension';
import { Button } from '@/components/ui/button';
import { OverviewTab } from '@/packages/components/extension-tabs';

export default function ExtensionDetailsPage() {
  const {
    runModalOpen,
    isRunning,
    isLoading,
    tab,
    extension,
    setRunModalOpen,
    t,
    setTab,
    tabs,
    hasExecutions,
    isExecsLoading,
    parsed,
    variableColumns,
    entryColumns,
    openRunIndex,
    openValidateIndex,
    handleRunExtension,
    handleChange,
    handleSubmit,
    requiredFields,
    values,
    errors,
    buttonText,
    isOnlyProxyDomain,
    noFieldsToShow,
    setOpenRunIndex,
    setOpenValidateIndex,
    actions
  } = useExtensionDetails();

  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <TabsWrapper
        value={tab}
        onValueChange={setTab}
        tabs={tabs}
        showTabsCondition={!isExecsLoading && hasExecutions}
        defaultContent={
          <OverviewTab
            extension={extension}
            isLoading={isLoading}
            parsed={parsed}
            variableColumns={variableColumns}
            entryColumns={entryColumns}
            openRunIndex={openRunIndex}
            openValidateIndex={openValidateIndex}
            onToggleRun={setOpenRunIndex}
            onToggleValidate={setOpenValidateIndex}
          />
        }
      >
        <PageHeader
          label={extension?.name || ''}
          description={extension?.author}
          badge={
            extension?.icon ? (
              <div className="h-10 w-10 rounded bg-accent flex items-center justify-center text-lg">
                {extension.icon}
              </div>
            ) : undefined
          }
          actions={
            <Button
              className="min-w-[112px]"
              onClick={() => setRunModalOpen(true)}
              disabled={!extension || isRunning}
            >
              {buttonText}
            </Button>
          }
        >
          <TabsWrapperList className="mt-4" />
        </PageHeader>
      </TabsWrapper>

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
