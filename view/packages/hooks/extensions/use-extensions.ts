'use client';

import { useEffect, useState } from 'react';
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
  useGetExtensionCategoriesQuery
} from '@/redux/services/extensions/extensionsApi';
import { useCreateTemplateDeploymentMutation } from '@/redux/services/deploy/applicationsApi';
import { SelectOption } from '@nixopus/ui';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';
import { useExtensionInput } from './use-extension-input';
import { useServerSelector } from '@/packages/hooks/deploy/use-server-selector';
import { DialogAction } from '@nixopus/ui';

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
  const serverSelector = useServerSelector();
  const [expanded, setExpanded] = useState(false);

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
    setCurrentPage(1);
  };

  const handleSortChange = (key: ExtensionSortField, direction: SortDirection) => {
    setSortConfig({ key, direction });
    setCurrentPage(1);
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

  const error = apiError ? 'Failed to load extensions' : null;

  const [createTemplateDeployment, { isLoading: isInstalling }] =
    useCreateTemplateDeploymentMutation();

  const handleRun = async (values: Record<string, unknown>) => {
    if (!selectedExtension) return;
    try {
      const app = await createTemplateDeployment({
        template_id: selectedExtension.extension_id,
        name: selectedExtension.name,
        variables: values,
        ...(serverSelector.selectedServerIds.length > 0 && {
          server_ids: serverSelector.selectedServerIds
        }),
        ...(serverSelector.routingStrategy && {
          routing_strategy: serverSelector.routingStrategy
        }),
        ...(serverSelector.primaryServerId && {
          primary_server_id: serverSelector.primaryServerId
        })
      }).unwrap();
      setRunModalOpen(false);
      router.push(`/apps/application/${app.id}`);
      toast.success(t('extensions.installSuccess') || 'App created successfully');
    } catch (e) {
      toast.error(t('extensions.installFailed') || 'Failed to create app');
    }
  };

  const sortOptions: SelectOption[] = [
    { value: 'name_asc', label: t('extensions.sortOptions.name') + ' (A-Z)' },
    { value: 'name_desc', label: t('extensions.sortOptions.name') + ' (Z-A)' }
  ];

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
    if (!runModalOpen) {
      setSelectedExtension(null);
    }
  }, [runModalOpen]);

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
      label: t('extensions.install'),
      onClick: handleSubmit,
      variant: 'default',
      disabled: isInstalling,
      loading: isInstalling
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
    handleRun,
    runModalOpen,
    setRunModalOpen,
    selectedExtension,
    sortOptions,
    expanded,
    setExpanded,
    actions,
    isOnlyProxyDomain,
    noFieldsToShow,
    values,
    errors,
    handleChange,
    handleSubmit,
    requiredFields,
    serverSelector
  };
}
