'use client';

import { RefreshCw, Play, StopCircle, Trash2, Loader2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import {
  useGetContainersQuery,
  useStartContainerMutation,
  useStopContainerMutation,
  useRemoveContainerMutation
} from '@/redux/services/container/containerApi';
import { cn } from '@/lib/utils';
import { useRouter } from 'next/navigation';
import ContainersLoading from './skeleton';
import Autoplay from 'embla-carousel-autoplay';
import { Carousel, CarouselContent, CarouselItem } from '@/components/ui/carousel';
import { useState } from 'react';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FeatureNames } from '@/types/feature-flags';
import { Skeleton } from '@/components/ui/skeleton';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import { useAppSelector } from '@/redux/hooks';
import { hasPermission } from '@/lib/permission';

const getGradientFromName = (name: string) => {
  const colors = [
    'from-blue-500/20 to-purple-500/20',
    'from-green-500/20 to-teal-500/20',
    'from-yellow-500/20 to-orange-500/20',
    'from-red-500/20 to-pink-500/20',
    'from-indigo-500/20 to-violet-500/20',
    'from-emerald-500/20 to-cyan-500/20'
  ];
  const index = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length;
  return colors[index];
};

export default function ContainersPage() {
  const { t } = useTranslation();
  const router = useRouter();
  const user = useAppSelector((state) => state.auth.user);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const { data: containers = [], isLoading, refetch } = useGetContainersQuery();
  const [startContainer] = useStartContainerMutation();
  const [stopContainer] = useStopContainerMutation();
  const [removeContainer] = useRemoveContainerMutation();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [containerToDelete, setContainerToDelete] = useState<string | null>(null);
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

  const canRead = hasPermission(user, 'container', 'read', activeOrg?.id);
  const canCreate = hasPermission(user, 'container', 'create', activeOrg?.id);
  const canUpdate = hasPermission(user, 'container', 'update', activeOrg?.id);
  const canDelete = hasPermission(user, 'container', 'delete', activeOrg?.id);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await refetch();
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleContainerAction = async (containerId: string, action: 'start' | 'stop' | 'remove') => {
    try {
      switch (action) {
        case 'start':
          await startContainer(containerId).unwrap();
          toast.success(t(`containers.${action}_success`));
          break;
        case 'stop':
          await stopContainer(containerId).unwrap();
          toast.success(t(`containers.${action}_success`));
          break;
        case 'remove':
          setContainerToDelete(containerId);
          break;
      }
    } catch (error) {
      toast.error(t(`containers.${action}_error`));
    }
  };

  const handleDeleteConfirm = async () => {
    if (!containerToDelete) return;
    try {
      await removeContainer(containerToDelete).unwrap();
      toast.success(t('containers.remove_success'));
      setContainerToDelete(null);
    } catch (error) {
      toast.error(t('containers.remove_error'));
    }
  };

  if (isLoading) {
    return <ContainersLoading />;
  }

  if (!canRead) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">{t('common.accessDenied')}</h2>
          <p className="text-muted-foreground">{t('common.noPermissionView')}</p>
        </div>
      </div>
    );
  }

  const featuredContainers = containers.slice(0, 3);
  const remainingContainers = containers.slice(3);

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureContainer)) {
    return <DisabledFeature />;
  }

  return (
    <div className="min-h-screen">
      <div className="relative">
        <div className="container mx-auto py-6 relative z-10">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-2xl font-bold">{t('containers.title')}</h1>
            <Button onClick={handleRefresh} variant="outline" size="sm" disabled={isRefreshing}>
              {
                isRefreshing ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="mr-2 h-4 w-4" />
                )
              }
              {t('containers.refresh')}
            </Button>
          </div>

          {featuredContainers.length > 0 && (
            <Carousel
              className="mx-auto mb-10 w-full"
              opts={{
                loop: true,
              }}
              plugins={[
                Autoplay({
                  delay: 3000,
                }),
              ]}
            >
              <CarouselContent>
                {featuredContainers.map((container) => (
                  <CarouselItem key={container.id}>
                    <div className="p-0">
                      <Card 
                        className={cn('overflow-hidden w-full relative cursor-pointer', getGradientFromName(container.name))}
                        onClick={() => router.push(`/containers/${container.id}`)}
                      >
                        <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f1a_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f1a_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
                        <div className="absolute inset-0 bg-gradient-to-br opacity-20 transition-opacity duration-300 group-hover:opacity-30" />
                        <CardContent className="flex flex-col items-center p-6 md:flex-row md:justify-between md:items-start relative z-10">
                          <div className="mb-4 md:mb-0 md:w-1/2">
                            <h2 className="mb-2 text-2xl font-bold text-primary">
                              {container.name}
                            </h2>
                            <p className="text-secondary-foreground">
                              {container.image}
                            </p>
                            <div className="mt-4 flex gap-2">
                              {container.status !== 'running' && canUpdate && (
                                <Button
                                  variant="outline"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleContainerAction(container.id, 'start');
                                  }}
                                >
                                  <Play className="mr-2 h-4 w-4" />
                                  Start
                                </Button>
                              )}
                              {container.status === 'running' && canUpdate && (
                                <Button
                                  variant="outline"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleContainerAction(container.id, 'stop');
                                  }}
                                >
                                  <StopCircle className="mr-2 h-4 w-4" />
                                  Stop
                                </Button>
                              )}
                              {canDelete && (
                                <Button
                                  variant="outline"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    handleContainerAction(container.id, 'remove');
                                  }}
                                >
                                  <Trash2 className="mr-2 h-4 w-4" />
                                  Remove
                                </Button>
                              )}
                            </div>
                          </div>
                          <div className="">
                            <div className="space-y-2">
                              <div className="text-sm">
                                <span className="font-medium">Status:</span>
                                <Badge variant={container.status === 'running' ? 'default' : 'secondary'} className="ml-2">
                                  {container.status}
                                </Badge>
                              </div>
                              <div className="text-sm">
                                <span className="font-medium">Ports:</span>
                                <div className="flex flex-wrap gap-2 mt-1">
                                  {container?.ports?.map((port: any) => (
                                    <Badge key={`${port.private_port}-${port.public_port}`} variant="outline">
                                      {port.public_port} → {port.private_port}
                                    </Badge>
                                  ))}
                                </div>
                              </div>
                              <div className="text-sm">
                                <span className="font-medium">Memory:</span>
                                <span className="ml-2">
                                  {(container.host_config.memory / (1024 * 1024)).toFixed(2)} MB
                                </span>
                              </div>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    </div>
                  </CarouselItem>
                ))}
              </CarouselContent>
            </Carousel>
          )}

          {remainingContainers.length > 0 && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {remainingContainers.map((container) => (
                <Card
                  key={container.id}
                  className={cn(
                    'group relative overflow-hidden transition-all duration-300 hover:shadow-lg cursor-pointer',
                    getGradientFromName(container.name)
                  )}
                  onClick={() => router.push(`/containers/${container.id}`)}
                >
                  <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f1a_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f1a_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
                  <div className="absolute inset-0 bg-gradient-to-br opacity-20 transition-opacity duration-300 group-hover:opacity-30" />
                  <CardContent className="relative p-6 z-10">
                    <div className="flex items-start justify-between">
                      <div className="space-y-2">
                        <h3 className="text-xl font-semibold">{container.name}</h3>
                        <p className="text-sm text-muted-foreground truncate">{container.image}</p>
                        <Badge variant={container.status === 'running' ? 'default' : 'secondary'}>
                          {container.status}
                        </Badge>
                      </div>
                      <div className="flex gap-2">
                        {container.status !== 'running' && canUpdate && (
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleContainerAction(container.id, 'start');
                            }}
                          >
                            <Play className="h-4 w-4" />
                          </Button>
                        )}
                        {container.status === 'running' && canUpdate && (
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleContainerAction(container.id, 'stop');
                            }}
                          >
                            <StopCircle className="h-4 w-4" />
                          </Button>
                        )}
                        {canDelete && (
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleContainerAction(container.id, 'remove');
                            }}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                    <div className="mt-4 space-y-2">
                      <div className="text-sm">
                        <span className="font-medium">Ports:</span>
                        <div className="flex flex-wrap gap-2 mt-1">
                          {container?.ports?.map((port: any) => (
                            <Badge key={`${port.private_port}-${port.public_port}`} variant="outline">
                              {port.public_port} → {port.private_port}
                            </Badge>
                          ))}
                        </div>
                      </div>
                      <div className="text-sm">
                        <span className="font-medium">Memory:</span>
                        <span className="ml-2">
                          {(container.host_config.memory / (1024 * 1024)).toFixed(2)} MB
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </div>
      <DeleteDialog
        title={t('containers.deleteDialog.title')}
        description={t('containers.deleteDialog.description')}
        onConfirm={handleDeleteConfirm}
        open={!!containerToDelete}
        onOpenChange={(open) => !open && setContainerToDelete(null)}
        variant="destructive"
        confirmText={t('containers.deleteDialog.confirm')}
        cancelText={t('containers.deleteDialog.cancel')}
        icon={Trash2}
      />
    </div>
  );
}