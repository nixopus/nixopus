import { Extension, ExtensionVariable } from '@/redux/types/extension';
import { TableColumn } from '@/components/ui/data-table';
import { translationKey } from '@/hooks/use-translation';
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
  setForkOpen: (open: boolean) => void;
  setConfirmOpen: (open: boolean) => void;
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
  forkOpen: boolean;
  confirmOpen: boolean;
  forkYaml: string;
  setForkYaml: (yaml: string) => void;
  preview: any;
  variableColumns: TableColumn<VariableData>[];
  doFork: () => void;
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
  onRemove?: (extension: Extension) => void;
  setForkOpen: (open: boolean) => void;
  setConfirmOpen: (open: boolean) => void;
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
  t: (key: translationKey) => string;
  forkOpen: boolean;
  confirmOpen: boolean;
  forkYaml: string;
  setForkYaml: (yaml: string) => void;
  preview: any;
  variableColumns: TableColumn<VariableData>[];
  doFork: () => void;
  isLoading: boolean;
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
