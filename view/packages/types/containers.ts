import React from 'react';
import { Container } from '@/redux/services/container/containerApi';
import { ContainerData } from '@/redux/types/monitor';
import { TableColumn } from '@/components/ui/data-table';
import { translationKey } from '@/packages/hooks/shared/use-translation';
import { ControllerRenderProps, UseFormReturn } from 'react-hook-form';
import {
  PresetType,
  FieldConfig,
  ResourceLimitsFormValues
} from '@/packages/hooks/containers/use-update-container-resources';
import { ParsedLogEntry } from '@/packages/hooks/containers/use-container-logs';

export type ViewMode = 'table' | 'card';

export const CONTAINERS_VIEW_STORAGE_KEY = 'containers_view';

export enum Action {
  START = 'start',
  STOP = 'stop',
  REMOVE = 'remove'
}

export type SortField = 'name' | 'status';

export type ContainerAction = 'start' | 'stop' | 'restart' | 'remove';

export interface ContainerActionsProps {
  container: any;
  onAction: (id: string, action: Action) => void;
}

export interface ActionButtonProps {
  icon: React.ElementType;
  onClick: (e: React.MouseEvent) => void;
  disabled?: boolean;
  tooltip?: string;
  variant?: 'success' | 'warning' | 'danger';
}

export interface ActionHeaderProps {
  handleRefresh: () => Promise<void>;
  isRefreshing: boolean;
  isFetching: boolean;
  t: (key: translationKey, params?: Record<string, string>) => string;
  setShowPruneImagesConfirm: React.Dispatch<React.SetStateAction<boolean>>;
  setShowPruneBuildCacheConfirm: React.Dispatch<React.SetStateAction<boolean>>;
}

export interface ContainerCardProps {
  container: Container;
  onClick: () => void;
  onAction: (id: string, action: Action) => void;
}

export interface ContainersTableProps {
  containersData: Container[];
  sortBy?: SortField;
  sortOrder?: 'asc' | 'desc';
  onSort?: (field: SortField) => void;
  onAction?: (id: string, action: Action) => void;
}

export interface SortableHeaderProps {
  label: string;
  field: SortField;
  currentSort: SortField;
  currentOrder: 'asc' | 'desc';
  onSort?: (field: SortField) => void;
}

export interface ContainerRowProps {
  container: Container;
  onClick: () => void;
  onAction?: (id: string, action: Action) => void;
}

export interface ContainersWidgetProps {
  containersData: ContainerData[];
  columns: TableColumn<ContainerData>[];
}

export interface ContainerDetailsHeaderProps {
  container: Container;
  isLoading: boolean;
  isProtected: boolean;
  handleContainerAction: (action: ContainerAction) => void;
  t: (key: translationKey, params?: Record<string, string>) => string;
}

export interface OverviewTabProps {
  container: Container;
}

export interface StatBlockProps {
  value: string | number;
  label: string;
  sublabel?: string;
  color?: 'emerald' | 'red' | 'amber' | 'blue';
  pulse?: boolean;
}

export interface ResourceGaugeProps {
  icon: React.ElementType;
  label: string;
  value: number;
  maxLabel: string;
  color: 'blue' | 'purple' | 'amber' | 'emerald';
  unlimited?: boolean;
  showBar?: boolean;
}

export interface PortFlowProps {
  port: {
    private_port: number;
    public_port: number;
    type: string;
  };
}

export interface InfoLineProps {
  icon: React.ElementType;
  label: string;
  value: string;
  displayValue?: string;
  sublabel?: string;
  mono?: boolean;
  copyable?: boolean;
  onCopy?: () => void;
  copied?: boolean;
}

export interface LogsTabProps {
  container: Container;
  logs: string;
  onLoadMore: () => void;
  onRefresh?: () => void;
}

export interface LogEntryProps {
  log: ParsedLogEntry;
  isExpanded: boolean;
  onToggle: () => void;
  isDense: boolean;
}

export interface ContainerImage {
  id: string;
  repo_tags: string[];
  repo_digests: string[];
  created: number;
  size: number;
  shared_size: number;
  virtual_size: number;
  labels: Record<string, string>;
}

export interface ImagesProps {
  containerId: string;
  imagePrefix: string;
}

export interface StatItemProps {
  icon: React.ElementType;
  value: string | number;
  label: string;
}

export interface ImageCardProps {
  image: ContainerImage;
  isFirst: boolean;
}

export interface DetailRowProps {
  icon: React.ElementType;
  label: string;
  value: string;
  displayValue?: string;
  sublabel?: string;
  mono?: boolean;
  copyable?: boolean;
  onCopy?: () => void;
  copied?: boolean;
}

export interface ResourceLimitsFormProps {
  container: Container;
}

export interface PresetButtonProps {
  presetKey: PresetType;
  memory: number;
  isActive: boolean;
  onSelect: (key: PresetType) => void;
}

export interface PresetGridProps {
  currentMemory: number;
  onPresetSelect: (key: PresetType) => void;
}

export interface ResourceFieldProps {
  config: FieldConfig;
  field: ControllerRenderProps<ResourceLimitsFormValues, FieldConfig['name']>;
}

export interface FormActionsProps {
  isLoading: boolean;
  isDirty: boolean;
  onReset: () => void;
  onCancel: () => void;
}

export interface ResourceFieldsProps {
  form: UseFormReturn<ResourceLimitsFormValues>;
}

export interface StatPillProps {
  value: number;
  label: string;
  color?: 'emerald' | 'zinc';
}

export interface TerminalProps {
  containerId: string;
}

export interface StatusIndicatorProps {
  isRunning: boolean;
  size?: 'sm' | 'md' | 'lg';
  showPulse?: boolean;
}

export interface CopyButtonProps {
  copied: boolean;
  onCopy: () => void;
  size?: 'sm' | 'md';
  className?: string;
  showText?: boolean;
}

export interface PortDisplayProps {
  port: {
    private_port: number;
    public_port: number;
    type: string;
  };
  variant?: 'pill' | 'flow' | 'inline';
  showType?: boolean;
}

export interface StatusBadgeProps {
  status: string;
  showDot?: boolean;
  className?: string;
}

export interface EmptyStateProps {
  icon: React.ElementType;
  message: string;
  className?: string;
}

export interface GroupedContainerViewProps {
  groups: Array<{ application_id: string; application_name: string; containers: Container[] }>;
  ungrouped?: Container[];
  viewMode: 'table' | 'card';
  onContainerClick: (container: Container) => void;
  onContainerAction: (id: string, action: 'start' | 'stop' | 'remove') => void;
  sortBy: 'name' | 'status';
  sortOrder: 'asc' | 'desc';
  onSort: (field: 'name' | 'status') => void;
}
