import { useEffect, useMemo, useState } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useGetExecutionLogsQuery,
  useGetExtensionQuery,
  useRunExtensionMutation,
  useForkExtensionMutation
} from '@/redux/services/extensions/extensionsApi';
import { useListExecutionsQuery } from '@/redux/services/extensions/extensionsApi';
import { useRef } from 'react';
import { FormattedLog } from '../../types/extension';
import { ExtensionLog } from '@/redux/types/extension';
import { Clock, Info, Terminal } from 'lucide-react';
import { CheckCircle2 } from 'lucide-react';
import { XCircle } from 'lucide-react';
import { Loader2 } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { DialogAction } from '@/components/ui/dialog-wrapper';
import { useExtensionInput } from '@/packages/hooks/extensions/use-extension-input';
import YAML from 'yaml';
import { TableColumn } from '@/components/ui/data-table';
import { Extension, ExtensionExecution } from '@/redux/types/extension';
import { TypographyMuted, TypographySmall } from '@/components/ui/typography';
import { LogsTab, OverviewTab } from '@/packages/components/extension-tabs';
import { toast } from 'sonner';
import { VariableData } from '@/packages/types/extension';

type LogState = 'info' | 'success' | 'error' | 'muted';

function getLogColor(state: LogState): string {
  const colorMap: Record<LogState, string> = {
    info: 'text-primary',
    success: 'text-green-600 dark:text-green-400',
    error: 'text-destructive',
    muted: 'text-muted-foreground'
  };
  return colorMap[state];
}

function getLogIcon(state: LogState): React.ReactNode {
  const iconMap: Record<LogState, React.ReactNode> = {
    info: <Clock className="h-4 w-4" />,
    success: <CheckCircle2 className="h-4 w-4" />,
    error: <XCircle className="h-4 w-4" />,
    muted: undefined
  };
  return iconMap[state];
}

function getLogState(
  log: ExtensionLog,
  stepCompletedIds: Set<string>,
  stepFailedIds: Set<string>
): LogState {
  const isStepCompleted = Boolean(log.step_id && stepCompletedIds.has(log.step_id));
  const isStepFailed = Boolean(log.step_id && stepFailedIds.has(log.step_id));

  if (log.message === 'execution_started') {
    return 'info';
  }
  if (log.message === 'execution_completed') {
    return 'success';
  }
  if (log.message.startsWith('step_started')) {
    if (isStepFailed) return 'error';
    if (isStepCompleted) return 'success';
    return 'info';
  }
  if (log.message.startsWith('step_completed')) {
    return 'success';
  }
  if (log.message.startsWith('step_failed')) {
    return 'error';
  }
  if (log.level === 'error' || log.level === 'ERROR') {
    return 'error';
  }
  return 'muted';
}

function useExtensionDetails() {
  const { t } = useTranslation();
  const params = useParams();
  const search = useSearchParams();
  const router = useRouter();
  const id = (params?.id as string) || '';
  const [selectedExecId, setSelectedExecId] = useState<string | null>(null);
  const [afterSeq, setAfterSeq] = useState<number>(0);
  const [allLogs, setAllLogs] = useState<ExtensionLog[]>([]);
  const [open, setOpen] = useState(false);
  const [collapsedLogs, setCollapsedLogs] = useState<Set<string>>(new Set());
  const logsEndRef = useRef<HTMLDivElement>(null);
  const prevExecIdRef = useRef<string | null>(null);
  const prevExtensionIdRef = useRef<string>(id);
  const [executionStatus, setExecutionStatus] = useState<string | null>(null);
  const poll =
    !!selectedExecId &&
    open &&
    executionStatus !== 'completed' &&
    executionStatus !== 'failed' &&
    executionStatus !== 'cancelled';
  const shouldFetch = !!selectedExecId && open;
  const { data: logsResp, refetch } = useGetExecutionLogsQuery(
    { executionId: selectedExecId || '', afterSeq, limit: 200 },
    { skip: !shouldFetch, pollingInterval: poll ? 2500 : undefined }
  );
  const { data: extension, isLoading } = useGetExtensionQuery({ id });
  const [tab, setTab] = useState<string>('overview');
  const [runModalOpen, setRunModalOpen] = useState(false);
  const [runExtension, { isLoading: isRunning }] = useRunExtensionMutation();
  const [forkOpen, setForkOpen] = useState(false);
  const [forkYaml, setForkYaml] = useState<string>('');
  const [forkExtension, { isLoading: isForking }] = useForkExtensionMutation();

  const handleRunExtension = async (values: Record<string, unknown>) => {
    if (!extension) return;
    const exec = await runExtension({
      extensionId: extension.extension_id,
      body: { variables: values }
    }).unwrap();
    setRunModalOpen(false);
    router.push(`/extensions/${extension.id}?exec=${exec.id}&openLogs=1`);
  };

  const { values, errors, handleChange, handleSubmit, requiredFields } = useExtensionInput({
    extension,
    open: runModalOpen,
    onSubmit: handleRunExtension,
    onClose: () => setRunModalOpen(false)
  });
  const { data: executions, isLoading: isExecsLoading } = useListExecutionsQuery(
    {
      extensionId: id
    },
    {
      skip: !id
    }
  );
  const initializedDefaultTab = useRef(false);
  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: () => setRunModalOpen(false),
      variant: 'ghost'
    }
  ];
  const isOnlyProxyDomain =
    requiredFields.length === 1 &&
    (requiredFields[0].variable_name.toLowerCase() === 'proxy_domain' ||
      requiredFields[0].variable_name.toLowerCase() === 'domain');
  const noFieldsToShow = requiredFields.length === 0;

  const hasExecutions = useMemo(() => (executions || []).length > 0, [executions]);

  useEffect(() => {
    if (tab === 'executions' && !isExecsLoading && !hasExecutions) {
      setTab('overview');
    }
  }, [tab, hasExecutions, isExecsLoading, setTab]);

  const buttonText =
    extension?.extension_type === 'install'
      ? t('extensions.install') || 'Install'
      : t('extensions.run') || 'Run';

  const [openRunIndex, setOpenRunIndex] = useState<number | null>(null);
  const [openValidateIndex, setOpenValidateIndex] = useState<number | null>(null);
  const parsed = useMemo(() => {
    try {
      if (!extension?.yaml_content) return undefined;
      const y = YAML.parse(extension.yaml_content || '');
      return {
        execution: y?.execution || {}
      } as any;
    } catch {
      return undefined;
    }
  }, [extension?.yaml_content]);

  const variableColumns: TableColumn<NonNullable<Extension['variables']>[0]>[] = useMemo(
    () => [
      {
        key: 'name',
        title: 'Name',
        dataIndex: 'variable_name',
        width: '25%'
      },
      {
        key: 'type',
        title: 'Type',
        dataIndex: 'variable_type',
        width: '17%'
      },
      {
        key: 'required',
        title: 'Required',
        render: (_, record) => (record.is_required ? 'Yes' : 'No'),
        width: '17%'
      },
      {
        key: 'default',
        title: 'Default',
        render: (_, record) => String(record.default_value ?? ''),
        width: '17%',
        className: 'truncate'
      },
      {
        key: 'description',
        title: 'Description',
        dataIndex: 'description',
        width: '24%'
      }
    ],
    []
  );

  const entryColumns: TableColumn<[string, any]>[] = useMemo(
    () => [
      {
        key: 'key',
        title: 'Key',
        render: (_: any, record: [string, any]) => record[0],
        width: '25%',
        className: 'text-muted-foreground'
      },
      {
        key: 'value',
        title: 'Value',
        render: (_: any, record: [string, any]) => {
          const v = record[1];
          return typeof v === 'object' ? (
            <pre className="whitespace-pre-wrap text-xs">{JSON.stringify(v, null, 2)}</pre>
          ) : (
            String(v)
          );
        },
        width: '75%',
        className: 'break-words'
      }
    ],
    []
  );

  const forkPreview = useMemo(() => {
    try {
      const y = YAML.parse(forkYaml || '');
      const variables = y?.variables || {};
      const variablesArray: VariableData[] = Object.entries(variables).map(
        ([key, val]: [string, any]) => ({
          name: key,
          type: val?.variable_type || val?.type || '',
          required: val?.is_required ? 'Yes' : 'No',
          default: String(val?.default_value ?? ''),
          description: val?.description || ''
        })
      );
      return {
        variables: variablesArray,
        execution: y?.execution || {},
        metadata: y?.metadata || {}
      } as any;
    } catch {
      return undefined;
    }
  }, [forkYaml]);

  const forkVariableColumns: TableColumn<VariableData>[] = useMemo(
    () => [
      { key: 'name', title: 'Name', dataIndex: 'name' },
      { key: 'type', title: 'Type', dataIndex: 'type' },
      { key: 'required', title: 'Required', dataIndex: 'required' },
      {
        key: 'default',
        title: 'Default',
        dataIndex: 'default',
        className: 'truncate max-w-[120px]'
      },
      { key: 'description', title: 'Description', dataIndex: 'description' }
    ],
    []
  );

  useEffect(() => {
    if (forkOpen) {
      setForkYaml(extension?.yaml_content || '');
    }
  }, [forkOpen, extension]);

  const doFork = async () => {
    try {
      await forkExtension({
        extensionId: extension?.extension_id || '',
        yaml_content: forkYaml || undefined
      }).unwrap();
      toast.success(t('extensions.forkSuccess') || 'Extension forked successfully');
      setForkOpen(false);
      router.push('/extensions');
    } catch (e) {
      toast.error(t('extensions.forkFailed') || 'Failed to fork extension');
    }
  };

  useEffect(() => {
    if (prevExtensionIdRef.current !== id) {
      setSelectedExecId(null);
      setAfterSeq(0);
      setAllLogs([]);
      setExecutionStatus(null);
      setOpen(false);
      prevExecIdRef.current = null;
      prevExtensionIdRef.current = id;
      return;
    }
  }, [id]);

  useEffect(() => {
    if (!selectedExecId) {
      setAfterSeq(0);
      setAllLogs([]);
      setExecutionStatus(null);
      prevExecIdRef.current = null;
      return;
    }

    const execIdChanged = prevExecIdRef.current !== selectedExecId;
    if (execIdChanged) {
      setAfterSeq(0);
      setAllLogs([]);
      setExecutionStatus(null);
      prevExecIdRef.current = selectedExecId;
      if (open) {
        refetch();
      }
    }
  }, [selectedExecId, open, refetch]);

  useEffect(() => {
    if (
      open &&
      selectedExecId &&
      prevExecIdRef.current === selectedExecId &&
      afterSeq === 0 &&
      allLogs.length === 0
    ) {
      refetch();
    }
  }, [open, selectedExecId, afterSeq, allLogs.length, refetch]);

  useEffect(() => {
    if (!logsResp || !selectedExecId) return;

    if (logsResp.execution_status) {
      setExecutionStatus(logsResp.execution_status);
    }

    if (logsResp.logs && logsResp.logs.length > 0) {
      setAfterSeq(logsResp.next_after);
      setAllLogs((prev) => {
        const existingIds = new Set(prev.map((l) => l.id));
        const newLogs = logsResp.logs.filter((l) => !existingIds.has(l.id));
        const merged = [...prev, ...newLogs];
        return merged.sort((a, b) => {
          const timeA = new Date(a.created_at).getTime();
          const timeB = new Date(b.created_at).getTime();
          if (timeA !== timeB) return timeA - timeB;
          return a.sequence - b.sequence;
        });
      });
    }
  }, [logsResp, selectedExecId]);

  useEffect(() => {
    const exec = search?.get('exec');
    const openLogs = search?.get('openLogs') === '1';
    if (exec && openLogs) {
      setSelectedExecId(exec);
      setOpen(true);
    }
  }, [search]);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [allLogs]);

  const formattedLogs = useMemo(() => {
    const stepCompletedIds = new Set<string>(
      allLogs
        .filter((log) => log.message?.startsWith('step_completed'))
        .map((log) => log.step_id)
        .filter((id): id is string => Boolean(id))
    );

    const stepFailedIds = new Set<string>(
      allLogs
        .filter((log) => log.message?.startsWith('step_failed'))
        .map((log) => log.step_id)
        .filter((id): id is string => Boolean(id))
    );

    return allLogs.map((log): FormattedLog => {
      const timestamp = new Date(log.created_at).toLocaleTimeString();
      const level = log.level.toUpperCase();
      let isVerbose = false;
      let progressInfo: FormattedLog['progressInfo'] | undefined;

      const logState = getLogState(log, stepCompletedIds, stepFailedIds);
      const color = getLogColor(logState);
      let icon = getLogIcon(logState);

      if (log.message.startsWith('step_started') && logState === 'info') {
        icon = <Loader2 className="h-4 w-4 animate-spin" />;
      }

      if (log.data) {
        const dataStr = typeof log.data === 'string' ? log.data : JSON.stringify(log.data);
        const isLargeData = dataStr.length > 5000;
        const isDockerProgress =
          typeof log.data === 'string' &&
          (log.data.includes('{"status":"Downloading"') ||
            log.data.includes('"status":"Pulling') ||
            log.data.includes('"status":"Extracting') ||
            log.data.includes('"status":"Verifying'));

        if (isLargeData || isDockerProgress) {
          isVerbose = true;
          if (isDockerProgress) {
            const lines = dataStr.split(/[\r\n]+/).filter((l) => l.trim());
            let lastProgress: string | undefined;
            let lastStatus: string | undefined;
            for (let i = lines.length - 1; i >= 0; i--) {
              const line = lines[i].trim();
              if (!line) continue;
              try {
                const parsed = JSON.parse(line);
                if (
                  parsed.status &&
                  (parsed.status.includes('Downloading') ||
                    parsed.status.includes('Extracting') ||
                    parsed.status.includes('Pulling'))
                ) {
                  if (parsed.progress) lastProgress = parsed.progress;
                  if (parsed.status) lastStatus = parsed.status;
                  if (lastProgress && lastStatus) break;
                }
              } catch {
                continue;
              }
            }
            if (lastProgress || lastStatus) {
              progressInfo = {
                progress: lastProgress || '',
                status: lastStatus || 'Processing'
              };
            }
          }
        }
      }

      return {
        id: log.id,
        timestamp,
        level,
        message: log.message,
        icon,
        color,
        data: log.data,
        isVerbose,
        progressInfo
      };
    });
  }, [allLogs]);

  useEffect(() => {
    setCollapsedLogs((prev) => {
      const next = new Set(prev);
      formattedLogs.forEach((log) => {
        if (log.isVerbose && log.data != null && !next.has(log.id)) {
          next.add(log.id);
        }
      });
      return next;
    });
  }, [formattedLogs]);

  const onOpenLogs = (execId: string) => {
    setSelectedExecId(execId);
    setOpen(true);
  };

  const toggleCollapse = (logId: string) => {
    setCollapsedLogs((prev) => {
      const next = new Set(prev);
      if (next.has(logId)) {
        next.delete(logId);
      } else {
        next.add(logId);
      }
      return next;
    });
  };

  const executionColumns: TableColumn<ExtensionExecution>[] = useMemo(
    () => [
      {
        key: 'id',
        title: t('extensions.executionId') || 'Execution ID',
        render: (_, record) => (
          <TypographySmall className="font-mono truncate">{record.id}</TypographySmall>
        ),
        width: '33%'
      },
      {
        key: 'status',
        title: t('extensions.status') || 'Status',
        render: (_, record) => <StatusBadge status={record.status} />,
        width: '17%'
      },
      {
        key: 'started_at',
        title: t('extensions.startedAt') || 'Started At',
        render: (_, record) => (
          <TypographyMuted>{new Date(record.started_at).toLocaleString()}</TypographyMuted>
        ),
        width: '25%'
      },
      {
        key: 'completed_at',
        title: t('extensions.completedAt') || 'Completed At',
        render: (_, record) => (
          <TypographyMuted>
            {record.completed_at ? new Date(record.completed_at).toLocaleString() : '-'}
          </TypographyMuted>
        ),
        width: '25%'
      }
    ],
    [t]
  );

  const tabs = useMemo(
    () => [
      {
        value: 'overview',
        label: t('extensions.overview') || 'Overview',
        icon: Info,
        content: (
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
        )
      },
      {
        value: 'executions',
        label: t('extensions.executions') || 'Executions',
        icon: Terminal,
        content: (
          <LogsTab
            executions={executions || []}
            executionColumns={executionColumns}
            isLoading={isExecsLoading}
            open={open}
            setOpen={setOpen}
            selectedExecId={selectedExecId || ''}
            onOpenLogs={onOpenLogs}
            formattedLogs={formattedLogs}
            collapsedLogs={collapsedLogs}
            toggleCollapse={toggleCollapse}
            logsEndRef={logsEndRef as React.RefObject<HTMLDivElement>}
          />
        )
      }
    ],
    [
      t,
      extension,
      isLoading,
      parsed,
      variableColumns,
      entryColumns,
      openRunIndex,
      openValidateIndex,
      executions,
      executionColumns,
      isExecsLoading,
      open,
      selectedExecId,
      onOpenLogs,
      formattedLogs,
      collapsedLogs,
      toggleCollapse,
      logsEndRef
    ]
  );

  useEffect(() => {
    const exec = search?.get('exec');
    const openLogs = search?.get('openLogs') === '1';
    if (exec && openLogs) {
      setTab('executions');
      initializedDefaultTab.current = true;
    }
  }, [search]);

  useEffect(() => {
    if (initializedDefaultTab.current) return;

    if (isExecsLoading) return;

    if (!executions) return;

    setTab(executions.length > 0 ? 'executions' : 'overview');
    initializedDefaultTab.current = true;
  }, [executions, isExecsLoading]);

  return {
    runModalOpen,
    runExtension,
    isRunning,
    isLoading,
    tab,
    extension,
    router,
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
    formattedLogs,
    collapsedLogs,
    toggleCollapse,
    logsEndRef,
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
  };
}

export default useExtensionDetails;

export function extractDockerProgress(dataStr: string): FormattedLog['progressInfo'] | undefined {
  try {
    const lines = dataStr.split(/[\r\n]+/).filter((l) => l.trim());
    let lastProgress: string | undefined;
    let lastStatus: string | undefined;

    for (let i = lines.length - 1; i >= 0; i--) {
      const line = lines[i].trim();
      if (!line) continue;

      try {
        const parsed = JSON.parse(line);
        if (
          parsed.status &&
          (parsed.status.includes('Downloading') ||
            parsed.status.includes('Extracting') ||
            parsed.status.includes('Pulling'))
        ) {
          if (parsed.progress) {
            lastProgress = parsed.progress;
          }
          if (parsed.status) {
            lastStatus = parsed.status;
          }
          if (lastProgress && lastStatus) break;
        }
      } catch {
        continue;
      }
    }

    if (lastProgress || lastStatus) {
      return {
        progress: lastProgress || '',
        status: lastStatus || 'Processing'
      };
    }
  } catch {}

  return undefined;
}

export function formatLogMessage(message: string, data?: unknown): string {
  if (!data) {
    if (message === 'execution_started') return 'Execution started';

    if (message === 'execution_completed') return 'Execution completed';

    return message;
  }

  try {
    const parsed = typeof data === 'string' ? JSON.parse(data) : data;

    if (message === 'step_started' && parsed.step_name) {
      const phase = parsed.phase ? ` (${parsed.phase})` : '';
      const order = parsed.order ? ` #${parsed.order}` : '';
      return `Starting: ${parsed.step_name}${phase}${order}`;
    }

    if (message === 'step_completed' && parsed.step_name) {
      return `Completed: ${parsed.step_name}`;
    }

    if (message === 'step_failed' && parsed.step_name) {
      return `Failed: ${parsed.step_name}`;
    }

    if (message.includes('Check') && parsed.output) {
      const output = typeof parsed.output === 'string' ? parsed.output.trim() : '';
      if (output) {
        return `${message}: ${output.split('\n')[0].substring(0, 80)}`;
      }
    }

    return message;
  } catch {
    return message;
  }
}

export function formatDataPreview(data: unknown): string {
  if (!data) return '';

  const dataStr = typeof data === 'string' ? data : JSON.stringify(data, null, 2);

  if (dataStr.length > 200) {
    return dataStr.substring(0, 200) + '...';
  }

  return dataStr;
}

export function formatVerboseData(data: unknown): string {
  if (!data) return '';

  return typeof data === 'string' ? data : JSON.stringify(data ?? null, null, 2);
}

export function useFormatLog(
  log: ExtensionLog,
  stepCompletedIds: Set<string>,
  stepFailedIds: Set<string>
): FormattedLog {
  const timestamp = new Date(log.created_at).toLocaleTimeString();
  const level = log.level.toUpperCase();

  let isVerbose = false;
  let progressInfo: FormattedLog['progressInfo'] | undefined;

  const logState = getLogState(log, stepCompletedIds, stepFailedIds);
  const color = getLogColor(logState);
  let icon = getLogIcon(logState);

  if (log.message.startsWith('step_started') && logState === 'info') {
    icon = <Loader2 className="h-4 w-4 animate-spin" />;
  }

  if (log.data) {
    const dataStr = typeof log.data === 'string' ? log.data : JSON.stringify(log.data);
    const isLargeData = dataStr.length > 5000;
    const isDockerProgress =
      typeof log.data === 'string' &&
      (log.data.includes('{"status":"Downloading"') ||
        log.data.includes('"status":"Pulling') ||
        log.data.includes('"status":"Extracting') ||
        log.data.includes('"status":"Verifying'));

    if (isLargeData || isDockerProgress) {
      isVerbose = true;

      if (isDockerProgress) {
        progressInfo = extractDockerProgress(dataStr);
      }
    }
  }

  return {
    id: log.id,
    timestamp,
    level,
    message: log.message,
    icon,
    color,
    data: log.data,
    isVerbose,
    progressInfo
  };
}

function getStatusBadgeClass(status: string): string {
  const s = (status || '').toLowerCase();
  if (s === 'completed') {
    return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400';
  }
  if (s === 'failed') {
    return 'bg-destructive/10 text-destructive';
  }
  return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400';
}

export function StatusBadge({ status }: { status: string }) {
  return <Badge className={getStatusBadgeClass(status)}>{status}</Badge>;
}
