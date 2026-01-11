import { Skeleton } from '@/components/ui/skeleton';

export const ClockCardSkeletonContent = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full space-y-3">
      <Skeleton className="h-12 w-40" /> {/* text-5xl time */}
      <Skeleton className="h-4 w-32" /> {/* text-sm date */}
    </div>
  );
};
