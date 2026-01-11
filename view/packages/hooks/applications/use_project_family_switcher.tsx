import { useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Application } from '@/redux/types/applications';
import { useGetProjectFamilyQuery } from '@/redux/services/deploy/applicationsApi';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ChevronsUpDown, Check } from 'lucide-react';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';

interface UseProjectFamilySwitcherProps {
  application: Application;
}

const getEnvironmentStyles = (environment: string) => {
  switch (environment) {
    case 'production':
      return 'border-emerald-500/30 text-emerald-500 bg-emerald-500/10';
    case 'staging':
      return 'border-amber-500/30 text-amber-500 bg-amber-500/10';
    case 'development':
      return 'border-blue-500/30 text-blue-500 bg-blue-500/10';
    default:
      return 'border-zinc-500/30 text-zinc-500 bg-zinc-500/10';
  }
};

export function useProjectFamilySwitcher({ application }: UseProjectFamilySwitcherProps) {
  const { t } = useTranslation();
  const router = useRouter();

  const { data: familyProjects, isLoading } = useGetProjectFamilyQuery(
    { familyId: application.family_id || '' },
    { skip: !application.family_id }
  );

  const handleSelectProject = (projectId: string) => {
    if (projectId !== application.id) {
      router.push(`/self-host/application/${projectId}`);
    }
  };

  const shouldShow = useMemo(() => {
    return application.family_id && !isLoading && familyProjects && familyProjects.length > 1;
  }, [application.family_id, isLoading, familyProjects]);

  const dropdownContent = useMemo(() => {
    if (!familyProjects) return null;

    return (
      <>
        <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
          {t('selfHost.applicationDetails.header.familySwitcher.title')}
        </div>
        {familyProjects.map((project) => (
          <DropdownMenuItem
            key={project.id}
            onClick={() => handleSelectProject(project.id)}
            className={cn(
              'flex items-center justify-between gap-2 cursor-pointer',
              project.id === application.id && 'bg-accent'
            )}
          >
            <div className="flex items-center gap-2 min-w-0">
              <span className="truncate font-medium">{project.name}</span>
              <Badge
                variant="outline"
                className={cn(
                  'text-[10px] px-1.5 py-0 capitalize shrink-0',
                  getEnvironmentStyles(project.environment)
                )}
              >
                {project.environment}
              </Badge>
            </div>
            {project.id === application.id && <Check className="h-4 w-4 shrink-0 text-primary" />}
          </DropdownMenuItem>
        ))}
      </>
    );
  }, [familyProjects, application.id, t, handleSelectProject]);

  const trigger = useMemo(
    () => (
      <Button
        variant="ghost"
        size="icon"
        className="h-8 w-8 text-muted-foreground hover:text-foreground"
        aria-label={t('selfHost.applicationDetails.header.familySwitcher.label')}
      >
        <ChevronsUpDown className="h-4 w-4" />
      </Button>
    ),
    [t]
  );

  return {
    shouldShow,
    isLoading,
    trigger,
    dropdownContent
  };
}
