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
};

export type ApplicationStatus = {
  id: string;
  application_id: string;
  status: Status;
  created_at: Date;
  updated_at: Date;
  application?: Application;
};

export type ApplicationLogs = {
  id: string;
  application_id: string;
  created_at: Date;
  updated_at: Date;
  log: string;
  application?: Application;
};

export type Status = 'started' | 'running' | 'stopped' | 'failed';

export type Environment = 'development' | 'staging' | 'production';

export type BuildPack = 'dockerfile' | 'dockerCompose' | 'static';
