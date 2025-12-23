import type { Node, Edge } from '@xyflow/react';
import type { LucideIcon } from 'lucide-react';
import type { UseFormReturn } from 'react-hook-form';
import type { z } from 'zod';
import type { Server as ServerType, ServerRole, CreateServerRequest } from '@/redux/types/servers';

export type FieldType = 'text' | 'number' | 'password' | 'textarea';
export type AuthMethod = 'key' | 'password';

export interface FormFieldConfig {
  name: string;
  label: string;
  placeholder: string;
  type: FieldType;
  icon?: LucideIcon;
  description?: string;
  colSpan?: number;
  className?: string;
}

export interface AuthTabConfig {
  value: AuthMethod;
  label: string;
  icon: LucideIcon;
  field: FormFieldConfig;
}

export interface MenuItemConfig {
  key: string;
  icon: LucideIcon;
  label: string;
  onClick?: () => void;
  disabled?: boolean;
  destructive?: boolean;
  separator?: boolean;
}

export interface ActionButtonConfig {
  key: string;
  icon: LucideIcon;
  label: string;
  onClick?: () => void;
  disabled?: boolean;
  destructive?: boolean;
  variant?: 'default' | 'outline' | 'destructive';
}

export interface MetricConfig {
  key: string;
  icon: LucideIcon;
  label: string;
  getValue: (metrics: ServerType['metrics']) => number;
}

export interface MetricRowConfig {
  key: string;
  icon: LucideIcon;
  label: string;
  value: number;
}

export interface MetricBarConfig {
  key: string;
  icon: LucideIcon;
  label: string;
  value: number;
}

export interface ServerNodeData extends Record<string, unknown> {
  server: ServerType;
  isSelected: boolean;
  onSelect: (server: ServerType) => void;
}

export type ServerNodeType = Node<ServerNodeData, 'serverNode'>;

export interface ClusterEdgeData extends Record<string, unknown> {
  isActive?: boolean;
  label?: string;
}

export type ClusterEdgeType = Edge<ClusterEdgeData, 'clusterEdge'>;

export interface AddServerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateServerRequest) => Promise<void>;
}

export interface ServerFormFieldProps<T extends Record<string, unknown>> {
  config: FormFieldConfig;
  form: UseFormReturn<T>;
}

export interface BasicFieldsSectionProps<T extends Record<string, unknown>> {
  fields: FormFieldConfig[];
  form: UseFormReturn<T>;
}

export interface AuthMethodTabsProps<T extends Record<string, unknown>> {
  authMethod: AuthMethod;
  setAuthMethod: (method: AuthMethod) => void;
  authTabs: AuthTabConfig[];
  authMethodLabel: string;
  form: UseFormReturn<T>;
}

export interface AdvancedFieldsSectionProps<T extends Record<string, unknown>> {
  fields: FormFieldConfig[];
  form: UseFormReturn<T>;
}

export interface ServerCardProps {
  server: ServerType;
  isSelected: boolean;
  onSelect: (server: ServerType) => void;
  onEdit: (server: ServerType) => void;
  onDelete: (server: ServerType) => void;
  onTestConnection: (serverId: string) => void;
  isTestingConnection: boolean;
}

export interface CardStatusIndicatorProps {
  status: ServerType['status'];
}

export interface CardServerIconProps {
  status: ServerType['status'];
  health: ServerType['health'];
  role: ServerType['role'];
}

export interface ServerLabelsProps {
  visibleLabels: string[];
  remainingCount: number;
  role: ServerType['role'];
  status: ServerType['status'];
  statusText: string;
}

export interface MetricsGridProps {
  metrics: NonNullable<ServerType['metrics']>;
  getMetricColor: (value: number) => string;
}

export interface CardStatusMessageProps {
  status: ServerType['status'];
  statusText: string;
}

export interface CardActionsProps {
  serverId: string;
  menuItems: MenuItemConfig[];
  onTestConnection: (serverId: string) => void;
  isTestingConnection: boolean;
  testConnectionLabel: string;
}

export interface ServerDetailsPanelProps {
  server: ServerType;
  onClose: () => void;
  onEdit: (server: ServerType) => void;
  onDelete: (server: ServerType) => void;
  onTestConnection: (serverId: string) => void;
  onJoinCluster: (serverId: string, role: ServerRole) => void;
  onLeaveCluster: (serverId: string) => void;
  onPromote: (serverId: string) => void;
  onDemote: (serverId: string) => void;
  isTestingConnection: boolean;
  hasCluster: boolean;
}

export interface PanelHeaderProps {
  server: ServerType;
  onClose: () => void;
}

export interface ServerBadgesProps {
  server: ServerType;
  statusText: string;
  healthText: string;
}

export interface MetricRowProps {
  config: MetricRowConfig;
  getBarColor: (value: number) => string;
  getTextColor: (value: number) => string;
}

export interface PanelMetricsSectionProps {
  metrics: NonNullable<ServerType['metrics']>;
  metricRows: MetricRowConfig[];
  getBarColor: (value: number) => string;
  getTextColor: (value: number) => string;
  containerLabel: string;
  lastCheckedLabel: string;
}

export interface ClusterActionsSectionProps {
  actions: ActionButtonConfig[];
  showNoClusterMessage: boolean;
  noClusterMessage: string;
  title: string;
}

export interface QuickActionsSectionProps {
  actions: ActionButtonConfig[];
  title: string;
}

export interface DeleteFooterProps {
  onDelete: () => void;
  label: string;
}

export interface NodeStatusIndicatorProps {
  status: ServerType['status'];
}

export interface NodeServerIconProps {
  status: ServerType['status'];
  health: ServerType['health'];
  role: ServerType['role'];
}

export interface NodeLabelsProps {
  role: ServerType['role'];
  roleLabel: string;
  visibleLabels: string[];
  remainingCount: number;
}

export interface MetricBarProps {
  config: MetricBarConfig;
  getColor: (value: number) => string;
}

export interface NodeMetricsSectionProps {
  metricBars: MetricBarConfig[];
  containerCountText: string;
  getMetricBarColor: (value: number) => string;
}

export interface NodeStatusMessageProps {
  status: ServerType['status'];
  offlineText: string;
  connectingText: string;
}

export interface ServerTopologyProps {
  servers: ServerType[];
  selectedServer: ServerType | null;
  onServerSelect: (server: ServerType | null) => void;
}

export interface UseAddServerDialogProps {
  onSubmit: (data: CreateServerRequest) => Promise<void>;
  onOpenChange: (open: boolean) => void;
}

export interface UseServerCardProps {
  server: ServerType;
  onEdit: (server: ServerType) => void;
  onDelete: (server: ServerType) => void;
}

export interface UseServerDetailsPanelProps {
  server: ServerType;
  onEdit: (server: ServerType) => void;
  onTestConnection: (serverId: string) => void;
  onJoinCluster: (serverId: string, role: ServerRole) => void;
  onLeaveCluster: (serverId: string) => void;
  onPromote: (serverId: string) => void;
  onDemote: (serverId: string) => void;
  isTestingConnection: boolean;
  hasCluster: boolean;
}

export interface UseServerNodeProps {
  server: ServerType;
}

export interface UseServerTopologyProps {
  servers: ServerType[];
  selectedServer: ServerType | null;
  onServerSelect: (server: ServerType | null) => void;
}
