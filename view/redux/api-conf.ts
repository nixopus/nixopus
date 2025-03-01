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
  ORGANIZATION_USERS = 'v1/organizations/users'
}

export enum USER_NOTIFICATION_SETTINGS {
  ADD_SMTP = 'v1/notification/smtp',
  GET_SMTP = 'v1/notification/smtp',
  UPDATE_SMTP = 'v1/notification/smtp',
  DELETE_SMTP = 'v1/notification/smtp'
}
