import { Skeleton } from '@/components/ui/skeleton';

export const WeatherCardSkeletonContent = () => {
  return (
    <div className="flex flex-col items-center justify-center space-y-3">
      <Skeleton className="h-16 w-16 rounded-full" />
      <Skeleton className="h-16 w-32" />
      <Skeleton className="h-4 w-40" />
    </div>
  );
};
