import React from 'react';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import { ExternalLink } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Skeleton } from '@/components/ui/skeleton';
import { Application } from '@/redux/types/applications';
import { Environment } from '@/redux/types/deploy-form';

function AppItem({
  name,
  domain,
  environment,
  updated_at,
  build_pack,
  port,
  id,
  status
}: Application) {
  const router = useRouter();
  return (
    <Card
      className="relative w-full  cursor-pointer overflow-hidden transition-all duration-300 hover:bg-muted hover:shadow-lg"
      onClick={() => router.push(`/self-host/configure/${id}?name=${name}`)}
    >
      <div className="absolute right-2 top-2">
        {domain && (
          <a
            href={`https://${domain}`}
            target="_blank"
            rel="noopener noreferrer"
            className="text-gray-500 transition-colors duration-200 hover:text-gray-700"
            title={`View on ${domain}`}
          >
            <ExternalLink size={20} />
          </a>
        )}
      </div>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between pr-8">
          <CardTitle className="text-xl font-bold">{name}</CardTitle>
        </div>
        <CardDescription className="self-end">
          <Badge>{status?.status}</Badge>
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col space-y-1 pb-2">
        <CardDescription>
          Last Deployed{' '}
          {new Date(updated_at || '').getDate() +
            '/' +
            (new Date(updated_at || '').getMonth() + 1) +
            '/' +
            new Date(updated_at || '').getFullYear() +
            ' ' +
            new Date(updated_at || '').toLocaleTimeString()}
        </CardDescription>
        <CardDescription>
          Type:{' '}
          {build_pack
            .replace(/([A-Z])/g, ' $1')
            .trim()
            .toLowerCase()}
        </CardDescription>
      </CardContent>
      <CardFooter className="flex items-center justify-between py-4">
        <Badge
          className={
            'text-xs ' +
            (environment === Environment.Development.toLowerCase()
              ? 'bg-yellow-500'
              : environment === Environment.Staging.toLowerCase()
                ? 'bg-orange-500'
                : 'bg-primary')
          }
        >
          {environment}
        </Badge>
        <Badge variant={'secondary'} className="text-xs">
          {port}
        </Badge>
      </CardFooter>
    </Card>
  );
}

export default AppItem;

export function AppItemSkeleton() {
  return (
    <Card className="relative w-full max-w-md">
      <div className="absolute right-2 top-2">
        <Skeleton className="h-5 w-5" />
      </div>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between pr-8">
          <CardTitle className="text-xl font-bold">
            <Skeleton className="h-6 w-32" />
          </CardTitle>
        </div>
        <CardDescription>
          <Skeleton className="h-4 w-full" />
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col space-y-1 pb-2">
        <CardDescription>
          <Skeleton className="h-4 w-48" />
        </CardDescription>
        <CardDescription>
          <Skeleton className="h-4 w-32" />
        </CardDescription>
      </CardContent>
      <CardFooter className="flex items-center justify-between py-4">
        <Skeleton className="h-5 w-16" />
        <Skeleton className="h-5 w-16" />
      </CardFooter>
    </Card>
  );
}
