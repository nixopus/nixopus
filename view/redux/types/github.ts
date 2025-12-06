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
  onCreateClick?: (createFn: () => void) => void;
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
  connector_id?: string; // Optional: if provided, update this specific connector
}

export interface GithubRepository {
  id: number;
  node_id: string;
  name: string;
  full_name: string;
  private: boolean;
  owner: Owner;
  html_url: string;
  description: string | null;
  fork: boolean;
  url: string;
  forks_url: string;
  keys_url: string;
  collaborators_url: string;
  teams_url: string;
  hooks_url: string;
  issue_events_url: string;
  events_url: string;
  assignees_url: string;
  branches_url: string;
  tags_url: string;
  blobs_url: string;
  git_tags_url: string;
  git_refs_url: string;
  trees_url: string;
  statuses_url: string;
  languages_url: string;
  stargazers_url: string;
  contributors_url: string;
  subscribers_url: string;
  subscription_url: string;
  commits_url: string;
  git_commits_url: string;
  comments_url: string;
  issue_comment_url: string;
  contents_url: string;
  compare_url: string;
  merges_url: string;
  archive_url: string;
  downloads_url: string;
  issues_url: string;
  pulls_url: string;
  milestones_url: string;
  notifications_url: string;
  labels_url: string;
  releases_url: string;
  deployments_url: string;
  created_at: string;
  updated_at: string;
  pushed_at: string;
  git_url: string;
  ssh_url: string;
  clone_url: string;
  svn_url: string;
  homepage: string | null;
  size: number;
  stargazers_count: number;
  watchers_count: number;
  language: string | null;
  has_issues: boolean;
  has_projects: boolean;
  has_downloads: boolean;
  has_wiki: boolean;
  has_pages: boolean;
  has_discussions: boolean;
  forks_count: number;
  mirror_url: string | null;
  archived: boolean;
  disabled: boolean;
  open_issues_count: number;
  license: License | null;
  allow_forking: boolean;
  is_template: boolean;
  web_commit_signoff_required: boolean;
  topics: string[];
  visibility: string;
  forks: number;
  open_issues: number;
  watchers: number;
  default_branch: string;
  permissions?: Permissions;
}

interface Owner {
  login: string;
  id: number;
  node_id: string;
  avatar_url: string;
  gravatar_id: string;
  url: string;
  html_url: string;
  followers_url: string;
  following_url: string;
  gists_url: string;
  starred_url: string;
  subscriptions_url: string;
  organizations_url: string;
  repos_url: string;
  events_url: string;
  received_events_url: string;
  type: string;
  site_admin: boolean;
}

interface License {
  key: string;
  name: string;
  spdx_id: string;
  url: string | null;
  node_id: string;
}

interface Permissions {
  admin: boolean;
  maintain: boolean;
  push: boolean;
  triage: boolean;
  pull: boolean;
}

export interface GithubRepositoryBranch {
  name: string;
  commit: {
    sha: string;
    url: string;
  };
  protected: boolean;
}
