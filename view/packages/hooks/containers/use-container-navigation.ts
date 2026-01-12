import { useRouter } from 'next/navigation';
import { Container } from '@/redux/services/container/containerApi';

export function useContainerNavigation() {
  const router = useRouter();

  const navigateToContainer = (containerId: string) => {
    router.push(`/containers/${containerId}`);
  };

  const handleRowClick = (container: Container) => {
    navigateToContainer(container.id);
  };

  return {
    navigateToContainer,
    handleRowClick
  };
}
