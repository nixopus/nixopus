import React from 'react';
import { Domain } from '@/redux/types/domain';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { DomainTypeTag } from './domain-type-tag';
import { DomainActions } from './domain-actions';
import { useTranslation } from '@/hooks/use-translation';

interface DomainsTableProps {
  domains: Domain[];
}

function DomainsTable({ domains }: DomainsTableProps) {
  const { t } = useTranslation();

  const columns: TableColumn<Domain>[] = [
    {
      key: 'name',
      title: t('settings.domains.table.headers.domain'),
      dataIndex: 'name'
    },
    {
      key: 'type',
      title: t('settings.domains.table.headers.type'),
      render: (_, domain) => <DomainTypeTag isWildcard={domain.name.startsWith('*')} />
    },
    {
      key: 'actions',
      title: t('settings.domains.table.headers.actions'),
      render: (_, domain) => <DomainActions domain={domain} />,
      align: 'right'
    }
  ];

  return (
      <DataTable
        data={domains}
        columns={columns}
        containerClassName="divide-y divide-border"
        showBorder={true}
    />
  );
}

export default DomainsTable;
