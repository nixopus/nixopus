import { Skeleton } from '@/components/ui/skeleton';

export const WeatherCardSkeletonContent = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full space-y-3">
      {/* Weather icon */}
      <Skeleton className="h-12 w-12 rounded-full" />
      {/* Temperature - text-5xl */}
      <Skeleton className="h-12 w-24" />
      {/* Description - text-sm */}
      <Skeleton className="h-4 w-32" />
      {/* Humidity & Wind stats row */}
      <div className="flex gap-4">
        <Skeleton className="h-3 w-14" /> {/* H: XX% */}
        <Skeleton className="h-3 w-16" /> {/* W: XX mph */}
      </div>
    </div>
  );
};
