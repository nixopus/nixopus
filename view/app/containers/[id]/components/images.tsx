import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { useTranslation } from '@/hooks/use-translation';
import { formatBytes, formatDate } from '@/lib/utils';
import { useGetImagesQuery } from '@/redux/services/container/imagesApi';
import { Loader2 } from 'lucide-react';

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

  return (
    <Card>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('containers.images.id')}</TableHead>
              <TableHead>{t('containers.images.tags')}</TableHead>
              <TableHead>{t('containers.images.created')}</TableHead>
              <TableHead>{t('containers.images.size')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center">
                  <Loader2 className="mx-auto h-6 w-6 animate-spin" />
                </TableCell>
              </TableRow>
            ) : (
              images?.map((image: Image) => (
                <TableRow key={image.id}>
                  <TableCell className="font-mono">{image.id.slice(0, 12)}</TableCell>
                  <TableCell>{image.repo_tags?.join(', ') || '<none>'}</TableCell>
                  <TableCell>{formatDate(new Date(image.created * 1000))}</TableCell>
                  <TableCell>{formatBytes(image.size)}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
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
