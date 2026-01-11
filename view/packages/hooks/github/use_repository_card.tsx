import { useMemo } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { GithubRepository } from '@/redux/types/github';
import { Lock, Unlock, Star, GitFork, AlertCircle } from 'lucide-react';

interface UseRepositoryCardProps extends GithubRepository {
  setSelectedRepository: (repo: string) => void;
}

const skeletonCount = 6;

export function useRepositoryCard({
  name,
  html_url: url,
  description,
  private: isPrivate,
  stargazers_count,
  forks_count,
  open_issues_count,
  license,
  topics,
  id,
  setSelectedRepository
}: UseRepositoryCardProps) {
  const { t } = useTranslation();

  const stats = useMemo(
    () => [
      {
        icon: Star,
        value: stargazers_count?.toLocaleString() || '0',
        key: 'stars'
      },
      {
        icon: GitFork,
        value: forks_count?.toLocaleString() || '0',
        key: 'forks'
      },
      {
        icon: AlertCircle,
        value: open_issues_count?.toLocaleString() || '0',
        key: 'issues'
      }
    ],
    [stargazers_count, forks_count, open_issues_count]
  );

  const visibilityBadge = useMemo(
    () => ({
      variant: (isPrivate ? 'secondary' : 'outline') as 'secondary' | 'outline',
      icon: isPrivate ? Lock : Unlock,
      label: isPrivate
        ? t('selfHost.repositoryCard.visibility.private')
        : t('selfHost.repositoryCard.visibility.public')
    }),
    [isPrivate, t]
  );

  const licenseBadge = useMemo(
    () => (license && license.spdx_id ? { label: license.spdx_id } : null),
    [license]
  );

  const displayTopics = useMemo(() => {
    if (!topics || topics.length === 0) return null;
    const visibleTopics = topics.slice(0, 2);
    const remainingCount = topics.length - 2;
    return {
      visible: visibleTopics,
      remainingCount: remainingCount > 0 ? remainingCount : 0
    };
  }, [topics]);

  const handleClick = () => {
    setSelectedRepository(id.toString());
  };

  const displayName = name || t('selfHost.repositoryCard.unnamed');

  const skeletonItems = useMemo(() => Array.from({ length: skeletonCount }, (_, i) => i), []);

  return {
    displayName,
    url,
    description,
    stats,
    visibilityBadge,
    licenseBadge,
    displayTopics,
    handleClick,
    skeletonItems
  };
}
