export interface GitHubAppManifest {
  name: string;
  url: string;
  hook_attributes: {
    url: string;
    active: boolean;
  };
  redirect_url: string;
  callback_urls: string[];
  public: boolean;
  default_permissions: {
    [key: string]: 'read' | 'write' | 'admin';
  };
  default_events: string[];
  setup_url?: string;
  setup_on_update?: boolean;
  request_oauth_on_install?: boolean;
}

export interface GitHubAppCredentials {
  id: number;
  slug: string;
  pem: string;
  client_id: string;
  client_secret: string;
  webhook_secret: string;
}

export interface GitHubAppProps {
  organization?: string;
  webhookUrl?: string;
  appUrl?: string;
  redirectUrl?: string;
  onSuccess?: (credentials: GitHubAppCredentials) => void;
  onError?: (error: Error) => void;
}

export type GitHubAppStatus = 'initial' | 'redirecting' | 'converting' | 'success' | 'error';

export interface GithubConnector {
  id: string;
  name: string;
  slug: string;
  pem: string;
  client_id: string;
  client_secret: string;
  webhook_secret: string;
  installation_id: string;
}

export interface CreateGithubConnectorRequest {
  app_id: string;
  slug: string;
  pem: string;
  client_id: string;
  client_secret: string;
  webhook_secret: string;
}

export interface UpdateGithubConnectorRequest {
  installation_id: string;
}
