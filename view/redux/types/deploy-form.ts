export enum Environment {
  Production = 'production',
  Staging = 'staging',
  Development = 'development'
}

export enum BuildPack {
  Dockerfile = 'dockerfile',
  DockerCompose = 'docker-compose',
  Static = 'static'
}

export enum ComposeSourceType {
  Repository = 'repository',
  URL = 'url',
  Raw = 'raw'
}
