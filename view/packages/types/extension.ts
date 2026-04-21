import { Extension, ExtensionVariable } from '@/redux/types/extension';
import { translationKey } from '@/packages/hooks/shared/use-translation';
import { DialogAction } from '@nixopus/ui';
import { useServerSelector } from '@/packages/hooks/deploy/use-server-selector';

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
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
}

export interface ExtensionCardProps {
  extension: Extension;
  onInstall?: (extension: Extension) => void;
  onViewDetails?: (extension: Extension) => void;
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
  t: (key: translationKey) => string;
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
  serverSelector?: ReturnType<typeof useServerSelector>;
}

export interface OverviewTabProps {
  extension?: Extension;
  isLoading?: boolean;
  variableColumns?: import('@nixopus/ui').TableColumn<NonNullable<Extension['variables']>[0]>[];
}
