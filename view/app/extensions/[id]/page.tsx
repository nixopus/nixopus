'use client';

import PageLayout from '@/packages/layouts/page-layout';
import TabsWrapper, { TabsWrapperList } from '@/components/ui/tabs-wrapper';
import useExtensionDetails from '../../../packages/hooks/extensions/use-extension-detail';
import { ExtensionInput, ExtensionForkDialog } from '@/packages/components/extension';
import { Button } from '@/components/ui/button';
import { OverviewTab } from '@/packages/components/extension-tabs';
import { GitFork } from 'lucide-react';
import SubPageHeader from '@/components/ui/sub-page-header';

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
    actions,
    forkOpen,
    setForkOpen,
    forkYaml,
    setForkYaml,
    forkPreview,
    forkVariableColumns,
    doFork,
    isForking
  } = useExtensionDetails();

  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <TabsWrapper
        value={tab}
        onValueChange={setTab}
        tabs={tabs}
        showTabsCondition={!isExecsLoading && hasExecutions}
        className="min-w-fit w-auto"
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
        <SubPageHeader
          title={extension?.name || ''}
          metadata={extension?.author}
          actions={
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                onClick={() => setForkOpen(true)}
                disabled={!extension || extension.parent_extension_id != null}
              >
                <GitFork className="mr-2 h-4 w-4" />
                {t('extensions.fork') || 'Fork'}
              </Button>
              <Button
                className="min-w-[112px]"
                onClick={() => setRunModalOpen(true)}
                disabled={!extension || isRunning}
              >
                {buttonText}
              </Button>
            </div>
          }
        >
          <TabsWrapperList className="mt-4" />
        </SubPageHeader>
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

      {extension && (
        <ExtensionForkDialog
          open={forkOpen}
          onOpenChange={setForkOpen}
          extension={extension}
          t={t}
          forkYaml={forkYaml}
          setForkYaml={setForkYaml}
          preview={forkPreview}
          variableColumns={forkVariableColumns}
          doFork={doFork}
          isLoading={isForking}
        />
      )}
    </PageLayout>
  );
}
