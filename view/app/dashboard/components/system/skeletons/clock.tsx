import { Skeleton } from '@/components/ui/skeleton';

export const ClockCardSkeletonContent = () => {
  return (
    <div className="flex flex-col items-center justify-center space-y-3">
      <Skeleton className="h-20 w-64" />
      <Skeleton className="h-4 w-56" />
    </div>
  );
};
