'use client';
import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { ContainerData } from '@/redux/types/monitor';
import getStatusColor from '../utils/get-status-color';
import truncateId from '../utils/truncate-id';
import { useTranslation } from '@/hooks/use-translation';

const ContainersTable = ({ containersData }: { containersData: ContainerData[] }) => {
  const { t } = useTranslation();
  const hasContainers = containersData && containersData.length > 0;

  return (
    <div className="rounded-md">
      <Table className="border-0">
        <TableHeader>
          <TableRow>
            <TableHead className="w-[100px]">
              {t('dashboard.containers.table.headers.id')}
            </TableHead>
            <TableHead>{t('dashboard.containers.table.headers.name')}</TableHead>
            <TableHead>{t('dashboard.containers.table.headers.image')}</TableHead>
            <TableHead>{t('dashboard.containers.table.headers.status')}</TableHead>
            <TableHead>{t('dashboard.containers.table.headers.ports')}</TableHead>
            <TableHead>{t('dashboard.containers.table.headers.created')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {hasContainers ? (
            containersData.map((container) => {
              const containerName =
                container.Names && container.Names.length > 0
                  ? container.Names[0].replace(/^\//, '')
                  : '-';

              const hasPorts = container.Ports && container.Ports.length > 0;
              const formattedDate = container.Created
                ? new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'long' }).format(
                    new Date(container.Created * 1000)
                  )
                : '-';

              return (
                <TableRow key={container.Id}>
                  <TableCell className="font-mono text-xs">{truncateId(container.Id)}</TableCell>
                  <TableCell>{containerName}</TableCell>
                  <TableCell className="max-w-[200px] truncate">{container.Image}</TableCell>
                  <TableCell>
                    <Badge className={getStatusColor(container.Status)}>
                      {container.State || 'Unknown'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {hasPorts ? (
                      <div className="flex flex-col gap-1">
                        {container.Ports.slice(0, 2).map((port, index) => (
                          <span key={index} className="text-xs">
                            {port.PrivatePort}
                            {port.PublicPort ? `:${port.PublicPort}` : ''}
                          </span>
                        ))}
                        {container.Ports.length > 2 && (
                          <span className="text-xs text-gray-500">
                            {t('dashboard.containers.table.morePorts').replace(
                              '{count}',
                              String(container.Ports.length - 2)
                            )}
                          </span>
                        )}
                      </div>
                    ) : (
                      '-'
                    )}
                  </TableCell>
                  <TableCell>{formattedDate}</TableCell>
                </TableRow>
              );
            })
          ) : (
            <TableRow>
              <TableCell colSpan={7} className="text-center py-6 text-gray-500">
                {t('dashboard.containers.table.noContainers')}
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
};

export default ContainersTable;
