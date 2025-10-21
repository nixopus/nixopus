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
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

const ContainersTable = ({ containersData }: { containersData: ContainerData[] }) => {
  const { t } = useTranslation();
  const hasContainers = containersData && containersData.length > 0;

  return (
    <div className="rounded-md">
      <Table className="border-0">
        <TableHeader>
          <TableRow>
            <TableHead className="w-[100px]">
              <TypographySmall>{t('dashboard.containers.table.headers.id')}</TypographySmall>
            </TableHead>
            <TableHead>
              <TypographySmall>{t('dashboard.containers.table.headers.name')}</TypographySmall>
            </TableHead>
            <TableHead>
              <TypographySmall>{t('dashboard.containers.table.headers.image')}</TypographySmall>
            </TableHead>
            <TableHead>
              <TypographySmall>{t('dashboard.containers.table.headers.status')}</TypographySmall>
            </TableHead>
            <TableHead>
              <TypographySmall>{t('dashboard.containers.table.headers.ports')}</TypographySmall>
            </TableHead>
            <TableHead>
              <TypographySmall>{t('dashboard.containers.table.headers.created')}</TypographySmall>
            </TableHead>
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
                  <TableCell className="font-mono">
                    <TypographySmall>{truncateId(container.Id)}</TypographySmall>
                  </TableCell>
                  <TableCell>
                    <TypographySmall>{containerName}</TypographySmall>
                  </TableCell>
                  <TableCell className="max-w-[200px] overflow-hidden">
                    <TypographySmall className="truncate whitespace-nowrap max-w-full block">
                      {container.Image}
                    </TypographySmall>
                  </TableCell>
                  <TableCell>
                    <Badge className={getStatusColor(container.Status)}>
                      {container.State || 'Unknown'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {hasPorts ? (
                      <div className="flex flex-col gap-1">
                        {container.Ports.slice(0, 2).map((port, index) => (
                          <TypographySmall key={index}>
                            {port.PrivatePort}
                            {port.PublicPort ? `:${port.PublicPort}` : ''}
                          </TypographySmall>
                        ))}
                        {container.Ports.length > 2 && (
                          <TypographyMuted>
                            {t('dashboard.containers.table.morePorts').replace(
                              '{count}',
                              String(container.Ports.length - 2)
                            )}
                          </TypographyMuted>
                        )}
                      </div>
                    ) : (
                      <TypographySmall>-</TypographySmall>
                    )}
                  </TableCell>
                  <TableCell>
                    <TypographySmall>{formattedDate}</TypographySmall>
                  </TableCell>
                </TableRow>
              );
            })
          ) : (
            <TableRow>
              <TableCell colSpan={7} className="text-center py-6">
                <TypographyMuted>{t('dashboard.containers.table.noContainers')}</TypographyMuted>
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
};

export default ContainersTable;
