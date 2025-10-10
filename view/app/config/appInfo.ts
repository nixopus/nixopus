export const appInfo = {
  // learn more about this on https://supertokens.com/docs/thirdpartyemailpassword/appinfo
  appName: 'Nixopus',
  apiDomain: process.env.NEXT_PUBLIC_API_URL?.replace('/api', ''),
  websiteDomain: process.env.NEXT_PUBLIC_WEBSITE_DOMAIN || 'http://localhost:3000',
  apiBasePath: '/auth',
  websiteBasePath: '/auth'
};
