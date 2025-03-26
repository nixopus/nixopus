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

const getStatusColor = (status: string) => {
  if (status?.toLowerCase().includes('running')) return 'bg-green-100 text-green-800';
  if (status?.toLowerCase().includes('exited')) return 'bg-red-100 text-red-800';
  if (status?.toLowerCase().includes('created')) return 'bg-blue-100 text-blue-800';
  return 'bg-gray-100 text-gray-800';
};

const truncateId = (id: string) => {
  return id?.substring(0, 12) || '';
};

const ContainersTable = ({ containersData }: { containersData: ContainerData[] }) => {
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
          {containersData && containersData.length > 0 ? (
            containersData.map((container) => (
              <TableRow key={container.Id}>
                <TableCell className="font-mono text-xs">{truncateId(container.Id)}</TableCell>
                <TableCell>
                  {container.Names && container.Names.length > 0
                    ? container.Names[0].replace(/^\//, '')
                    : '-'}
                </TableCell>
                <TableCell className="max-w-[200px] truncate">{container.Image}</TableCell>
                <TableCell>
                  <Badge className={getStatusColor(container.Status)}>
                    {container.State || 'Unknown'}
                  </Badge>
                </TableCell>
                <TableCell>
                  {container.Ports && container.Ports.length > 0 ? (
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
                <TableCell>
                  {container.Created
                    ? new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'long' }).format(
                        new Date(container.Created * 1000)
                      )
                    : '-'}
                </TableCell>
              </TableRow>
            ))
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
