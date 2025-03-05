'use client';
import { useAppSelector } from '@/redux/hooks';
import { GithubRepository } from '@/redux/types/github';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

/**
 * @description
 * This hook is used to manage the state of the repository when the user navigates to /dashboard/create/:id
 * It takes the repository id from the url and finds the corresponding repository in the redux store
 * If the repository is not found, it redirects to /dashboard
 * @returns an object with the selected repository
 */
function useFindRepository() {
  const repositories = useAppSelector((state) => state.githubConnector.repositories);
  const [repository, setRepository] = useState<GithubRepository | null>(null);
  const router = useRouter();
  const pathname = usePathname();
  const repositoryId = pathname?.split('/').pop();

  // this will trigger when user refreshes the page, or there is no state of repositories in redux store
  useEffect(() => {
    // if there are no repositories, simply redirect to dashboard
    if (repositories.length <= 0) {
      router.push(`/dashboard`);
      return;
    }
    // find the selected repository by matching the id in the url
    const selectedRepository = repositories.find(
      (repo: GithubRepository) => repo.id.toString() === repositoryId
    );

    if (!selectedRepository) {
      return;
    }

    setRepository(selectedRepository);
  }, [repositories]);

  return {
    repository
  };
}

export default useFindRepository;
