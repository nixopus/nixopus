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
import { ContainerData } from '../hooks/use_monitor';
import getStatusColor from '../utils/getStatusColor';
import truncateId from '../utils/truncateId';

const ContainersTable = ({ containersData }: { containersData: ContainerData[] }) => {
  const hasContainers = containersData && containersData.length > 0;

  return (
    <div className="rounded-md">
      <Table className="border-0">
        <TableHeader>
          <TableRow>
            <TableHead className="w-[100px]">ID</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Image</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Ports</TableHead>
            <TableHead>Created</TableHead>
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
                            +{container.Ports.length - 2} more
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
                No containers found
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
};

export default ContainersTable;
