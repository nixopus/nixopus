'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { BadgeGroup, BadgeGroupItem } from '@/components/ui/badge-group';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { DataTable, type TableColumn } from '@/components/ui/data-table';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { useTranslation } from '@/hooks/use-translation';
import { useMemo } from 'react';
import { ChevronDown } from 'lucide-react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import React from 'react';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { ChevronRight, Download } from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  formatDataPreview,
  formatLogMessage,
  formatVerboseData
} from '@/packages/hooks/extensions/use-extension-detail';
import { OverviewTabProps, LogsTabProps } from '@/packages/types/extension';

export function OverviewTab({
  extension,
  isLoading,
  parsed,
  variableColumns,
  entryColumns,
  openRunIndex,
  openValidateIndex,
  onToggleRun,
  onToggleValidate
}: OverviewTabProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return <Skeleton className="h-40 w-full" />;
  }

  const runSteps = parsed?.execution?.run || [];
  const validateSteps = parsed?.execution?.validate || [];
  const hasRun = runSteps.length > 0;
  const hasValidate = validateSteps.length > 0;

  const createStepColumns = (
    label: string,
    openIndex: number | null,
    onToggle: (index: number | null) => void
  ): TableColumn<any>[] =>
    useMemo(
      () => [
        {
          key: 'step',
          title: label,
          render: (step, _, index) => {
            const entries = Object.entries(step || {}).filter(
              ([k]) => k !== 'name' && k !== 'type'
            );
            const isOpen = openIndex === index;
            return (
              <Collapsible open={isOpen} onOpenChange={(open) => onToggle(open ? index : null)}>
                <CollapsibleTrigger className="w-full text-left flex items-start gap-2 group">
                  <TypographyMuted className="mt-0.5 shrink-0">{index + 1}.</TypographyMuted>
                  <div className="flex-1 min-w-0">
                    <TypographySmall className="font-medium">
                      {step?.name || step?.type || 'Step'}
                    </TypographySmall>
                    {step?.type && (
                      <Badge variant="outline" className="ml-2">
                        {step.type}
                      </Badge>
                    )}
                  </div>
                  <ChevronDown
                    className={`h-4 w-4 shrink-0 transition-transform duration-200 ${
                      isOpen ? 'rotate-180' : ''
                    }`}
                  />
                </CollapsibleTrigger>
                {entries.length > 0 && entryColumns && (
                  <CollapsibleContent className="mt-3">
                    <DataTable
                      data={entries}
                      columns={entryColumns}
                      showBorder={true}
                      striped={false}
                      containerClassName="rounded-md"
                    />
                  </CollapsibleContent>
                )}
              </Collapsible>
            );
          }
        }
      ],
      [label, openIndex, onToggle, entryColumns]
    );

  const runStepColumns = createStepColumns(
    t('extensions.runSteps') || 'Run steps',
    openRunIndex ?? null,
    onToggleRun || (() => {})
  );

  const validateStepColumns = createStepColumns(
    t('extensions.validateSteps') || 'Validate steps',
    openValidateIndex ?? null,
    onToggleValidate || (() => {})
  );

  return (
    <div className="space-y-6">
      <CardWrapper description={extension?.description}>
        <BadgeGroup>
          {extension?.category && (
            <BadgeGroupItem variant="secondary">{extension.category}</BadgeGroupItem>
          )}
          {extension?.extension_type && (
            <BadgeGroupItem variant="outline">{extension.extension_type}</BadgeGroupItem>
          )}
          {extension?.version && <BadgeGroupItem>v{extension.version}</BadgeGroupItem>}
          {extension?.is_verified && <BadgeGroupItem>Verified</BadgeGroupItem>}
        </BadgeGroup>
      </CardWrapper>

      {extension?.variables && extension.variables.length > 0 && variableColumns && (
        <CardWrapper title="Variables">
          <DataTable
            data={extension.variables}
            columns={variableColumns}
            showBorder={true}
            striped={false}
          />
        </CardWrapper>
      )}

      {(hasRun || hasValidate) && (
        <CardWrapper title={t('extensions.execution') || 'Execution'}>
          <div className="flex flex-col gap-4">
            {hasRun && (
              <DataTable
                data={runSteps}
                columns={runStepColumns}
                showBorder={true}
                striped={false}
                rowClassName="border-0"
              />
            )}
            {hasValidate && (
              <DataTable
                data={validateSteps}
                columns={validateStepColumns}
                showBorder={true}
                striped={false}
                rowClassName="border-0"
              />
            )}
          </div>
        </CardWrapper>
      )}
    </div>
  );
}

export function LogsTab({
  executions,
  executionColumns,
  isLoading,
  open,
  setOpen,
  selectedExecId,
  onOpenLogs,
  formattedLogs,
  collapsedLogs,
  toggleCollapse,
  logsEndRef
}: LogsTabProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return <Skeleton className="h-40 w-full" />;
  }

  return (
    <>
      <CardWrapper title={t('extensions.executions') || 'Executions'}>
        <DataTable
          data={executions || []}
          columns={executionColumns}
          loading={isLoading}
          emptyMessage={t('extensions.noExecutions') || 'No executions yet.'}
          onRowClick={(record) => onOpenLogs(record.id)}
          showBorder={true}
          striped={false}
        />
      </CardWrapper>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent side="right" className="sm:max-w-3xl flex flex-col h-full">
          <SheetHeader>
            <SheetTitle>{t('extensions.logs') || 'Execution Logs'}</SheetTitle>
          </SheetHeader>
          <TypographyMuted className="mb-3 mt-4">
            {t('extensions.executionId') || 'Execution ID'}:{' '}
            <TypographySmall className="font-mono">{selectedExecId}</TypographySmall>
          </TypographyMuted>
          {formattedLogs.length === 0 ? (
            <TypographyMuted className="text-center py-8 flex-1">No logs yet...</TypographyMuted>
          ) : (
            <div className="flex-1 overflow-y-auto space-y-1 pr-2 min-h-0">
              {formattedLogs.map((log) => {
                const hasVerboseData = log.isVerbose && log.data != null;
                const isCollapsed = collapsedLogs.has(log.id);

                return (
                  <div
                    key={log.id}
                    className={cn(
                      'border-l-2 pl-3 py-2 text-sm transition-colors space-y-2',
                      log.color,
                      'bg-muted/30 hover:bg-muted/50'
                    )}
                  >
                    <div className="flex items-start gap-2">
                      <div className="flex-shrink-0 mt-0.5">{log.icon}</div>
                      <div className="flex items-center gap-2 flex-wrap flex-1 min-w-0">
                        <TypographyMuted className="text-xs font-mono">
                          {log.timestamp}
                        </TypographyMuted>
                        <Badge variant="outline" className="text-xs px-1.5 py-0">
                          {log.level}
                        </Badge>
                        <TypographySmall className="font-medium">
                          {formatLogMessage(log.message, log.data)}
                        </TypographySmall>
                      </div>
                    </div>

                    {log.progressInfo && (
                      <div className="bg-accent/50 border border-accent rounded p-2 flex items-center gap-2">
                        <Download className="h-3.5 w-3.5 text-accent-foreground shrink-0" />
                        <div className="flex-1 min-w-0">
                          <TypographySmall className="font-medium">
                            {log.progressInfo.status}
                          </TypographySmall>
                          {log.progressInfo.progress && (
                            <TypographyMuted className="mt-0.5">
                              {log.progressInfo.progress}
                            </TypographyMuted>
                          )}
                        </div>
                      </div>
                    )}

                    {log.data != null && !hasVerboseData && (
                      <div className="bg-muted/50 border rounded p-2">
                        <TypographySmall className="font-mono text-xs break-all">
                          {formatDataPreview(log.data)}
                        </TypographySmall>
                      </div>
                    )}

                    {hasVerboseData && log.data != null && (
                      <Collapsible open={!isCollapsed} onOpenChange={() => toggleCollapse(log.id)}>
                        <CollapsibleTrigger className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground">
                          {isCollapsed ? (
                            <ChevronRight className="h-3 w-3" />
                          ) : (
                            <ChevronDown className="h-3 w-3" />
                          )}
                          <TypographySmall>
                            {isCollapsed ? 'Show verbose output' : 'Hide verbose output'}
                          </TypographySmall>
                        </CollapsibleTrigger>
                        <CollapsibleContent>
                          <div className="bg-muted/50 border rounded p-3 mt-2">
                            <pre className="text-xs font-mono whitespace-pre-wrap break-all overflow-x-auto max-h-96 overflow-y-auto">
                              {formatVerboseData(log.data)}
                            </pre>
                          </div>
                        </CollapsibleContent>
                      </Collapsible>
                    )}
                  </div>
                );
              })}
              <div ref={logsEndRef} />
            </div>
          )}
        </SheetContent>
      </Sheet>
    </>
  );
}
