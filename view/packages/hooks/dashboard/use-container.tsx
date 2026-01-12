import { TableColumn } from '@/components/ui/data-table';
import { ContainerData } from '@/redux/types/monitor';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { TypographyMuted, TypographySmall } from '@/components/ui/typography';
import { Badge } from '@/components/ui/badge';
import { truncateId, getStatusColor } from '@/packages/utils/dashboard';

export function useContainer() {
  const { t } = useTranslation();

  const columns: TableColumn<ContainerData>[] = [
    {
      key: 'id',
      title: t('dashboard.containers.table.headers.id'),
      dataIndex: 'Id',
      width: '100px',
      render: (id) => <TypographySmall className="font-mono">{truncateId(id)}</TypographySmall>
    },
    {
      key: 'name',
      title: t('dashboard.containers.table.headers.name'),
      render: (_, container) => {
        const containerName =
          container.Names && container.Names.length > 0
            ? container.Names[0].replace(/^\//, '')
            : '-';
        return <TypographySmall>{containerName}</TypographySmall>;
      }
    },
    {
      key: 'image',
      title: t('dashboard.containers.table.headers.image'),
      dataIndex: 'Image',
      className: 'max-w-[200px] overflow-hidden',
      render: (image) => (
        <TypographySmall className="truncate whitespace-nowrap max-w-full block">
          {image}{' '}
        </TypographySmall>
      )
    },
    {
      key: 'status',
      title: t('dashboard.containers.table.headers.status'),
      render: (_, container) => (
        <Badge className={getStatusColor(container.Status)}>{container.State || 'Unknown'}</Badge>
      )
    },
    {
      key: 'ports',
      title: t('dashboard.containers.table.headers.ports'),
      render: (_, container) => {
        const hasPorts = container.Ports && container.Ports.length > 0;

        if (!hasPorts) {
          return <TypographySmall>-</TypographySmall>;
        }

        return (
          <div className="flex flex-col gap-1">
            {container.Ports.slice(0, 2).map((port, index) => (
              <TypographySmall key={index}>
                {port.PrivatePort}
                {port.PublicPort ? `:${port.PublicPort}` : ''}{' '}
              </TypographySmall>
            ))}
            {container.Ports.length > 2 && (
              <TypographyMuted>
                {' '}
                {t('dashboard.containers.table.morePorts').replace(
                  '{count}',
                  String(container.Ports.length - 2)
                )}{' '}
              </TypographyMuted>
            )}{' '}
          </div>
        );
      }
    },
    {
      key: 'created',
      title: t('dashboard.containers.table.headers.created'),
      dataIndex: 'Created',
      render: (created) => {
        const formattedDate = created
          ? new Intl.DateTimeFormat(undefined, {
              day: 'numeric',
              month: 'long'
            }).format(new Date(created * 1000))
          : '-';
        return <TypographySmall>{formattedDate}</TypographySmall>;
      }
    }
  ];
  return columns;
}
