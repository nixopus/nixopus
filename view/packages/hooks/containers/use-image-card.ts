import { useState } from 'react';
import { ContainerImage } from '@/packages/types/containers';
import { formatImageId } from '@/packages/utils/container-helpers';

export function useImageCard(image: ContainerImage, isFirst: boolean) {
  const [expanded, setExpanded] = useState(isFirst);
  const createdDate = new Date(image.created * 1000);
  const primaryTag = image.repo_tags?.[0] || '<none>';
  const hasLabels = image.labels && Object.keys(image.labels).length > 0;
  const imageId = formatImageId(image.id);

  const toggleExpanded = () => setExpanded(!expanded);

  return {
    expanded,
    toggleExpanded,
    createdDate,
    primaryTag,
    hasLabels,
    imageId
  };
}
