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
import { ExternalLink, Clock, Package, Server } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Application } from '@/redux/types/applications';
import { Environment } from '@/redux/types/deploy-form';
import { Skeleton } from '@/components/ui/skeleton';

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
  const formattedDate = updated_at
    ? new Date(updated_at).toLocaleString('en-US', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    : 'N/A';

  const formattedBuildPack = build_pack
    .replace(/([A-Z])/g, ' $1')
    .trim()
    .toLowerCase();

  const getEnvironmentStyles = () => {
    switch (environment) {
      case Environment.Development.toLowerCase():
        return 'bg-yellow-500 text-white font-medium';
      case Environment.Staging.toLowerCase():
        return 'bg-orange-400 text-black font-medium';
      default:
        return 'bg-primary text-primary-foreground font-medium';
    }
  };

  const getStatusStyles = () => {
    return status?.status === 'failed'
      ? 'bg-destructive text-white'
      : status?.status === 'deployed'
        ? 'bg-success text-success-foreground'
        : 'bg-secondary text-secondary-foreground';
  };

  return (
    <Card
      className={`relative w-full bg-secondary cursor-pointer overflow-hidden transition-all duration-300 hover:shadow-xl  group`}
      onClick={() => {
        router.push(`/self-host/application/${id}`);
      }}
    >
      {/* <div className="absolute inset-0 bg-gradient-to-br from-transparent to-muted opacity-50"></div> */}

      <div className="absolute right-3 top-3 z-10 transition-transform duration-300 group-hover:scale-110">
        {domain && (
          <a
            href={`https://${domain}`}
            target="_blank"
            rel="noopener noreferrer"
            className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-muted-foreground transition-colors duration-200 hover:bg-primary hover:text-primary-foreground"
            title={`View on ${domain}`}
            onClick={(e) => e.stopPropagation()}
          >
            <ExternalLink size={16} />
          </a>
        )}
      </div>

      <CardHeader className="pb-2">
        <div className="flex items-center justify-between pr-8">
          <CardTitle className="text-xl font-bold tracking-tight group-hover:text-primary transition-colors duration-300">
            {name}
          </CardTitle>
        </div>
      </CardHeader>

      <CardContent className="flex flex-col space-y-3 pb-2 z-10 relative">
        <div className="flex items-center text-sm text-muted-foreground">
          <Clock size={16} className="mr-2 text-muted-foreground/70" />
          <CardDescription className="text-sm">{formattedDate}</CardDescription>
        </div>

        <div className="flex items-center text-sm text-muted-foreground">
          <Package size={16} className="mr-2 text-muted-foreground/70" />
          <CardDescription className="text-sm capitalize">{formattedBuildPack}</CardDescription>
        </div>
      </CardContent>

      <CardFooter className="flex items-center justify-between pt-2 pb-3 border-t border-border z-10 relative">
        <Badge className={`${getEnvironmentStyles()} shadow-sm px-3 py-1`}>{environment}</Badge>

        <div className="flex items-center">
          <Server size={14} className="mr-1 text-muted-foreground/70" />
          <Badge variant="outline" className="text-xs font-mono">
            {port}
          </Badge>
        </div>
      </CardFooter>
      <div className="absolute inset-0 bg-gradient-to-br from-accent/10 to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100"></div>
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
