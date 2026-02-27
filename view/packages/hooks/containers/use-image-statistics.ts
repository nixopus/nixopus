import { useMemo } from 'react';
import { ContainerImage } from '@/packages/types/containers';

export function useImageStatistics(images: ContainerImage[]) {
  const statistics = useMemo(() => {
    const totalSize = images.reduce((acc, img) => acc + img.size, 0);
    const totalLayers = images.length;
    return { totalSize, totalLayers };
  }, [images]);

  return statistics;
}
