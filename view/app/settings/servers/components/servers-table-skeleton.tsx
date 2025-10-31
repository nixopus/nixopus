import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import type { Server } from '@/redux/types/server';

function ServersTableSkeleton() {
  const { t } = useTranslation();

  const columns: TableColumn<Server>[] = [
    { key: 'name', title: t('servers.table.headers.name') },
    { key: 'host', title: t('servers.table.headers.host') },
    { key: 'port', title: t('servers.table.headers.port') },
    { key: 'username', title: t('servers.table.headers.username') },
    { key: 'created_at', title: t('servers.table.headers.created') },
    { key: 'actions', title: '', width: '70px', align: 'right' }
  ];

  return (
    <DataTable<Server>
      data={[]}
      columns={columns}
      loading={true}
      loadingRows={5}
      showBorder
    />
  );
}

export default ServersTableSkeleton;
