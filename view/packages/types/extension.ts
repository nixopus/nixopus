import { Extension, ExtensionExecution, ExtensionVariable } from '@/redux/types/extension';
import { TableColumn } from '@/components/ui/data-table';
import { translationKey } from '@/packages/hooks/shared/use-translation';
import { DialogAction } from '@/components/ui/dialog-wrapper';

export type CategoryBadgesProps = {
  categories: string[];
  selected?: string | null;
  onChange?: (value: string | null) => void;
  className?: string;
  showAll?: boolean;
};

export interface ExtensionsGridProps {
  extensions?: Extension[];
  isLoading?: boolean;
  error?: string;
  onInstall?: (extension: Extension) => void;
  onViewDetails?: (extension: Extension) => void;
  onForkClick?: (extension: Extension) => void;
  setConfirmOpen: (open: boolean) => void;
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
  forkOpen: boolean;
  setForkOpen: (open: boolean) => void;
  confirmOpen: boolean;
  forkYaml: string;
  setForkYaml: (yaml: string) => void;
  preview: any;
  variableColumns: TableColumn<VariableData>[];
  doFork: () => void;
  selectedExtension?: Extension | null;
}

export interface ExtensionForkDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension: Extension;
  t: (key: translationKey) => string;
  forkYaml: string;
  setForkYaml: (yaml: string) => void;
  preview: any;
  variableColumns: TableColumn<VariableData>[];
  doFork: () => void;
  isLoading: boolean;
}

export interface VariableData {
  name: string;
  type: string;
  required: string;
  default: string;
  description: string;
}

export interface ExtensionCardProps {
  extension: Extension;
  onInstall?: (extension: Extension) => void;
  onViewDetails?: (extension: Extension) => void;
  onFork?: (extension: Extension) => void;
  onForkClick?: (extension: Extension) => void;
  onRemove?: (extension: Extension) => void;
  setConfirmOpen: (open: boolean) => void;
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
  t: (key: translationKey) => string;
  confirmOpen: boolean;
}

export interface ExtensionInputProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension?: Extension | null;
  onSubmit?: (values: Record<string, unknown>) => void;
  t: (key: translationKey) => string;
  actions: DialogAction[];
  isOnlyProxyDomain: boolean;
  noFieldsToShow: boolean;
  values: Record<string, unknown>;
  errors: Record<string, string>;
  handleChange: (name: string, value: unknown) => void;
  handleSubmit: () => void;
  requiredFields: ExtensionVariable[];
}

export interface FormattedLog {
  id: string;
  timestamp: string;
  level: string;
  message: string;
  icon?: React.ReactNode;
  color: string;
  data?: unknown;
  isVerbose?: boolean;
  progressInfo?: {
    progress?: string;
    status?: string;
    id?: string;
  };
}

export interface LogsTabProps {
  executions: ExtensionExecution[];
  executionColumns: TableColumn<ExtensionExecution>[];
  isLoading: boolean;
  open: boolean;
  setOpen: (open: boolean) => void;
  selectedExecId: string;
  onOpenLogs: (execId: string) => void;
  formattedLogs: FormattedLog[];
  collapsedLogs: Set<string>;
  toggleCollapse: (logId: string) => void;
  logsEndRef: React.RefObject<HTMLDivElement>;
}

export interface OverviewTabProps {
  extension?: Extension;
  isLoading?: boolean;
  parsed?: any;
  variableColumns?: TableColumn<NonNullable<Extension['variables']>[0]>[];
  entryColumns?: TableColumn<[string, any]>[];
  openRunIndex?: number | null;
  openValidateIndex?: number | null;
  onToggleRun?: (index: number | null) => void;
  onToggleValidate?: (index: number | null) => void;
}
