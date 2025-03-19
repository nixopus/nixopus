import { Domain } from './domain';
import { User } from './user';

export type Application = {
  id: string;
  name: string;
  port: number;
  environment: Environment;
  build_variables: string;
  environment_variables: string;
  build_pack: BuildPack;
  repository: string;
  branch: string;
  pre_run_command: string;
  post_run_command: string;
  domain_id: string;
  user_id: string;
  created_at: Date;
  updated_at: Date;
  domain?: Domain;
  user?: User;
  status?: ApplicationStatus;
  logs?: ApplicationLogs[];
  deployments?: ApplicationDeployment[];
};

export type ApplicationStatus = {
  id: string;
  application_id: string;
  status: Status;
  created_at: Date;
  updated_at: Date;
  application?: Application;
};

export type ApplicationDeployment = {
  id: string;
  application_id: string;
  created_at: string;
  updated_at: string;

  application?: Application;
  status?: ApplicationDeploymentStatus;
  logs?: ApplicationLogs[];
  commit_hash?: string;
  container_id?: string;
  container_name?: string;
  container_image?: string;
  container_status?: string;
};

export type ApplicationDeploymentStatus = {
  id: string;
  application_deployment_id: string;
  status: Status;
  created_at: string;
  updated_at: string;

  application_deployment?: ApplicationDeployment;
};

export type ApplicationLogs = {
  id: string;
  application_id: string;
  created_at: string;
  updated_at: string;
  log: string;
  application_deployment_id: string;
  application_deployment?: ApplicationDeployment;
  application?: Application;
};

export type Status = 'failed' | 'cloning' | 'building' | 'deploying' | 'deployed';

export type Environment = 'development' | 'staging' | 'production';

export type BuildPack = 'dockerfile' | 'dockerCompose' | 'static';

export interface CreateApplicationRequest {
  name: string;
  environment: Environment;
  branch: string;
  port: number;
  domain_id: string;
  repository: string;
  build_pack: BuildPack;
  env_variables: Record<string, string>;
  build_variables: Record<string, string>;
  pre_run_commands: string;
  post_run_commands: string;
}

export interface UpdateDeploymentRequest {
  name?: string;
  pre_run_command?: string;
  post_run_command?: string;
  build_variables?: Record<string, string>;
  environment_variables?: Record<string, string>;
  port?: number;
  id?: string;
  force?: boolean;
}

export interface ReDeployApplicationRequest {
  id: string;
  force: boolean;
  force_without_cache: boolean;
}
