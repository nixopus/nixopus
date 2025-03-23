export enum AUTHURLS {
  USER_LOGIN = 'v1/auth/login',
  USER_DETAILS = 'v1/user',
  REFRESH_TOKEN = 'v1/auth/refresh-token'
}

export enum USERURLS {
  USER_ORGANIZATIONS = 'v1/user/organizations',
  CREATE_ORGANIZATION = 'v1/organizations',
  ADD_USER_TO_ORGANIZATION = 'v1/organizations/user',
  UPDATE_USER_NAME = 'v1/user/name',
  REQUEST_PASSWORD_RESET_LINK = 'v1/auth/request-password-reset',
  ORGANIZATION_USERS = 'v1/organizations/users',
  CREATE_USER = 'v1/auth/create-user',
}

export enum USER_NOTIFICATION_SETTINGS {
  ADD_SMTP = 'v1/notification/smtp',
  GET_SMTP = 'v1/notification/smtp',
  UPDATE_SMTP = 'v1/notification/smtp',
  DELETE_SMTP = 'v1/notification/smtp',

  GET_PREFERENCES = 'v1/notification/preferences',
  UPDATE_PREFERENCES = 'v1/notification/preferences'
}

export enum DOMAIN_SETTINGS {
  ADD_DOMAIN = 'v1/domain',
  GET_DOMAINS = 'v1/domain/all',
  UPDATE_DOMAIN = 'v1/domain',
  DELETE_DOMAIN = 'v1/domain',
  GENERATE_RANDOM_SUBDOMAIN= "v1/domain/generate"
}

export enum GITHUB_CONNECTOR {
  ADD_GITHUB_CONNECTOR = 'v1/github-connector',
  GET_GITHUB_CONNECTORS = 'v1/github-connector/all',
  UPDATE_GITHUB_CONNECTOR = 'v1/github-connector',
  DELETE_GITHUB_CONNECTOR = 'v1/github-connector',
  ALL_REPOSITORIES = 'v1/github-connector/repositories'
}

export enum DEPLOY {
  GET_APPLICATIONS = 'v1/deploy/applications',
  CREATE_APPLICATION = 'v1/deploy/application',
  GET_APPLICATION = 'v1/deploy/application',
  UPDATE_APPLICATION = 'v1/deploy/application',
  REDEPLOY_APPLICATION = 'v1/deploy/application/redeploy',
  DEPLOYMENT = 'v1/deploy/application/deployments',
  DELETE_APPLICATION = 'v1/deploy/application',
  ROLLBACK_APPLICATION = 'v1/deploy/application/rollback',
  RESTART_APPLICATION = 'v1/deploy/application/restart'
}

export enum SOCKET_EVENTS {
  MONITOR_APPLICATION_DEPLOYMENT = 'monitor_application_deployment'
}

export enum SOCKET_ACTIONS {
  SUBSCRIBE = 'subscribe'
}
