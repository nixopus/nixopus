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
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import truncateId from '@/app/dashboard/components/utils/truncate-id';
import getStatusColor from '@/app/dashboard/components/utils/get-status-color';
import { Container } from '@/redux/services/container/containerApi';
import { useRouter } from 'next/navigation';
import { ArrowUpDown } from 'lucide-react';
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
    const hasContainers = containersData && containersData.length > 0;
    const router = useRouter();

    return (
        <div className="rounded-md">
            <Table className="border">
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-[100px]">
                            <TypographySmall>{t('dashboard.containers.table.headers.id')}</TypographySmall>
                        </TableHead>
                        <TableHead onClick={() => onSort && onSort('name')} className={`select-none ${onSort ? 'cursor-pointer' : ''}`}>
                            <div className="flex items-center gap-1">
                                <TypographySmall>{t('dashboard.containers.table.headers.name')}</TypographySmall>
                                <ArrowUpDown className={`h-3 w-3 ${sortBy === 'name' ? 'opacity-100' : 'opacity-40'}`} />
                                <span className="sr-only">Sort</span>
                            </div>
                        </TableHead>
                        <TableHead>
                            <TypographySmall>{t('dashboard.containers.table.headers.image')}</TypographySmall>
                        </TableHead>
                        <TableHead onClick={() => onSort && onSort('status')} className={`select-none ${onSort ? 'cursor-pointer' : ''}`}>
                            <div className="flex items-center gap-1">
                                <TypographySmall>{t('dashboard.containers.table.headers.status')}</TypographySmall>
                                <ArrowUpDown className={`h-3 w-3 ${sortBy === 'status' ? 'opacity-100' : 'opacity-40'}`} />
                                <span className="sr-only">Sort</span>
                            </div>
                        </TableHead>
                        <TableHead>
                            <TypographySmall>{t('dashboard.containers.table.headers.ports')}</TypographySmall>
                        </TableHead>
                        <TableHead>
                            <TypographySmall>{t('dashboard.containers.table.headers.created')}</TypographySmall>
                        </TableHead>
                        {onAction && (
                            <TableHead>
                                <TypographySmall>Actions</TypographySmall>
                            </TableHead>
                        )}
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {hasContainers ? (
                        containersData.map((container) => {
                            const containerName = container.name

                            const hasPorts = container.ports && container.ports.length > 0;
                            const formattedDate = container.created
                                ? new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short', year: 'numeric' }).format(
                                    new Date(container.created)
                                )
                                : '-';

                            return (
                                <TableRow key={container.id} onClick={() => router.push(`/containers/${container.id}`)}
                                    className='cursor-pointer'
                                >
                                    <TableCell>
                                        <TypographySmall>{truncateId(container.id)}</TypographySmall>
                                    </TableCell>
                                    <TableCell>
                                        <TypographySmall>{containerName}</TypographySmall>
                                    </TableCell>
                                    <TableCell className="max-w-[200px] overflow-hidden">
                                        <TypographySmall className='truncate whitespace-nowrap max-w-full block'>{container.image}</TypographySmall>
                                    </TableCell>
                                    <TableCell>
                                        <Badge className={getStatusColor(container.status)}>
                                            {container.state || 'Unknown'}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>
                                        {hasPorts ? (
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
                                        ) : (
                                            <TypographySmall>-</TypographySmall>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <TypographySmall>{formattedDate}</TypographySmall>
                                    </TableCell>
                                    {onAction && (
                                        <TableCell onClick={(e) => e.stopPropagation()}>
                                            <ContainerActions container={container} onAction={onAction} />
                                        </TableCell>
                                    )}
                                </TableRow>
                            );
                        })
                    ) : (
                        <TableRow>
                            <TableCell colSpan={onAction ? 7 : 6} className="text-center py-6">
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
