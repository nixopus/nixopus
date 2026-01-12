import { TableColumn } from '@/components/ui/data-table';
import { TypographySmall } from '@/components/ui/typography';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { MountData } from '@/packages/types/dashboard';

export function useDiskMountsColumns(): TableColumn<MountData>[] {
  const { t } = useTranslation();

  return [
    {
      key: 'mount',
      title: t('dashboard.disk.table.headers.mount'),
      dataIndex: 'mountPoint',
      className: 'text-xs pr-1 sm:pr-2',
      render: (mountPoint) => <TypographySmall className="text-xs">{mountPoint}</TypographySmall>
    },
    {
      key: 'size',
      title: t('dashboard.disk.table.headers.size'),
      dataIndex: 'size',
      className: 'text-xs pr-1 sm:pr-2',
      render: (size) => <TypographySmall className="text-xs">{size}</TypographySmall>
    },
    {
      key: 'used',
      title: t('dashboard.disk.table.headers.used'),
      dataIndex: 'used',
      className: 'text-xs pr-1 sm:pr-2',
      render: (used) => <TypographySmall className="text-xs">{used}</TypographySmall>
    },
    {
      key: 'capacity',
      title: t('dashboard.disk.table.headers.percentage'),
      dataIndex: 'capacity',
      className: 'text-xs',
      render: (capacity) => <TypographySmall className="text-xs">{capacity}</TypographySmall>
    }
  ];
}
