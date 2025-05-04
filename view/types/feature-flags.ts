export interface FeatureFlag {
  id: string;
  organization_id: string;
  feature_name: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface UpdateFeatureFlagRequest {
  feature_name: string;
  is_enabled: boolean;
}

export interface GetFeatureFlagsResponse extends Array<FeatureFlag> {}

export type FeatureName = (typeof FeatureNames)[keyof typeof FeatureNames];

export enum FeatureNames {
  FeatureTerminal = 'terminal',
  FeatureFileManager = 'file_manager',
  FeatureMonitoring = 'monitoring',
  FeatureProxyConfig = 'proxy_config',
  FeatureGithubConnector = 'github_connector',
  FeatureAudit = 'audit',
  FeatureNotifications = 'notifications',
  FeatureDomain = 'domain',
  FeatureSelfHosted = 'self_hosted',
  FeatureContainer = 'container'
}

export const featureGroups: Record<string, FeatureName[]> = {
  infrastructure: [
    FeatureNames.FeatureDomain,
    FeatureNames.FeatureTerminal,
    FeatureNames.FeatureFileManager,
    FeatureNames.FeatureProxyConfig
  ],
  development: [
    FeatureNames.FeatureGithubConnector,
    FeatureNames.FeatureSelfHosted,
    FeatureNames.FeatureContainer
  ],
  monitoring: [FeatureNames.FeatureMonitoring, FeatureNames.FeatureAudit],
  notifications: [FeatureNames.FeatureNotifications]
};
