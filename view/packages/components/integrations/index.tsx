'use client';

import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRBAC } from '@/packages/utils/rbac';
import { MainPageHeader, SearchBar, PaginationWrapper } from '@nixopus/ui';
import PageLayout from '@/packages/layouts/page-layout';
import { IntegrationCard } from './integration-card';
import { IntegrationConfigModal } from './integration-config-modal';
import { MCPProviderCard } from './mcp-provider-card';
import { MCPProviderModal } from './mcp-provider-modal';
import { SortSelect } from '@/components/ui/sort-selector';
import type { SortOption } from '@/components/ui/sort-selector';
import { useIntegrations } from '@/packages/hooks/integrations/use-integrations';
import { getBaseUrl } from '@/redux/conf';
import type {
  IntegrationDefinition,
  IntegrationId
} from '@/packages/hooks/integrations/use-integrations';
import type { MCPProvider } from '@/redux/types/mcp';
import type { SMTPFormData } from '@/redux/types/notification';

type UnifiedCard =
  | { type: 'integration'; name: string; data: IntegrationDefinition }
  | { type: 'mcp'; name: string; data: MCPProvider };

type SortKey = 'name';
const SORT_OPTIONS: SortOption<{ name: SortKey }>[] = [
  { value: 'name', label: 'Name A → Z', direction: 'asc' },
  { value: 'name', label: 'Name Z → A', direction: 'desc' }
];

export function IntegrationsPage() {
  const { t } = useTranslation();
  const { canAccessResource } = useRBAC();
  const {
    integrations,
    isLoading,
    selectedIntegration,
    openModal,
    closeModal,
    getConfigForIntegration,
    handleOnSave,
    handleCreateWebhookConfig,
    handleUpdateWebhookConfig,
    handleDeleteWebhookConfig,
    handleDeleteSMTPConfiguration,
    slackConfig,
    discordConfig,
    mcpServers,
    mcpCatalog,
    catalogTotalCount,
    catalogPage,
    setCatalogPage,
    catalogPageSize,
    search,
    handleSearchChange,
    handleSortChange
  } = useIntegrations();

  const [selectedMCPProvider, setSelectedMCPProvider] = useState<MCPProvider | null>(null);
  const [iconBaseUrl, setIconBaseUrl] = useState('');
  const [currentSort, setCurrentSort] = useState(SORT_OPTIONS[0]);

  useEffect(() => {
    getBaseUrl()
      .then(setIconBaseUrl)
      .catch(() => {});
  }, []);

  const canConfigure =
    canAccessResource('notification', 'create') || canAccessResource('notification', 'update');
  const canDelete = canAccessResource('notification', 'delete');
  const canConfigureMCP = canAccessResource('mcp', 'create') || canAccessResource('mcp', 'update');
  const canDeleteMCP = canAccessResource('mcp', 'delete');

  const handleSaveWebhook = async (
    data: { webhook_url: string; is_active: boolean },
    type: 'slack' | 'discord'
  ) => {
    const existing = type === 'slack' ? slackConfig : discordConfig;
    if (existing) {
      await handleUpdateWebhookConfig({
        type,
        webhook_url: data.webhook_url,
        is_active: data.is_active
      });
    } else {
      await handleCreateWebhookConfig({ type, webhook_url: data.webhook_url });
    }
  };

  const onSortChange = (opt: SortOption<{ name: SortKey }>) => {
    setCurrentSort(opt);
    handleSortChange(String(opt.value), opt.direction);
  };

  // Merge static integrations and catalog providers into one unified array,
  // then sort the whole thing together so all cards honour the selected direction.
  const displayCards = useMemo<UnifiedCard[]>(() => {
    const q = search.toLowerCase();

    const staticCards: UnifiedCard[] = integrations
      .filter((i) => !q || i.name.toLowerCase().includes(q))
      .map((i) => ({ type: 'integration', name: i.name, data: i }));

    const mcpCards: UnifiedCard[] = mcpCatalog.map((p) => ({
      type: 'mcp',
      name: p.name,
      data: p
    }));

    const all = [...staticCards, ...mcpCards];
    all.sort((a, b) => {
      const cmp = a.name.localeCompare(b.name);
      return currentSort.direction === 'asc' ? cmp : -cmp;
    });
    return all;
  }, [integrations, mcpCatalog, search, currentSort.direction]);

  const selectedMCPServer = selectedMCPProvider
    ? (mcpServers.find((s) => s.provider_id === selectedMCPProvider.id) ?? null)
    : null;

  const catalogTotalPages = Math.ceil(catalogTotalCount / catalogPageSize);

  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <MainPageHeader label={t('integrations.title' as any)} highlightLabel={false} />

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex-grow max-w-sm">
          <SearchBar
            searchTerm={search}
            handleSearchChange={handleSearchChange}
            label="Search integrations…"
          />
        </div>
        <SortSelect<{ name: SortKey }>
          options={SORT_OPTIONS}
          currentSort={currentSort}
          onSortChange={onSortChange}
          placeholder="Sort by"
        />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {displayCards.map((card) => {
          if (card.type === 'integration') {
            return (
              <IntegrationCard
                key={card.data.id}
                integration={card.data}
                config={getConfigForIntegration(card.data.id as IntegrationId)}
                isLoading={isLoading}
                onConfigure={openModal}
                canConfigure={canConfigure}
              />
            );
          }
          return (
            <MCPProviderCard
              key={card.data.id}
              provider={card.data}
              server={mcpServers.find((s) => s.provider_id === card.data.id) ?? null}
              isLoading={isLoading}
              onConfigure={setSelectedMCPProvider}
              canConfigure={canConfigureMCP}
              iconBaseUrl={iconBaseUrl}
            />
          );
        })}
      </div>

      {!isLoading && displayCards.length === 0 && (
        <p className="text-sm text-muted-foreground text-center py-8">
          No integrations match &quot;{search}&quot;.
        </p>
      )}

      {catalogTotalPages > 1 && (
        <div className="flex justify-center mt-2">
          <PaginationWrapper
            currentPage={catalogPage}
            totalPages={catalogTotalPages}
            onPageChange={setCatalogPage}
          />
        </div>
      )}

      {selectedIntegration && (
        <IntegrationConfigModal
          integration={selectedIntegration}
          config={getConfigForIntegration(selectedIntegration.id as IntegrationId)}
          onClose={closeModal}
          onSaveSMTP={async (data: SMTPFormData) => {
            await handleOnSave(data);
          }}
          onSaveWebhook={handleSaveWebhook}
          onDeleteSMTP={handleDeleteSMTPConfiguration}
          onDeleteWebhook={handleDeleteWebhookConfig}
          canDelete={canDelete}
          canConfigure={canConfigure}
          isLoading={isLoading}
        />
      )}

      {selectedMCPProvider && (
        <MCPProviderModal
          provider={selectedMCPProvider}
          server={selectedMCPServer}
          onClose={() => setSelectedMCPProvider(null)}
          canDelete={canDeleteMCP}
        />
      )}
    </PageLayout>
  );
}
