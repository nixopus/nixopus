export const appInfo = {
  // learn more about this on https://supertokens.com/docs/thirdpartyemailpassword/appinfo
  appName: 'Nixopus',
  apiDomain: process.env.API_URL?.replace('/api', '') || 'http://localhost:8080',
  websiteDomain: process.env.WEBSITE_DOMAIN || 'http://localhost:3000',
  apiBasePath: '/auth',
  websiteBasePath: '/auth'
};
