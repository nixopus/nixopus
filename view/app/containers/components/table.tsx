'use client';
import React from 'react';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { Badge } from '@/components/ui/badge';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import truncateId from '@/app/dashboard/components/utils/truncate-id';
import getStatusColor from '@/app/dashboard/components/utils/get-status-color';
import { Container } from '@/redux/services/container/containerApi';
import { useRouter } from 'next/navigation';
import { ContainerActions } from './actions';
import { Action } from './card';

type SortField = 'name' | 'status';

const ContainersTable = ({
  containersData,
  sortBy = 'name',
  sortOrder = 'asc',
  onSort,
  onAction
}: {
  containersData: Container[];
  sortBy?: SortField;
  sortOrder?: 'asc' | 'desc';
  onSort?: (field: SortField) => void;
  onAction?: (id: string, action: Action) => void;
}) => {
  const { t } = useTranslation();
  const router = useRouter();

  const handleRowClick = (container: Container) => {
    router.push(`/containers/${container.id}`);
  };

  const handleSort = (field: string) => {
    if (onSort && (field === 'name' || field === 'status')) {
      onSort(field as SortField);
    }
  };

  const columns: TableColumn<Container>[] = [
    {
      key: 'id',
      title: t('dashboard.containers.table.headers.id'),
      dataIndex: 'id',
      width: '100px',
      render: (id) => <TypographySmall>{truncateId(id)}</TypographySmall>
    },
    {
      key: 'name',
      title: t('dashboard.containers.table.headers.name'),
      dataIndex: 'name',
      sortable: true,
      render: (name) => <TypographySmall>{name}</TypographySmall>
    },
    {
      key: 'image',
      title: t('dashboard.containers.table.headers.image'),
      dataIndex: 'image',
      className: 'max-w-[200px] overflow-hidden',
      render: (image) => (
        <TypographySmall className="truncate whitespace-nowrap max-w-full block">
          {image}
        </TypographySmall>
      )
    },
    {
      key: 'status',
      title: t('dashboard.containers.table.headers.status'),
      dataIndex: 'state',
      sortable: true,
      render: (state, container) => (
        <Badge className={getStatusColor(container.status)}>
          {state || 'Unknown'}
        </Badge>
      )
    },
    {
      key: 'ports',
      title: t('dashboard.containers.table.headers.ports'),
      render: (_, container) => {
        const hasPorts = container.ports && container.ports.length > 0;
        
        if (!hasPorts) {
          return <TypographySmall>-</TypographySmall>;
        }

        return (
          <div className="flex flex-col gap-1">
            {container.ports.slice(0, 2).map((port, index) => (
              <TypographySmall key={index}>
                {port.private_port}
                {port.public_port ? `:${port.public_port}` : ''}
              </TypographySmall>
            ))}
            {container.ports.length > 2 && (
              <TypographyMuted>
                {t('dashboard.containers.table.morePorts').replace(
                  '{count}',
                  String(container.ports.length - 2)
                )}
              </TypographyMuted>
            )}
          </div>
        );
      }
    },
    {
      key: 'created',
      title: t('dashboard.containers.table.headers.created'),
      dataIndex: 'created',
      render: (created) => {
        const formattedDate = created
          ? new Intl.DateTimeFormat(undefined, {
              day: 'numeric',
              month: 'short',
              year: 'numeric'
            }).format(new Date(created))
          : '-';
        return <TypographySmall>{formattedDate}</TypographySmall>;
      }
    }
  ];

  if (onAction) {
    columns.push({
      key: 'actions',
      title: 'Actions',
      render: (_, container) => (
        <div onClick={(e) => e.stopPropagation()}>
          <ContainerActions container={container} onAction={onAction} />
        </div>
      )
    });
  }

  return (
      <DataTable
        data={containersData}
        columns={columns}
        onRowClick={handleRowClick}
        onSort={handleSort}
        sortConfig={{
          field: sortBy,
          order: sortOrder
        }}
        emptyMessage={t('dashboard.containers.table.noContainers')}
        hoverable={true}
        showBorder={true}
      />
  );
};

export default ContainersTable;
