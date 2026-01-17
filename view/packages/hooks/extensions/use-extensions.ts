'use client';

import { useEffect, useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Extension,
  ExtensionListParams,
  SortDirection,
  ExtensionSortField,
  ExtensionCategory
} from '@/redux/types/extension';
import {
  useGetExtensionsQuery,
  useRunExtensionMutation,
  useCancelExecutionMutation,
  useGetExtensionCategoriesQuery,
  useDeleteExtensionMutation,
  useForkExtensionMutation
} from '@/redux/services/extensions/extensionsApi';
import { SelectOption } from '@/components/ui/select-wrapper';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';
import YAML from 'yaml';
import { TableColumn } from '@/components/ui/data-table';
import { VariableData } from '@/packages/types/extension';
import { useExtensionInput } from './use-extension-input';
import { DialogAction } from '@/components/ui/dialog-wrapper';

export function useExtensions() {
  const router = useRouter();
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<{
    key: ExtensionSortField;
    direction: SortDirection;
  }>({
    key: 'name',
    direction: 'asc'
  });
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(9);
  const [runModalOpen, setRunModalOpen] = useState(false);
  const [selectedExtension, setSelectedExtension] = useState<Extension | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<ExtensionCategory | null>(null);
  const { t } = useTranslation();
  const [forkOpen, setForkOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleteExtension] = useDeleteExtensionMutation();
  const [expanded, setExpanded] = useState(false);

  const onDelete = async (extension: Extension) => {
    try {
      await deleteExtension({ id: extension.id }).unwrap();
      toast.success(t('extensions.deleteSuccess') || 'Removed');
    } catch (e) {
      toast.error(t('extensions.deleteFailed') || 'Remove failed');
    }
  };

  const queryParams: ExtensionListParams = {
    search: searchTerm || undefined,
    category: selectedCategory || undefined,
    sort_by: sortConfig.key,
    sort_dir: sortConfig.direction,
    page: currentPage,
    page_size: itemsPerPage
  };

  const { data: response, isLoading, error: apiError } = useGetExtensionsQuery(queryParams);

  const { data: categories = [] } = useGetExtensionCategoriesQuery();

  const extensions = response?.extensions || [];
  const totalPages = response?.total_pages || 0;
  const totalExtensions = response?.total || 0;

  const handleSearchChange = (value: string) => {
    setSearchTerm(value);
    setCurrentPage(1); // Reset to first page when searching
  };

  const handleSortChange = (key: ExtensionSortField, direction: SortDirection) => {
    setSortConfig({ key, direction });
    setCurrentPage(1); // Reset to first page when sorting
  };

  const handleCategoryChange = (value: string | null) => {
    setSelectedCategory((value as ExtensionCategory) || null);
    setCurrentPage(1);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleInstall = (extension: Extension) => {
    setSelectedExtension(extension);
    setRunModalOpen(true);
  };

  const handleViewDetails = (extension: Extension) => {
    router.push(`/extensions/${extension.id}`);
  };

  const handleForkClick = (extension: Extension) => {
    setSelectedExtension(extension);
    setForkOpen(true);
  };

  const error = apiError ? 'Failed to load extensions' : null;

  const [runExtensionMutation] = useRunExtensionMutation();
  const [cancelExecutionMutation] = useCancelExecutionMutation();

  const handleRun = async (values: Record<string, unknown>) => {
    if (!selectedExtension) return;
    const exec = await runExtensionMutation({
      extensionId: selectedExtension.extension_id,
      body: { variables: values }
    }).unwrap();
    setRunModalOpen(false);
    router.push(`/extensions/${selectedExtension.id}?exec=${exec.id}&openLogs=1`);
  };

  const handleCancel = async (executionId: string) => {
    await cancelExecutionMutation({ executionId });
  };

  const sortOptions: SelectOption[] = [
    { value: 'name_asc', label: t('extensions.sortOptions.name') + ' (A-Z)' },
    { value: 'name_desc', label: t('extensions.sortOptions.name') + ' (Z-A)' }
  ];

  const [forkYaml, setForkYaml] = useState<string>('');
  const [forkExtension] = useForkExtensionMutation();

  const preview = useMemo(() => {
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

  const variableColumns: TableColumn<VariableData>[] = [
    { key: 'name', title: 'Name', dataIndex: 'name' },
    { key: 'type', title: 'Type', dataIndex: 'type' },
    { key: 'required', title: 'Required', dataIndex: 'required' },
    { key: 'default', title: 'Default', dataIndex: 'default', className: 'truncate max-w-[120px]' },
    { key: 'description', title: 'Description', dataIndex: 'description' }
  ];

  useEffect(() => {
    if (forkOpen && selectedExtension) {
      setForkYaml(selectedExtension.yaml_content || '');
    }
  }, [forkOpen, selectedExtension]);

  useEffect(() => {
    if (selectedExtension?.id && runModalOpen && extensions.length > 0) {
      const currentExtension = extensions.find((e) => e.id === selectedExtension.id);
      if (currentExtension) {
        const currentVars = JSON.stringify(currentExtension.variables || []);
        const selectedVars = JSON.stringify(selectedExtension.variables || []);
        if (currentVars !== selectedVars) {
          setSelectedExtension(currentExtension);
        }
      }
    }
  }, [extensions, runModalOpen, selectedExtension?.id]);

  useEffect(() => {
    if (!runModalOpen && !forkOpen) {
      setSelectedExtension(null);
    }
  }, [runModalOpen, forkOpen]);

  const doFork = async () => {
    try {
      await forkExtension({
        extensionId: selectedExtension?.extension_id || '',
        yaml_content: forkYaml || undefined
      }).unwrap();
      toast.success(t('extensions.forkSuccess'));
      setForkOpen(false);
    } catch (e) {
      toast.error(t('extensions.forkFailed'));
    }
  };

  const { values, errors, handleChange, handleSubmit, requiredFields } = useExtensionInput({
    extension: selectedExtension,
    open: runModalOpen,
    onSubmit: handleRun,
    onClose: () => setRunModalOpen(false)
  });

  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: () => setRunModalOpen(false),
      variant: 'ghost'
    },
    {
      label: t('extensions.run'),
      onClick: handleSubmit,
      variant: 'default'
    }
  ];

  const isOnlyProxyDomain =
    requiredFields.length === 1 &&
    (requiredFields[0].variable_name.toLowerCase() === 'proxy_domain' ||
      requiredFields[0].variable_name.toLowerCase() === 'domain');

  const noFieldsToShow = requiredFields.length === 0;

  return {
    extensions,
    isLoading,
    error,
    categories,
    searchTerm,
    sortConfig,
    currentPage,
    totalPages,
    totalExtensions,
    handleSearchChange,
    handleSortChange,
    selectedCategory,
    handleCategoryChange,
    handlePageChange,
    handleInstall,
    handleViewDetails,
    handleForkClick,
    handleRun,
    handleCancel,
    runModalOpen,
    setRunModalOpen,
    selectedExtension,
    sortOptions,
    forkOpen,
    setForkOpen,
    confirmOpen,
    setConfirmOpen,
    expanded,
    setExpanded,
    onDelete,
    forkYaml,
    setForkYaml,
    preview,
    variableColumns,
    doFork,
    actions,
    isOnlyProxyDomain,
    noFieldsToShow,
    values,
    errors,
    handleChange,
    handleSubmit,
    requiredFields
  };
}
