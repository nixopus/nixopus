import { User } from './user';
import { HealthCheck } from './healthcheck';

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
  domain?: string;
  user?: User;
  status?: ApplicationStatus;
  logs?: ApplicationLogs[];
  deployments?: ApplicationDeployment[];
  dockerfile_path?: string;
  base_path?: string;
  labels?: string[];
  family_id?: string;
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
};

export type ApplicationLogsResponse = {
  logs: ApplicationLogs[];
  total_count: number;
  current_page: number;
  total_pages: number;
};

export type Status =
  | 'draft'
  | 'failed'
  | 'cloning'
  | 'building'
  | 'deploying'
  | 'deployed'
  | 'started'
  | 'running'
  | 'stopped';

export type Environment = 'development' | 'staging' | 'production';

export type BuildPack = 'dockerfile' | 'dockerCompose' | 'static';

export interface CreateApplicationRequest {
  name: string;
  environment: Environment;
  branch: string;
  port: number;
  domain: string;
  repository: string;
  build_pack: BuildPack;
  environment_variables: Record<string, string>;
  build_variables: Record<string, string>;
  pre_run_command: string;
  post_run_command: string;
  dockerfile_path: string;
  base_path: string;
}

export interface UpdateDeploymentRequest {
  name?: string;
  environment?: Environment;
  pre_run_command?: string;
  post_run_command?: string;
  build_variables?: Record<string, string>;
  environment_variables?: Record<string, string>;
  port?: number;
  id?: string;
  force?: boolean;
  dockerfile_path?: string;
  base_path?: string;
}

export interface ReDeployApplicationRequest {
  id: string;
  force: boolean;
  force_without_cache: boolean;
}

// CreateProjectRequest is used to create a project without triggering deployment.
// Only name, domain, and repository are required. Other fields have defaults.
export interface CreateProjectRequest {
  name: string;
  domain: string;
  repository: string;
  environment?: Environment;
  build_pack?: BuildPack;
  branch?: string;
  pre_run_command?: string;
  post_run_command?: string;
  build_variables?: Record<string, string>;
  environment_variables?: Record<string, string>;
  port?: number;
  dockerfile_path?: string;
  base_path?: string;
}

// DeployProjectRequest is used to trigger deployment of an existing project.
export interface DeployProjectRequest {
  id: string;
}

// DuplicateProjectRequest is used to create a duplicate of an existing project with a different environment.
export interface DuplicateProjectRequest {
  source_project_id: string;
  domain: string;
  environment: Environment;
  branch?: string;
}

// ProjectFamilyResponse contains the projects in a family.
export interface ProjectFamilyResponse {
  projects: Application[];
}

// Health Check Component Props
export interface HealthCheckChartProps {
  applicationId: string;
  setDialogOpen: (open: boolean) => void;
  dialogOpen: boolean;
}

export interface HealthCheckCardProps {
  application: Application;
}

export interface HealthCheckDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  application: Application;
  healthCheck?: HealthCheck;
}
