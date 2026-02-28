export const SUGGESTED_ENVIRONMENTS = ['production', 'staging', 'development'] as const;

export enum BuildPack {
  Dockerfile = 'dockerfile',
  DockerCompose = 'dockerCompose',
  Static = 'static'
}
