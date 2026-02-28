import { BuildPack } from '@/redux/types/deploy-form';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

interface UseDeploymentConfigurationProps {
  branch?: string;
  domains?: string[];
  build_pack?: BuildPack;
  env_variables?: Record<string, string>;
  build_variables?: Record<string, string>;
  domainsEditable?: boolean;
}

export function useDeploymentConfiguration({
  branch = '',
  domains = [],
  build_pack = BuildPack.Dockerfile,
  env_variables = {},
  build_variables = {},
  domainsEditable = false
}: UseDeploymentConfigurationProps) {
  const { t } = useTranslation();

  const isDockerCompose = build_pack === BuildPack.DockerCompose;

  const dockerConfigFields = [
    {
      label: t('selfHost.configuration.fields.basePath.label'),
      name: 'base_path',
      placeholder: '/',
      descriptionText: t('selfHost.configuration.fields.basePath.description')
    },
    {
      label: isDockerCompose
        ? t('selfHost.configuration.fields.dockerComposePath.label')
        : t('selfHost.configuration.fields.dockerfilePath.label'),
      name: 'DockerfilePath',
      placeholder: isDockerCompose ? 'docker-compose.yml' : 'Dockerfile',
      descriptionText: isDockerCompose
        ? t('selfHost.configuration.fields.dockerComposePath.description')
        : t('selfHost.configuration.fields.dockerfilePath.description')
    }
  ];

  const envVariableEditors = [
    {
      label: t('selfHost.configuration.fields.environmentVariables.label'),
      name: 'environment_variables',
      defaultValues: env_variables
    },
    {
      label: t('selfHost.configuration.fields.buildVariables.label'),
      name: 'build_variables',
      defaultValues: build_variables
    }
  ];

  const commandFields = [
    {
      label: t('selfHost.configuration.fields.preRunCommands.label'),
      name: 'pre_run_command',
      placeholder: t('selfHost.configuration.fields.preRunCommands.placeholder')
    },
    {
      label: t('selfHost.configuration.fields.postRunCommands.label'),
      name: 'post_run_command',
      placeholder: t('selfHost.configuration.fields.postRunCommands.placeholder')
    }
  ];

  const readOnlyFields = [
    {
      label: t('selfHost.configuration.fields.branch.label'),
      value: branch,
      description: t('selfHost.configuration.fields.branch.description')
    },
    ...(domainsEditable
      ? []
      : [
          {
            label: t('selfHost.configuration.fields.domain.label'),
            value: domains && domains.length > 0 ? domains.join(', ') : '',
            description: t('selfHost.configuration.fields.domain.description')
          }
        ]),
    {
      label: t('selfHost.configuration.fields.buildPack.label'),
      value: build_pack,
      description: t('selfHost.configuration.fields.buildPack.description')
    }
  ];

  return {
    dockerConfigFields,
    envVariableEditors,
    commandFields,
    readOnlyFields
  };
}
