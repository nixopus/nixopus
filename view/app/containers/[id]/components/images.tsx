import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { useTranslation } from '@/hooks/use-translation';
import { formatBytes, formatDate } from '@/lib/utils';
import { useGetImagesQuery } from '@/redux/services/container/imagesApi';

interface Image {
  id: string;
  repo_tags: string[];
  repo_digests: string[];
  created: number;
  size: number;
  shared_size: number;
  virtual_size: number;
  labels: Record<string, string>;
}

export function Images({ containerId, imagePrefix }: { containerId: string; imagePrefix: string }) {
  const { data: images = [], isLoading } = useGetImagesQuery({ containerId, imagePrefix });
  const { t } = useTranslation();

  if (isLoading) {
    return <ImagesSectionSkeleton />;
  }

  const columns: TableColumn<Image>[] = [
    {
      key: 'id',
      title: t('containers.images.id'),
      dataIndex: 'id',
      render: (id) => <span className="font-mono">{id.slice(0, 12)}</span>
    },
    {
      key: 'tags',
      title: t('containers.images.tags'),
      dataIndex: 'repo_tags',
      render: (tags) => tags?.join(', ') || '<none>'
    },
    {
      key: 'created',
      title: t('containers.images.created'),
      dataIndex: 'created',
      render: (created) => formatDate(new Date(created * 1000))
    },
    {
      key: 'size',
      title: t('containers.images.size'),
      dataIndex: 'size',
      render: (size) => formatBytes(size)
    }
  ];

  return (
    <DataTable
      data={images}
      columns={columns}
      loading={isLoading}
      loadingRows={5}
      showBorder={true}
      hoverable={false}
    />
  );
}

export function ImagesSectionSkeleton() {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <Skeleton className="h-6 w-32" />
              <Skeleton className="h-4 w-48" />
            </div>
            <div className="flex gap-2">
              <Skeleton className="h-9 w-32" />
              <Skeleton className="h-9 w-32" />
            </div>
          </div>
          <div className="space-y-2">
            <div className="grid grid-cols-5 gap-4">
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-4 w-16" />
            </div>
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="grid grid-cols-5 gap-4">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-4 w-16" />
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
