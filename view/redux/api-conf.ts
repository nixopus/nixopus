export enum AUTHURLS {
  USER_LOGIN = 'v1/auth/login',
  USER_REGISTER = 'v1/auth/register',
  USER_DETAILS = 'v1/user',
  REFRESH_TOKEN = 'v1/auth/refresh-token',
  REQUEST_PASSWORD_RESET = 'v1/auth/request-password-reset',
  RESET_PASSWORD = 'v1/auth/reset-password',
  SEND_VERIFICATION = 'v1/auth/send-verification-email',
  VERIFY_EMAIL = 'v1/auth/verify-email',
  LOGOUT = 'v1/auth/logout',
  SETUP_TWO_FACTOR = 'v1/auth/setup-2fa',
  VERIFY_TWO_FACTOR = 'v1/auth/verify-2fa',
  DISABLE_TWO_FACTOR = 'v1/auth/disable-2fa',
  TWO_FACTOR_LOGIN = 'v1/auth/2fa-login',
  IS_ADMIN_REGISTERED = 'v1/auth/is-admin-registered'
}

export enum USERURLS {
  USER_ORGANIZATIONS = 'v1/user/organizations',
  CREATE_ORGANIZATION = 'v1/organizations',
  ADD_USER_TO_ORGANIZATION = 'v1/organizations/add-user',
  REMOVE_USER_FROM_ORGANIZATION = 'v1/organizations/remove-user',
  UPDATE_USER_ROLE = 'v1/organizations/update-user-role',
  UPDATE_USER_NAME = 'v1/user/name',
  ORGANIZATION_USERS = 'v1/organizations/users',
  CREATE_USER = 'v1/auth/create-user',
  GET_RESOURCES = 'v1/organizations/resources',
  GET_SETTINGS = 'v1/user/settings',
  UPDATE_FONT = 'v1/user/settings/font',
  UPDATE_THEME = 'v1/user/settings/theme',
  UPDATE_LANGUAGE = 'v1/user/settings/language',
  UPDATE_AUTO_UPDATE = 'v1/user/settings/auto-update',
  CHECK_FOR_UPDATES = 'v1/update/check',
  PERFORM_UPDATE = 'v1/update',
  UPDATE_AVATAR = 'v1/user/avatar',
  SEND_INVITE = 'v1/organizations/invite/send',
  RESEND_INVITE = 'v1/organizations/invite/resend',
  GET_PREFERENCES = 'v1/user/preferences',
  UPDATE_PREFERENCES = 'v1/user/preferences',
  GET_ORGANIZATION_SETTINGS = 'v1/organizations/settings',
  UPDATE_ORGANIZATION_SETTINGS = 'v1/organizations/settings'
}

export enum USER_NOTIFICATION_SETTINGS {
  ADD_SMTP = 'v1/notification/smtp',
  GET_SMTP = 'v1/notification/smtp',
  UPDATE_SMTP = 'v1/notification/smtp',
  DELETE_SMTP = 'v1/notification/smtp',

  GET_PREFERENCES = 'v1/notification/preferences',
  UPDATE_PREFERENCES = 'v1/notification/preferences',

  GET_WEBHOOK = 'v1/notification/webhook',
  CREATE_WEBHOOK = 'v1/notification/webhook',
  UPDATE_WEBHOOK = 'v1/notification/webhook',
  DELETE_WEBHOOK = 'v1/notification/webhook'
}

export enum DOMAIN_SETTINGS {
  ADD_DOMAIN = 'v1/domain',
  GET_DOMAINS = 'v1/domains',
  UPDATE_DOMAIN = 'v1/domain',
  DELETE_DOMAIN = 'v1/domain',
  GENERATE_RANDOM_SUBDOMAIN = 'v1/domain/generate'
}

export enum GITHUB_CONNECTOR {
  ADD_GITHUB_CONNECTOR = 'v1/github-connector',
  GET_GITHUB_CONNECTORS = 'v1/github-connector/all',
  UPDATE_GITHUB_CONNECTOR = 'v1/github-connector',
  DELETE_GITHUB_CONNECTOR = 'v1/github-connector',
  ALL_REPOSITORIES = 'v1/github-connector/repositories',
  GET_REPOSITORY_BRANCHES = 'v1/github-connector/repository/branches'
}

export enum DEPLOY {
  GET_APPLICATIONS = 'v1/deploy/applications',
  CREATE_APPLICATION = 'v1/deploy/application',
  CREATE_PROJECT = 'v1/deploy/application/project',
  DEPLOY_PROJECT = 'v1/deploy/application/project/deploy',
  DUPLICATE_PROJECT = 'v1/deploy/application/project/duplicate',
  GET_PROJECT_FAMILY = 'v1/deploy/application/project/family',
  GET_FAMILY_ENVIRONMENTS = 'v1/deploy/application/project/family/environments',
  GET_APPLICATION = 'v1/deploy/application',
  UPDATE_APPLICATION = 'v1/deploy/application',
  REDEPLOY_APPLICATION = 'v1/deploy/application/redeploy',
  DEPLOYMENT = 'v1/deploy/application/deployments',
  DELETE_APPLICATION = 'v1/deploy/application',
  ROLLBACK_APPLICATION = 'v1/deploy/application/rollback',
  RESTART_APPLICATION = 'v1/deploy/application/restart',
  GET_APPLICATION_LOGS = 'v1/deploy/application/logs/{application_id}',
  GET_DEPLOYMENT_LOGS = 'v1/deploy/application/deployments/{deployment_id}/logs',
  GET_APPLICATION_DEPLOYMENTS = 'v1/deploy/application/deployments',
  UPDATE_APPLICATION_LABELS = 'v1/deploy/application/labels'
}

export enum SOCKET_EVENTS {
  MONITOR_APPLICATION_DEPLOYMENT = 'monitor_application_deployment',
  MONITOR_HEALTH_CHECK = 'monitor_health_check'
}

export enum SOCKET_ACTIONS {
  SUBSCRIBE = 'subscribe'
}

export enum FILEMANAGERURLS {
  LIST_FILES_AT_PATH = 'v1/file-manager',
  CREATE_DIRECTORY = 'v1/file-manager/create-directory',
  DELETE_DIRECTORY = 'v1/file-manager/delete-directory',
  MOVE_FOLDER_FILES_RECURSIVELY_OR_RENAME = 'v1/file-manager/move-directory',
  COPY_FOLDER_FILES_RECURSIVELY = 'v1/file-manager/copy-directory',
  CALCULATE_DIRECTORY_SIZE = 'v1/file-manager/directories/size',
  UPLOAD_FILE = 'v1/file-manager/upload'
}

export enum AUDITURLS {
  GET_RECENT_AUDIT_LOGS = 'v1/audit/logs'
}

export const FEATURE_FLAGS = {
  GET_FEATURE_FLAGS: 'v1/feature-flags',
  UPDATE_FEATURE_FLAG: 'v1/feature-flags',
  CHECK_FEATURE_ENABLED: 'v1/feature-flags/check'
};

export enum CONTAINERURLS {
  GET_CONTAINERS = 'v1/container',
  GET_CONTAINER = 'v1/container/{container_id}',
  START_CONTAINER = 'v1/container/{container_id}/start',
  STOP_CONTAINER = 'v1/container/{container_id}/stop',
  REMOVE_CONTAINER = 'v1/container/{container_id}',
  GET_CONTAINER_LOGS = 'v1/container/{container_id}/logs',
  UPDATE_CONTAINER_RESOURCES = 'v1/container/{container_id}/resources'
}

export enum IMAGEURLS {
  GET_IMAGES = 'v1/container/images',
  PRUNE_IMAGES = 'v1/container/prune/images',
  PRUNE_BUILD_CACHE = 'v1/container/prune/build-cache'
}

export enum EXTENSIONURLS {
  GET_EXTENSIONS = 'v1/extensions',
  GET_CATEGORIES = 'v1/extensions/categories',
  GET_EXTENSION = 'v1/extensions/{id}',
  GET_EXTENSION_BY_ID = 'v1/extensions/by-extension-id/{extension_id}',
  FORK_EXTENSION = 'v1/extensions/{extension_id}/fork',
  DELETE_EXTENSION = 'v1/extensions/{id}',
  RUN_EXTENSION = 'v1/extensions/{extension_id}/run',
  CANCEL_EXECUTION = 'v1/extensions/execution/{execution_id}/cancel',
  GET_EXECUTION = 'v1/extensions/execution/{execution_id}',
  LIST_EXECUTIONS = 'v1/extensions/by-extension-id/{extension_id}/executions',
  GET_EXECUTION_LOGS = 'v1/extensions/execution/{execution_id}/logs'
}

export enum HEALTHCHECKURLS {
  GET_HEALTH_CHECK = 'v1/healthcheck',
  CREATE_HEALTH_CHECK = 'v1/healthcheck',
  UPDATE_HEALTH_CHECK = 'v1/healthcheck',
  DELETE_HEALTH_CHECK = 'v1/healthcheck',
  TOGGLE_HEALTH_CHECK = 'v1/healthcheck/toggle',
  GET_HEALTH_CHECK_RESULTS = 'v1/healthcheck/results',
  GET_HEALTH_CHECK_STATS = 'v1/healthcheck/stats'
}
