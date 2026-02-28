export const SUGGESTED_ENVIRONMENTS = ['production', 'staging', 'development'] as const;

export enum BuildPack {
  Dockerfile = 'dockerfile',
  DockerCompose = 'docker-compose',
  Static = 'static'
}
